//go:build !testing_aws

package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gookit/goutil/dump"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"
)

const kubeconfigTemplate = `apiVersion: v1
clusters:
- cluster:
    server: {{.ClusterEndpoint}}
    certificate-authority-data: {{.CertificateAuthorityData}}
  name: {{.ClusterName}}
contexts:
- context:
    cluster: {{.ClusterName}}
    user: {{.ClusterName}}
  name: {{.ClusterName}}
current-context: {{.ClusterName}}
kind: Config
preferences: {}
users:
- name: {{.ClusterName}}
  user:
    token: {{.Token}}
`

const (
	initialNicMinDelay             = time.Second * 1
	initialNicMaxDelay             = time.Second * 5
	initialSubnetMinDelay          = time.Second * 1
	initialSubnetMaxDelay          = time.Second * 5
	initialInstanceMinDelay        = time.Second * 5
	initialInstanceMaxDelay        = time.Second * 10
	initialNicWaiterTime           = time.Second * 10
	initialSubnetWaiterTime        = time.Second * 10
	instanceInitialWaiterTime      = time.Minute * 10
	initialNicDeletionWaiterTime   = time.Second * 30
	instanceInitialTerminationTime = time.Second * 200
	managedClusterActiveWaiter     = time.Minute * 15
	managedNodeGroupActiveWaiter   = time.Minute * 10
)

func ProvideClient() AwsGo {
	return &AwsClient{}
}

type AwsClient struct {
	region    string
	vpc       string
	stsClient *sts.Client
	ec2Client *ec2.Client
	eksClient *eks.Client
	config    *aws.Config
	iam       *iam.Client
	storage   ksctlTypes.StorageFactory
}

func newEC2Client(region string) (aws.Config, error) {
	_session, err := config.LoadDefaultConfig(awsCtx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			),
		),
	)
	if err != nil {
		return _session, ksctlErrors.ErrInternal.Wrap(
			log.NewError(awsCtx, "Failed Init aws session", "Reason", err),
		)
	}
	log.Success(awsCtx, "AWS Session created successfully")
	return _session, nil
}

func (client *AwsClient) AuthorizeSecurityGroupIngress(ctx context.Context,
	parameter ec2.AuthorizeSecurityGroupIngressInput) error {

	_, err := client.ec2Client.AuthorizeSecurityGroupIngress(ctx, &parameter)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Authorizing Security Group Ingress", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error) {
	resp, err := client.ec2Client.DescribeImages(awsCtx, filter)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(
			log.NewError(awsCtx, "failed to describe images", "Reason", err),
		)
	}
	if len(resp.Images) == 0 {
		return "", ksctlErrors.ErrInternal.Wrap(
			log.NewError(awsCtx, "no images found"),
		)
	}

	var savedImages []types.Image

	for _, i := range resp.Images {
		if trustedSource(*i.OwnerId) && *i.Public {
			savedImages = append(savedImages, i)
		}
	}

	sort.Slice(savedImages, func(i, j int) bool {
		return *savedImages[i].CreationDate > *savedImages[j].CreationDate
	})

	for x := 0; x < 2; x++ {
		i := savedImages[x]
		if i.ImageOwnerAlias != nil {
			log.Debug(awsCtx, "ownerAlias", *i.ImageOwnerAlias)
		}
		log.Debug(awsCtx, "Printing amis", "creationdate", *i.CreationDate, "public", *i.Public, "ownerid", *i.OwnerId, "architecture", i.Architecture.Values(), "name", *i.Name, "imageid", *i.ImageId)
	}

	selectedAMI := *savedImages[0].ImageId

	return selectedAMI, nil
}

// trustedSource: helper recieved from https://ubuntu.com/tutorials/search-and-launch-ubuntu-22-04-in-aws-using-cli#2-search-for-the-right-ami
func trustedSource(id string) bool {
	// 679593333241
	// 099720109477
	if strings.Compare(id, "679593333241") != 0 && strings.Compare(id, "099720109477") != 0 {
		return false
	}
	return true
}

