package aws

import (
	"context"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"
)

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
)

func ProvideClient() AwsGo {
	return &AwsGoClient{}
}

func ProvideMockClient() AwsGo {
	return &AwsGoMockClient{}
}

type AwsGo interface {
	AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error

	InitClient(storage ksctlTypes.StorageFactory) error

	ListLocations() ([]string, error)

	ListVMTypes() (ec2.DescribeInstanceTypesOutput, error)

	BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error)

	BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteVirtNet(ctx context.Context, storage ksctlTypes.StorageFactory) error

	BeginDeleteSubNet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetID string) error

	DeleteSSHKey(ctx context.Context, name string) error

	BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)

	BeginDeleteVM(vmname string) error

	BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error)

	BeginDeleteNIC(nicID string) error

	BeginDeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory) error

	BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error)

	BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)

	BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error

	DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error)

	FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error)
	GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error)
	AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error

	ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error
	InstanceInitialWaiter(ctx context.Context, instanceInput string) error

	ModifyVpcAttribute(ctx context.Context) error
	ModifySubnetAttribute(ctx context.Context) error

	SetRegion(string)
	SetVpc(string) string
}

type AwsGoClient struct {
	region    string
	vpc       string
	ec2Client *ec2.Client
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

func (awsclient *AwsGoClient) AuthorizeSecurityGroupIngress(ctx context.Context,
	parameter ec2.AuthorizeSecurityGroupIngressInput) error {

	_, err := awsclient.ec2Client.AuthorizeSecurityGroupIngress(ctx, &parameter)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Authorizing Security Group Ingress", "Reason", err),
		)
	}

	return nil
}

func (awsclient *AwsGoClient) FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error) {
	resp, err := awsclient.ec2Client.DescribeImages(awsCtx, filter)
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

func (awsclient *AwsGoClient) GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error) {
	azs, err := awsclient.ec2Client.DescribeAvailabilityZones(awsCtx, &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
	})
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "failed to describe availability zones", "Reason", err),
		)
	}

	return azs, nil
}

func (awsclient *AwsGoClient) BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic, err := awsclient.ec2Client.CreateNetworkInterface(ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Network Interface", "Reason", err),
		)
	}

	nicExistsWaiter := ec2.NewNetworkInterfaceAvailableWaiter(awsclient.ec2Client, func(nicwaiter *ec2.NetworkInterfaceAvailableWaiterOptions) {
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

func (awsclient *AwsGoClient) BeginCreateSubNet(ctx context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := awsclient.ec2Client.CreateSubnet(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Subnet", "Reason", err),
		)
	}

	subnetExistsWaiter := ec2.NewSubnetAvailableWaiter(awsclient.ec2Client, func(subnetwaiter *ec2.SubnetAvailableWaiterOptions) {
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

func (awsclient *AwsGoClient) BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

	runResult, err := awsclient.ec2Client.RunInstances(ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Instance", "Reason", err),
		)
	}

	return runResult, err
}

func (awsclient *AwsGoClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := awsclient.ec2Client.CreateInternetGateway(awsCtx, &gatewayparameter)
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Internet Gateway", "Reason", err),
		)
	}

	_, err = awsclient.ec2Client.AttachInternetGateway(awsCtx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Attaching Internet Gateway", "Reason", err),
		)
	}

	mainStateDocument.CloudInfra.Aws.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId

	routeTable, err := awsclient.ec2Client.CreateRouteTable(awsCtx, &routeTableparameter)
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Route Table", "Reason", err),
		)
	}

	mainStateDocument.CloudInfra.Aws.RouteTableID = *routeTable.RouteTable.RouteTableId

	_, err = awsclient.ec2Client.CreateRoute(awsCtx, &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
		RouteTableId:         aws.String(mainStateDocument.CloudInfra.Aws.RouteTableID),
	})
	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Route", "Reason", err),
		)
	}

	_, err = awsclient.ec2Client.AssociateRouteTable(awsCtx, &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
		SubnetId:     aws.String(mainStateDocument.CloudInfra.Aws.SubnetID),
	})

	if err != nil {
		return nil, nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error assiciating route table", "Reason", err),
		)
	}

	return routeTable, createInternetGateway, err
}