func (client *AwsClient) GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error) {
	azs, err := client.ec2Client.DescribeAvailabilityZones(awsCtx, &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
	})
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "failed to describe availability zones", "Reason", err),
		)
	}

	return azs, nil
}

func (client *AwsClient) BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic, err := client.ec2Client.CreateNetworkInterface(ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Network Interface", "Reason", err),
		)
	}

	nicExistsWaiter := ec2.NewNetworkInterfaceAvailableWaiter(client.ec2Client, func(nicwaiter *ec2.NetworkInterfaceAvailableWaiterOptions) {
		nicwaiter.MinDelay = initialNicMinDelay
		nicwaiter.MaxDelay = initialNicMaxDelay
	})

	describeNICInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{*nic.NetworkInterface.NetworkInterfaceId},
	}

	err = nicExistsWaiter.Wait(awsCtx, describeNICInput, initialNicWaiterTime)

	if err != nil {
		return nil, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Waiting for Network Interface", "Reason", err),
		)
	}

	return nic, err
}

func (client *AwsClient) BeginCreateSubNet(ctx context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := client.ec2Client.CreateSubnet(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Subnet", "Reason", err),
		)
	}

	subnetExistsWaiter := ec2.NewSubnetAvailableWaiter(client.ec2Client, func(subnetwaiter *ec2.SubnetAvailableWaiterOptions) {
		subnetwaiter.MinDelay = initialSubnetMinDelay
		subnetwaiter.MaxDelay = initialSubnetMaxDelay
	})

	describeSubnetInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{*subnet.Subnet.SubnetId},
	}

	err = subnetExistsWaiter.Wait(ctx, describeSubnetInput, initialSubnetWaiterTime)
	if err != nil {
		return nil, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Waiting for Subnet", "Reason", err),
		)
	}

	return subnet, err
}

func (client *AwsClient) BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

	runResult, err := client.ec2Client.RunInstances(ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Instance", "Reason", err),
		)
	}

	return runResult, err
}

func (client *AwsClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := client.ec2Client.CreateInternetGateway(awsCtx, &gatewayparameter)
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Internet Gateway", "Reason", err),
		)
	}

	_, err = client.ec2Client.AttachInternetGateway(awsCtx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Attaching Internet Gateway", "Reason", err),
		)
	}

	mainStateDocument.CloudInfra.Aws.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId

	routeTable, err := client.ec2Client.CreateRouteTable(awsCtx, &routeTableparameter)
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Route Table", "Reason", err),
		)
	}

	mainStateDocument.CloudInfra.Aws.RouteTableID = *routeTable.RouteTable.RouteTableId

	_, err = client.ec2Client.CreateRoute(awsCtx, &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
		RouteTableId:         aws.String(mainStateDocument.CloudInfra.Aws.RouteTableID),
	})
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Route", "Reason", err),
		)
	}

	for i := 0; i < len(mainStateDocument.CloudInfra.Aws.SubnetIDs); i++ {
		_, err = client.ec2Client.AssociateRouteTable(awsCtx, &ec2.AssociateRouteTableInput{
			RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
			SubnetId:     aws.String(mainStateDocument.CloudInfra.Aws.SubnetIDs[i]),
		})
		if err != nil {
			return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error assiciating route table", "Reason", err),
			)
		}
	}

	return routeTable, createInternetGateway, err
}

func (client *AwsClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := client.ec2Client.CreateVpc(awsCtx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating VPC", "Reason", err),
		)
	}

	vpcExistsWaiter := ec2.NewVpcExistsWaiter(client.ec2Client, func(vpcwaiter *ec2.VpcExistsWaiterOptions) {
		vpcwaiter.MinDelay = 1
		vpcwaiter.MaxDelay = 5
	})

	describeVpcInput := &ec2.DescribeVpcsInput{
		VpcIds: []string{*vpc.Vpc.VpcId},
	}

	err = vpcExistsWaiter.Wait(awsCtx, describeVpcInput, initialSubnetWaiterTime)
	if err != nil {
		return nil, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Waiting for VPC", "Reason", err),
		)
	}

	return vpc, err
}

func (client *AwsClient) BeginDeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	_, err := client.ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
	})
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Deleting VPC", "Reason", err),
		)
	}

	return nil

}

func (client *AwsClient) BeginDeleteNIC(nicID string) error {
	initialWater := time.Now()
	// TODO(praful): use the helpers.Backoff
	//  also why do we wait for the nic to be available when it is deleting
	for {
		nic, err := client.ec2Client.DescribeNetworkInterfaces(awsCtx, &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{nicID},
		})
		if err != nil {
			return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error Describing Network Interface", "Reason", err),
			)
		}
		if nic.NetworkInterfaces[0].Status == "available" {
			break
		}
		if time.Since(initialWater) > initialNicDeletionWaiterTime {
			return ksctlErrors.ErrTimeOut.Wrap(
				log.NewError(awsCtx, "Error Waiting for Network Interface Timeout", "Reason", err),
			)
		}
	}
	_, err := client.ec2Client.DeleteNetworkInterface(awsCtx, &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(nicID),
	})
	if err != nil {
		log.Success(awsCtx, "skipped already deleted the nic", "id", nicID)
		return nil
	}

	return nil
}

func (client *AwsClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	_, err := client.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGrpID),
	})
	if err != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Deleting Security Group", "Reason", err),
		)
	}
	return nil
}

func (client *AwsClient) BeginDeleteSubNet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetID string) error {

	_, err := client.ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Deleting Subnet", "Reason", err),
		)
	}

	return nil

}

func (client *AwsClient) BeginDeleteVM(instanceID string) error {

	_, err := client.ec2Client.TerminateInstances(awsCtx, &ec2.TerminateInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "failed to delete instance", "Reason", err),
		)
	}

	ec2TerminatedWaiter := ec2.NewInstanceTerminatedWaiter(client.ec2Client, func(itwo *ec2.InstanceTerminatedWaiterOptions) {
		itwo.MinDelay = initialInstanceMinDelay
		itwo.MaxDelay = initialInstanceMaxDelay
	})

	describeEc2Inp := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	err = ec2TerminatedWaiter.Wait(awsCtx, describeEc2Inp, instanceInitialTerminationTime)
	if err != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "failed to wait for instance to terminate", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) InstanceInitialWaiter(ctx context.Context, instanceID string) error {

	instanceExistsWaiter := ec2.NewInstanceStatusOkWaiter(client.ec2Client, func(instancewaiter *ec2.InstanceStatusOkWaiterOptions) {
		instancewaiter.MinDelay = initialInstanceMinDelay
		instancewaiter.MaxDelay = initialInstanceMaxDelay
	})

	describeInstanceInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{instanceID},
	}

	err := instanceExistsWaiter.Wait(ctx, describeInstanceInput, instanceInitialWaiterTime)
	if err != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Waiting for Instance", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

	naclresp, err := client.ec2Client.CreateNetworkAcl(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Network ACL", "Reason", err),
		)
	}

	_, err = client.ec2Client.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
		NetworkAclId: aws.String(*naclresp.NetworkAcl.NetworkAclId),
		RuleNumber:   aws.Int32(100),
		Protocol:     aws.String("-1"),
		RuleAction:   types.RuleActionAllow,
		CidrBlock:    aws.String("0.0.0.0/0"),
		Egress:       aws.Bool(true),
	})
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Network ACL Entry", "Reason", err),
		)
	}

	return naclresp, nil
}