func (awsclient *AwsGoClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := awsclient.ec2Client.CreateVpc(awsCtx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating VPC", "Reason", err),
		)
	}

	vpcExistsWaiter := ec2.NewVpcExistsWaiter(awsclient.ec2Client, func(vpcwaiter *ec2.VpcExistsWaiterOptions) {
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

func (awsclient *AwsGoClient) BeginDeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	_, err := awsclient.ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
	})
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Deleting VPC", "Reason", err),
		)
	}

	return nil

}

func (awsclient *AwsGoClient) BeginDeleteNIC(nicID string) error {
	initialWater := time.Now()
	// TODO(praful): use the helpers.Backoff
	// also why do we wait for the nic to be available when it is deleting
	for {
		nic, err := awsclient.ec2Client.DescribeNetworkInterfaces(awsCtx, &ec2.DescribeNetworkInterfacesInput{
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
	_, err := awsclient.ec2Client.DeleteNetworkInterface(awsCtx, &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(nicID),
	})
	if err != nil {
		// TODO(praful): need to fix this
		log.Success(awsCtx, "skipped already deleted the nic", "id", nicID)
		return nil
	}

	return nil
}

func (awsclient *AwsGoClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	_, err := awsclient.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGrpID),
	})
	if err != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Deleting Security Group", "Reason", err),
		)
	}
	return nil
}

func (awsclient *AwsGoClient) BeginDeleteSubNet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetID string) error {

	_, err := awsclient.ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(awsCtx, "Error Deleting Subnet", "Reason", err),
		)
	}

	return nil

}

func (awsgo *AwsGoClient) BeginDeleteVM(instanceID string) error {

	_, err := awsgo.ec2Client.TerminateInstances(awsCtx, &ec2.TerminateInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "failed to delete instance", "Reason", err),
		)
	}

	ec2TerminatedWaiter := ec2.NewInstanceTerminatedWaiter(awsgo.ec2Client, func(itwo *ec2.InstanceTerminatedWaiterOptions) {
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

func (awsclient *AwsGoClient) InstanceInitialWaiter(ctx context.Context, instanceID string) error {

	instanceExistsWaiter := ec2.NewInstanceStatusOkWaiter(awsclient.ec2Client, func(instancewaiter *ec2.InstanceStatusOkWaiterOptions) {
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

func (awsclient *AwsGoClient) BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

	naclresp, err := awsclient.ec2Client.CreateNetworkAcl(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Network ACL", "Reason", err),
		)
	}

	_, err = awsclient.ec2Client.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
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

func (awsclient *AwsGoClient) BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup, err := awsclient.ec2Client.CreateSecurityGroup(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Creating Security Group", "Reason", err),
		)
	}

	return securitygroup, nil
}

func (awsclient *AwsGoClient) BeginDeleteVirtNet(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	if mainStateDocument.CloudInfra.Aws.RouteTableID == "" {
		log.Success(awsCtx, "skipped already deleted the route table")
	} else {
		_, err := awsclient.ec2Client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
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
		_, err := awsclient.ec2Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
			VpcId:             aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
		})

		if err != nil {
			return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(awsCtx, "Error Detaching Internet Gateway", "Reason", err),
			)
		}

		_, err = awsclient.ec2Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
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

		_, err := awsclient.ec2Client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
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

func (awsclient *AwsGoClient) DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error) {

	instanceipinput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}

	instanceinforesponse, err := awsclient.ec2Client.DescribeInstances(ctx, instanceipinput)
	if err != nil {
		return instanceinforesponse, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Describing Instances", "Reason", err),
		)
	}

	return instanceinforesponse, nil
}

func (awsclient *AwsGoClient) AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error {

	_, err := awsclient.ec2Client.AuthorizeSecurityGroupEgress(ctx, &parameter)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Authorizing Security Group Egress", "Reason", err),
		)
	}

	return nil
}

func (awsclient *AwsGoClient) DeleteSSHKey(ctx context.Context, name string) error {

	_, err := awsclient.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Deleting Key Pair", "Reason", err),
		)
	}

	return nil
}

func (awsclient *AwsGoClient) InitClient(storage ksctlTypes.StorageFactory) error {

	err := awsclient.setRequiredENVVAR(storage, awsCtx)
	if err != nil {
		return err
	}

	awsclient.storage = storage
	awsclient.region = mainStateDocument.Region
	ec2client, err := newEC2Client(awsclient.region)
	if err != nil {
		return err
	}

	awsclient.ec2Client = ec2.NewFromConfig(ec2client)
	return nil
}

func (awsclient *AwsGoClient) ImportKeyPair(ctx context.Context, input *ec2.ImportKeyPairInput) error {

	if _, err := awsclient.ec2Client.ImportKeyPair(ctx, input); err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Importing Key Pair", "Reason", err),
		)
	}

	return nil
}

func (awsclient *AwsGoClient) setRequiredENVVAR(storage ksctlTypes.StorageFactory, _ context.Context) error {

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

	credentials, err := storage.ReadCredentials(consts.CloudAws)
	if err != nil {
		return err
	}

	if credentials.Aws == nil {
		return ksctlErrors.ErrNilCredentials.Wrap(
			log.NewError(awsCtx, "no entry for aws present"),
		)
	}

	err = os.Setenv("AWS_ACCESS_KEY_ID", credentials.Aws.AccessKeyId)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(
			log.NewError(awsCtx, "failed to set environmenet variable", "Reason", err),
		)
	}
	err = os.Setenv("AWS_SECRET_ACCESS_KEY", credentials.Aws.SecretAccessKey)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(
			log.NewError(awsCtx, "failed to set environmenet variable", "Reason", err),
		)
	}
	return nil
}

func (awsclient *AwsGoClient) ListLocations() ([]string, error) {

	parameter := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	result, err := awsclient.ec2Client.DescribeRegions(awsCtx, parameter)
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

func (awsclient *AwsGoClient) ListVMTypes() (ec2.DescribeInstanceTypesOutput, error) {
	var vmTypes ec2.DescribeInstanceTypesOutput
	input := &ec2.DescribeInstanceTypesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("current-generation"),
				Values: []string{"true"},
			},
		},
	}

	// TODO(praful): why is their a infinite loop here
	for {
		output, err := awsclient.ec2Client.DescribeInstanceTypes(awsCtx, input)
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

func (awsclient *AwsGoClient) ModifyVpcAttribute(ctx context.Context) error {

	modifyvpcinput := &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := awsclient.ec2Client.ModifyVpcAttribute(awsCtx, modifyvpcinput)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Modifying VPC Attribute", "Reason", err),
		)
	}

	return nil
}

func (awsclient *AwsGoClient) ModifySubnetAttribute(ctx context.Context) error {

	modifyusbnetinput := &ec2.ModifySubnetAttributeInput{
		SubnetId: aws.String(mainStateDocument.CloudInfra.Aws.SubnetID),
		MapPublicIpOnLaunch: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := awsclient.ec2Client.ModifySubnetAttribute(awsCtx, modifyusbnetinput)
	if err != nil {
		return ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(awsCtx, "Error Modifying Subnet Attribute", "Reason", err),
		)
	}

	return nil
}

func (awsclient *AwsGoClient) SetRegion(region string) {
	awsclient.region = region
	log.Debug(awsCtx, "region set", "code", awsclient.region)
}

func (awsclient *AwsGoClient) SetVpc(vpc string) string {
	awsclient.vpc = vpc
	log.Print(awsCtx, "vpc set", "name", awsclient.vpc)

	return awsclient.vpc
}

type AwsGoMockClient struct {
	region string
}

func (*AwsGoMockClient) AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error {
	return nil
}

func (*AwsGoMockClient) AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error {
	return nil
}

func (*AwsGoMockClient) BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic := &ec2.CreateNetworkInterfaceOutput{
		NetworkInterface: &types.NetworkInterface{
			NetworkInterfaceId: aws.String("test-nic-1234567890"),
			TagSet: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(*parameter.TagSpecifications[0].Tags[0].Value),
				},
			},
		},
	}

	return nic, nil

}

func (awsgoclient *AwsGoMockClient) BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet := &ec2.CreateSubnetOutput{
		Subnet: &types.Subnet{
			SubnetId: aws.String("3456d25f36g474g546"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(*parameter.TagSpecifications[0].Tags[0].Value),
				},
			},
		},
	}

	return subnet, nil
}

func (*AwsGoMockClient) BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

	instances := &ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String("test-instance-1234567890"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("test-instance"),
					},
				},
			},
		},
	}

	return instances, nil
}