func (client *AwsClient) BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup, err := client.ec2Client.CreateSecurityGroup(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Security Group", "Reason", err),
		)
	}

	return securitygroup, nil
}

func (client *AwsClient) BeginDeleteVirtNet(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	if mainStateDocument.CloudInfra.Aws.RouteTableID == "" {
		log.Success(awsCtx, "skipped already deleted the route table")
	} else {
		_, err := client.ec2Client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: aws.String(mainStateDocument.CloudInfra.Aws.RouteTableID),
		})
		if err != nil {
			return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error Deleting Route Table", "Reason", err),
			)
		}
		log.Success(awsCtx, "deleted the route table", "id", mainStateDocument.CloudInfra.Aws.RouteTableID)
		mainStateDocument.CloudInfra.Aws.RouteTableID = ""
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
	}

	if mainStateDocument.CloudInfra.Aws.GatewayID == "" {
		log.Success(awsCtx, "skipped already deleted the internet gateway")
	} else {
		_, err := client.ec2Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
			VpcId:             aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
		})

		if err != nil {
			return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error Detaching Internet Gateway", "Reason", err),
			)
		}

		_, err = client.ec2Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
		})
		if err != nil {
			return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error Deleting Internet Gateway", "Reason", err),
			)
		}
		id := mainStateDocument.CloudInfra.Aws.GatewayID
		mainStateDocument.CloudInfra.Aws.GatewayID = ""
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}

		log.Success(awsCtx, "deleted the internet gateway", "id", id)

	}

	if mainStateDocument.CloudInfra.Aws.NetworkAclID == "" {
		// TODO(praful)!: resolve this empty branch
	} else {

		_, err := client.ec2Client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: aws.String(mainStateDocument.CloudInfra.Aws.NetworkAclID),
		})

		if err != nil {
			return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error Deleting Network ACL", "Reason", err),
			)
		}

		id := mainStateDocument.CloudInfra.Aws.NetworkAclID
		mainStateDocument.CloudInfra.Aws.NetworkAclID = ""
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
		log.Success(awsCtx, "deleted the network acl", "id", id)

	}
	return nil
}

func (client *AwsClient) DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error) {

	instanceipinput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}

	instanceinforesponse, err := client.ec2Client.DescribeInstances(ctx, instanceipinput)
	if err != nil {
		return instanceinforesponse, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Describing Instances", "Reason", err),
		)
	}

	return instanceinforesponse, nil
}

func (client *AwsClient) AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error {

	_, err := client.ec2Client.AuthorizeSecurityGroupEgress(ctx, &parameter)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Authorizing Security Group Egress", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) DeleteSSHKey(ctx context.Context, name string) error {

	_, err := client.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Deleting Key Pair", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) InitClient(storage ksctlTypes.StorageFactory) error {

	err := client.setRequiredENVVAR(storage, awsCtx)
	if err != nil {
		return err
	}

	client.storage = storage
	client.region = mainStateDocument.Region
	session, err := newEC2Client(client.region)
	if err != nil {
		return err
	}

	client.config = &session
	client.stsClient = sts.NewFromConfig(session)
	client.ec2Client = ec2.NewFromConfig(session)
	client.eksClient = eks.NewFromConfig(session)
	client.iam = iam.NewFromConfig(session)
	return nil
}