func (*AwsGoMockClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	routeTable := &ec2.CreateRouteTableOutput{
		RouteTable: &types.RouteTable{
			RouteTableId: aws.String("3456d25f36g474g546"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-route-table"),
				},
			},
		},
	}
	createInternetGateway := &ec2.CreateInternetGatewayOutput{
		InternetGateway: &types.InternetGateway{
			InternetGatewayId: aws.String("3456d25f36g474g546"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-internet-gateway"),
				},
			},
		},
	}

	return routeTable, createInternetGateway, nil
}

func (awsclient *AwsGoMockClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc := &ec2.CreateVpcOutput{
		Vpc: &types.Vpc{
			VpcId: aws.String("3456d25f36g474g546"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(*parameter.TagSpecifications[0].Tags[0].Value),
				},
			},
		},
	}
	return vpc, nil
}

func (*AwsGoMockClient) BeginDeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	mainStateDocument.CloudInfra.Aws.VpcId = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(awsCtx, "Error Writing State File", "Reason", err)
	}

	log.Success(awsCtx, "deleted the vpc", "id", mainStateDocument.CloudInfra.Aws.VpcId)

	return nil

}

func (*AwsGoMockClient) BeginDeleteNIC(nicID string) error {

	return nil
}

func (*AwsGoMockClient) FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error) {
	return "ami-1234567890", nil
}

func (*AwsGoMockClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	return nil
}

func (*AwsGoMockClient) GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error) {
	return &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []types.AvailabilityZone{
			{
				ZoneName: aws.String("us-east-1a"),
			},
		},
	}, nil
}

func (*AwsGoMockClient) BeginDeleteSubNet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetID string) error {

	mainStateDocument.CloudInfra.Aws.SubnetID = ""
	mainStateDocument.CloudInfra.Aws.SubnetName = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(awsCtx, "Error Writing State File", "Reason", err)
	}

	log.Success(awsCtx, "deleted the subnet ", mainStateDocument.CloudInfra.Aws.SubnetName)

	return nil

}

func (*AwsGoMockClient) BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

	naclresp := &ec2.CreateNetworkAclOutput{
		NetworkAcl: &types.NetworkAcl{
			NetworkAclId: aws.String("test-nacl-1234567890"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-nacl"),
				},
			},
		},
	}

	return naclresp, nil
}

func (*AwsGoMockClient) BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup := &ec2.CreateSecurityGroupOutput{
		GroupId: aws.String("test-security-group-1234567890"),
	}

	return securitygroup, nil
}

func (*AwsGoMockClient) BeginDeleteVM(vmname string) error {
	return nil
}

func (*AwsGoMockClient) BeginDeleteVirtNet(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	mainStateDocument.CloudInfra.Aws.GatewayID = ""
	mainStateDocument.CloudInfra.Aws.RouteTableID = ""
	mainStateDocument.CloudInfra.Aws.NetworkAclID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(awsCtx, "Error Writing State File", "Reason", err)
	}

	return nil
}

func (*AwsGoMockClient) CreateSSHKey() error {
	return nil
}

func (*AwsGoMockClient) DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error) {

	instanceinforesponse := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
						PublicIpAddress:  aws.String("A.B.C.D"),
						PrivateIpAddress: aws.String("192.169.1.2"),
					},
				},
			},
		},
	}

	return instanceinforesponse, nil
}

func (*AwsGoMockClient) DeleteSSHKey(ctx context.Context, name string) error {

	return nil
}

func (*AwsGoMockClient) InstanceInitialWaiter(ctx context.Context, instanceID string) error {
	return nil
}

func (*AwsGoMockClient) InitClient(storage ksctlTypes.StorageFactory) error {
	return nil
}

func (*AwsGoMockClient) ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error {

	return nil
}

func (awsclient *AwsGoMockClient) ListLocations() ([]string, error) {

	return []string{"fake-region"}, nil
}

func (*AwsGoMockClient) ListVMTypes() (ec2.DescribeInstanceTypesOutput, error) {
	return ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []types.InstanceTypeInfo{
			{
				InstanceType: "fake",
			},
		},
	}, nil
}

func (*AwsGoMockClient) ModifyVpcAttribute(ctx context.Context) error {
	return nil
}

func (*AwsGoMockClient) ModifySubnetAttribute(ctx context.Context) error {
	return nil
}
func (a *AwsGoMockClient) SetRegion(string) {
	a.region = "fake-region"
}

func (*AwsGoMockClient) SetVpc(string) string {
	return "fake-vpc"
}