func (client *AwsClient) ImportKeyPair(ctx context.Context, input *ec2.ImportKeyPairInput) error {

	if _, err := client.ec2Client.ImportKeyPair(ctx, input); err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Importing Key Pair", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) setRequiredENVVAR(storage ksctlTypes.StorageFactory, _ context.Context) error {

	envacesskeyid := os.Getenv("AWS_ACCESS_KEY_ID")
	envkeysecret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if len(envacesskeyid) != 0 && len(envkeysecret) != 0 {
		return nil
	}

	msg := "environment vars not set:"

	if len(envacesskeyid) == 0 {
		msg = msg + " AWS_ACCESS_KEY_ID"
	}

	if len(envkeysecret) == 0 {
		msg = msg + " AWS_SECRET_ACCESS_KEY"
	}

	log.Debug(awsCtx, msg)

	credentialsDocument, err := storage.ReadCredentials(consts.CloudAws)
	if err != nil {
		return err
	}

	if credentialsDocument.Aws == nil {
		return ksctlErrors.ErrNilCredentials.Wrap(
			log.NewError(awsCtx, "no entry for aws present"),
		)
	}

	err = os.Setenv("AWS_ACCESS_KEY_ID", credentialsDocument.Aws.AccessKeyId)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(
			log.NewError(awsCtx, "failed to set environmenet variable", "Reason", err),
		)
	}
	err = os.Setenv("AWS_SECRET_ACCESS_KEY", credentialsDocument.Aws.SecretAccessKey)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(
			log.NewError(awsCtx, "failed to set environmenet variable", "Reason", err),
		)
	}
	return nil
}

func (client *AwsClient) ListLocations() ([]string, error) {

	parameter := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	result, err := client.ec2Client.DescribeRegions(awsCtx, parameter)
	if err != nil {
		return nil, ksctlErrors.ErrInvalidCloudRegion.Wrap(
			log.NewError(awsCtx, "Error Describing Regions", "Reason", err),
		)
	}

	validRegion := []string{}
	for _, region := range result.Regions {
		validRegion = append(validRegion, *region.RegionName)
	}
	return validRegion, nil
}

func (client *AwsClient) ListVMTypes() (ec2.DescribeInstanceTypesOutput, error) {
	var vmTypes ec2.DescribeInstanceTypesOutput
	input := &ec2.DescribeInstanceTypesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("current-generation"),
				Values: []string{"true"},
			},
		},
	}

	for {
		output, err := client.ec2Client.DescribeInstanceTypes(awsCtx, input)
		if err != nil {
			return vmTypes, ksctlErrors.ErrInvalidCloudVMSize.Wrap(
				log.NewError(awsCtx, "Error Describing Instance Types", "Reason", err),
			)
		}
		vmTypes.InstanceTypes = append(vmTypes.InstanceTypes, output.InstanceTypes...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return vmTypes, nil
}

func (client *AwsClient) ModifyVpcAttribute(ctx context.Context) error {

	modifyvpcinput := &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := client.ec2Client.ModifyVpcAttribute(awsCtx, modifyvpcinput)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Modifying VPC Attribute", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) ModifySubnetAttribute(ctx context.Context, i int) error {

	modifyusbnetinput := &ec2.ModifySubnetAttributeInput{
		SubnetId: aws.String(mainStateDocument.CloudInfra.Aws.SubnetIDs[i]),
		MapPublicIpOnLaunch: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := client.ec2Client.ModifySubnetAttribute(awsCtx, modifyusbnetinput)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Modifying Subnet Attribute", "Reason", err),
		)
	}

	return nil
}

func (client *AwsClient) SetRegion(region string) {
	client.region = region
	log.Debug(awsCtx, "region set", "code", client.region)
}

func (client *AwsClient) SetVpc(vpc string) string {
	client.vpc = vpc
	log.Print(awsCtx, "vpc set", "name", client.vpc)

	return client.vpc
}

func (client *AwsClient) BeginCreateEKS(ctx context.Context, paramter *eks.CreateClusterInput) (*eks.CreateClusterOutput, error) {

	resp, err := client.eksClient.CreateCluster(ctx, paramter)
	if err != nil {
		return nil, err
	}

	mainStateDocument.CloudInfra.Aws.ManagedClusterName = *resp.Cluster.Name
	mainStateDocument.CloudInfra.Aws.ManagedClusterArn = *resp.Cluster.Arn
	if err := client.storage.Write(mainStateDocument); err != nil {
		return nil, err
	}

	waiter := eks.NewClusterActiveWaiter(client.eksClient)

	describeCluster := eks.DescribeClusterInput{
		Name: resp.Cluster.Name,
	}
	xxxx, err := waiter.WaitForOutput(ctx, &describeCluster, managedClusterActiveWaiter)
	if err != nil {
		return nil, err
	}
	dump.NewWithOptions(dump.SkipPrivate()).Println(xxxx.Cluster)
	return resp, nil
}

func (client *AwsClient) BeginCreateNodeGroup(ctx context.Context, paramter *eks.CreateNodegroupInput) (*eks.CreateNodegroupOutput, error) {
	resp, err := client.eksClient.CreateNodegroup(ctx, paramter)
	if err != nil {
		return nil, err
	}
	mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName = *resp.Nodegroup.NodegroupName
	mainStateDocument.CloudInfra.Aws.ManagedNodeGroupArn = *resp.Nodegroup.NodegroupArn
	if err := client.storage.Write(mainStateDocument); err != nil {
		return nil, err
	}

	waiter := eks.NewNodegroupActiveWaiter(client.eksClient)

	describeNodeGroup := &eks.DescribeNodegroupInput{
		NodegroupName: resp.Nodegroup.NodegroupName,
		ClusterName:   aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
	}
	xxxx, err := waiter.WaitForOutput(ctx, describeNodeGroup, managedNodeGroupActiveWaiter)
	// TODO: should we use this if we are going to use WaitForOutput eks.DescribeNodegroupOutput
	if err != nil {
		return nil, err
	}
	dump.NewWithOptions(dump.SkipPrivate()).Println("xxxx==>", xxxx)

	return resp, nil
}

func (client *AwsClient) BeginDeleteNodeGroup(ctx context.Context, parameter *eks.DeleteNodegroupInput) (*eks.DeleteNodegroupOutput, error) {

	resp, err := client.eksClient.DeleteNodegroup(ctx, parameter)
	if err != nil {
		return nil, err
	}

	waiter := eks.NewNodegroupDeletedWaiter(client.eksClient)

	describeNodeGroup := eks.DescribeNodegroupInput{
		NodegroupName: aws.String(*resp.Nodegroup.NodegroupName),
	}

	_, err = waiter.WaitForOutput(ctx, &describeNodeGroup, managedNodeGroupActiveWaiter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (client *AwsClient) BeginDeleteManagedCluster(ctx context.Context, parameter *eks.DeleteClusterInput) (*eks.DeleteClusterOutput, error) {

	resp, err := client.eksClient.DeleteCluster(ctx, parameter)
	if err != nil {
		return nil, err
	}

	waiter := eks.NewClusterDeletedWaiter(client.eksClient)

	describeCluster := eks.DescribeClusterInput{
		Name: aws.String(*resp.Cluster.Name),
	}

	_, err = waiter.WaitForOutput(ctx, &describeCluster, managedClusterActiveWaiter)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (client *AwsClient) DescribeCluster(ctx context.Context, parameter *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error) {
	resp, err := client.eksClient.DescribeCluster(ctx, parameter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (client *AwsClient) BeginCreateIAM(ctx context.Context, node string, parameter *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	createRoleResp, err := client.iam.CreateRole(ctx, parameter)
	if err != nil {
		return nil, err
	}

	switch node {
	case "controlplane":
		attachClusterPolicyInput := &iam.AttachRolePolicyInput{
			RoleName:  aws.String(*createRoleResp.Role.RoleName),
			PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"),
		}

		_, err = client.iam.AttachRolePolicy(ctx, attachClusterPolicyInput)
		if err != nil {
			return nil, err
		}
	case "worker":
		fmt.Println("AmazonEKSWorkerNodePolicy")
		policyArn := []string{"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy", "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy", "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"}
		for _, policy := range policyArn {
			attachWorkerPolicyInput := &iam.AttachRolePolicyInput{
				RoleName:  aws.String(*createRoleResp.Role.RoleName),
				PolicyArn: aws.String(policy),
			}

			_, err = client.iam.AttachRolePolicy(ctx, attachWorkerPolicyInput)
			if err != nil {
				return nil, err
			}
		}
	}

	return createRoleResp, nil
}

func NewSTSTokenRetriver(client StsPresignClientInteface) STSTokenRetriever {
	return STSTokenRetriever{PresignClient: client}
}

func newCustomHTTPPresignerV4(client sts.HTTPPresignerV4, headers map[string]string) sts.HTTPPresignerV4 {
	return &customHTTPPresignerV4{
		client:  client,
		headers: headers,
	}
}

func (p *customHTTPPresignerV4) PresignHTTP(
	ctx context.Context, credentials aws.Credentials, r *http.Request,
	payloadHash string, service string, region string, signingTime time.Time,
	optFns ...func(*v4.SignerOptions),
) (url string, signedHeader http.Header, err error) {
	for key, val := range p.headers {
		r.Header.Add(key, val)
	}
	return p.client.PresignHTTP(ctx, credentials, r, payloadHash, service, region, signingTime, optFns...)
}
func (s *STSTokenRetriever) GetToken(ctx context.Context, clusterName string, cfg aws.Config) string {
	out, err := s.PresignClient.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(opt *sts.PresignOptions) {
		k8sHeader := "x-k8s-aws-id"
		opt.Presigner = newCustomHTTPPresignerV4(opt.Presigner, map[string]string{
			k8sHeader:       clusterName,
			"X-Amz-Expires": "60",
		})
	})
	if err != nil {
		panic(err)
	}
	tokenPrefix := "k8s-aws-v1."
	token := fmt.Sprintf("%s%s", tokenPrefix, base64.RawURLEncoding.EncodeToString([]byte(out.URL))) //RawURLEncoding
	return token

}

func (client *AwsClient) GetKubeConfig(ctx context.Context, clusterName string) (string, error) {
	clusterData, err := client.eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return "", err
	}
	preSignClient := sts.NewPresignClient(client.stsClient)
	tokenRetriver := NewSTSTokenRetriver(preSignClient)
	token := tokenRetriver.GetToken(context.Background(), clusterName, *client.config)

	data := KubeConfigData{
		ClusterEndpoint:          *clusterData.Cluster.Endpoint,
		CertificateAuthorityData: *clusterData.Cluster.CertificateAuthority.Data,
		ClusterName:              *clusterData.Cluster.Name,
		Token:                    token,
	}

	tmpl, err := template.New("kubeconfig").Parse(kubeconfigTemplate)
	if err != nil {
		fmt.Println("Error creating template:", err)
		os.Exit(1)
	}

	var kubeconfig bytes.Buffer
	err = tmpl.Execute(&kubeconfig, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		os.Exit(1)
	}

	return kubeconfig.String(), nil
}

func (client *AwsClient) BeginDeleteIAM(ctx context.Context, parameter *iam.DeleteRoleInput, node string) (*iam.DeleteRoleOutput, error) {

	switch node {
	case "controlplane":
		detachClusterPolicyInput := &iam.DetachRolePolicyInput{
			RoleName:  parameter.RoleName,
			PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"),
		}

		_, err := client.iam.DetachRolePolicy(ctx, detachClusterPolicyInput)
		if err != nil {
			return nil, err
		}

	case "worker":
		policyArn := []string{"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy", "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy", "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"}
		for _, policy := range policyArn {

			detachWorkerPolicyInput := &iam.DetachRolePolicyInput{
				RoleName:  parameter.RoleName,
				PolicyArn: aws.String(policy),
			}

			_, err := client.iam.DetachRolePolicy(ctx, detachWorkerPolicyInput)
			if err != nil {
				return nil, err
			}

		}
	}

	resp, err := client.iam.DeleteRole(ctx, parameter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
