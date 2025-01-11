// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !testing_aws

package aws

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
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
    user: aws
  name: {{.ClusterName}}
current-context: {{.ClusterName}}
kind: Config
preferences: {}
users:
- name: aws
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
	managedClusterDeletionWaiter   = time.Minute * 10
	managedNodeGroupActiveWaiter   = time.Minute * 10
	managedNodeGroupDeletionWaiter = time.Minute * 15
)

func ProvideClient() CloudSDK {
	return &AwsClient{}
}

type AwsClient struct {
	stsClient *sts.Client
	ec2Client *ec2.Client
	eksClient *eks.Client
	config    *aws.Config
	iam       *iam.Client

	b *Provider
}

func (l *AwsClient) newEC2Client(region string) (aws.Config, error) {
	_session, err := config.LoadDefaultConfig(l.b.ctx,
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
		return _session, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			l.b.l.NewError(l.b.ctx, "Failed Init aws session", "Reason", err),
		)
	}
	l.b.l.Success(l.b.ctx, "AWS Session created successfully")
	return _session, nil
}

func (l *AwsClient) AuthorizeSecurityGroupIngress(ctx context.Context,
	parameter ec2.AuthorizeSecurityGroupIngressInput) error {

	_, err := l.ec2Client.AuthorizeSecurityGroupIngress(ctx, &parameter)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Authorizing Security Group Ingress", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error) {
	resp, err := l.ec2Client.DescribeImages(l.b.ctx, filter)
	if err != nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			l.b.l.NewError(l.b.ctx, "failed to describe images", "Reason", err),
		)
	}
	if len(resp.Images) == 0 {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			l.b.l.NewError(l.b.ctx, "no images found"),
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
			l.b.l.Debug(l.b.ctx, "ownerAlias", *i.ImageOwnerAlias)
		}
		l.b.l.Debug(l.b.ctx, "Printing amis", "creationdate", *i.CreationDate, "public", *i.Public, "ownerid", *i.OwnerId, "architecture", i.Architecture.Values(), "name", *i.Name, "imageid", *i.ImageId)
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

func (l *AwsClient) GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error) {
	azs, err := l.ec2Client.DescribeAvailabilityZones(l.b.ctx, &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "failed to describe availability zones", "Reason", err),
		)
	}

	return azs, nil
}

func (l *AwsClient) BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic, err := l.ec2Client.CreateNetworkInterface(ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Network Interface", "Reason", err),
		)
	}

	nicExistsWaiter := ec2.NewNetworkInterfaceAvailableWaiter(l.ec2Client, func(nicwaiter *ec2.NetworkInterfaceAvailableWaiterOptions) {
		nicwaiter.MinDelay = initialNicMinDelay
		nicwaiter.MaxDelay = initialNicMaxDelay
	})

	describeNICInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{*nic.NetworkInterface.NetworkInterfaceId},
	}

	err = nicExistsWaiter.Wait(l.b.ctx, describeNICInput, initialNicWaiterTime)

	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "Error Waiting for Network Interface", "Reason", err),
		)
	}

	return nic, err
}

func (l *AwsClient) BeginCreateSubNet(ctx context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := l.ec2Client.CreateSubnet(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Subnet", "Reason", err),
		)
	}

	subnetExistsWaiter := ec2.NewSubnetAvailableWaiter(l.ec2Client, func(subnetwaiter *ec2.SubnetAvailableWaiterOptions) {
		subnetwaiter.MinDelay = initialSubnetMinDelay
		subnetwaiter.MaxDelay = initialSubnetMaxDelay
	})

	describeSubnetInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{*subnet.Subnet.SubnetId},
	}

	err = subnetExistsWaiter.Wait(ctx, describeSubnetInput, initialSubnetWaiterTime)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "Error Waiting for Subnet", "Reason", err),
		)
	}

	return subnet, err
}

func (l *AwsClient) BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

	runResult, err := l.ec2Client.RunInstances(ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Instance", "Reason", err),
		)
	}

	return runResult, err
}

func (l *AwsClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := l.ec2Client.CreateInternetGateway(l.b.ctx, &gatewayparameter)
	if err != nil {
		return nil, nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Internet Gateway", "Reason", err),
		)
	}

	_, err = l.ec2Client.AttachInternetGateway(l.b.ctx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		return nil, nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Attaching Internet Gateway", "Reason", err),
		)
	}

	l.b.state.CloudInfra.Aws.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId

	routeTable, err := l.ec2Client.CreateRouteTable(l.b.ctx, &routeTableparameter)
	if err != nil {
		return nil, nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Route Table", "Reason", err),
		)
	}

	l.b.state.CloudInfra.Aws.RouteTableID = *routeTable.RouteTable.RouteTableId

	_, err = l.ec2Client.CreateRoute(l.b.ctx, &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(l.b.state.CloudInfra.Aws.GatewayID),
		RouteTableId:         aws.String(l.b.state.CloudInfra.Aws.RouteTableID),
	})
	if err != nil {
		return nil, nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Route", "Reason", err),
		)
	}

	for i := 0; i < len(l.b.state.CloudInfra.Aws.SubnetIDs); i++ {
		_, err = l.ec2Client.AssociateRouteTable(l.b.ctx, &ec2.AssociateRouteTableInput{
			RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
			SubnetId:     aws.String(l.b.state.CloudInfra.Aws.SubnetIDs[i]),
		})
		if err != nil {
			return nil, nil, ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				l.b.l.NewError(l.b.ctx, "Error assiciating route table", "Reason", err),
			)
		}
	}

	return routeTable, createInternetGateway, err
}

func (l *AwsClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := l.ec2Client.CreateVpc(l.b.ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating VPC", "Reason", err),
		)
	}

	vpcExistsWaiter := ec2.NewVpcExistsWaiter(l.ec2Client, func(vpcwaiter *ec2.VpcExistsWaiterOptions) {
		vpcwaiter.MinDelay = 1
		vpcwaiter.MaxDelay = 5
	})

	describeVpcInput := &ec2.DescribeVpcsInput{
		VpcIds: []string{*vpc.Vpc.VpcId},
	}

	err = vpcExistsWaiter.Wait(l.b.ctx, describeVpcInput, initialSubnetWaiterTime)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "Error Waiting for VPC", "Reason", err),
		)
	}

	return vpc, err
}

func (l *AwsClient) BeginDeleteVpc(ctx context.Context) error {

	_, err := l.ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(l.b.state.CloudInfra.Aws.VpcId),
	})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Deleting VPC", "Reason", err),
		)
	}

	return nil

}

func (l *AwsClient) BeginDeleteNIC(nicID string) error {
	initialWater := time.Now()
	// TODO(praful): use the helpers.Backoff
	//  also why do we wait for the nic to be available when it is deleting
	for {
		nic, err := l.ec2Client.DescribeNetworkInterfaces(l.b.ctx, &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{nicID},
		})
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				l.b.l.NewError(l.b.ctx, "Error Describing Network Interface", "Reason", err),
			)
		}
		if nic.NetworkInterfaces[0].Status == "available" {
			break
		}
		if time.Since(initialWater) > initialNicDeletionWaiterTime {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrTimeOut,
				l.b.l.NewError(l.b.ctx, "Error Waiting for Network Interface Timeout", "Reason", err),
			)
		}
	}
	_, err := l.ec2Client.DeleteNetworkInterface(l.b.ctx, &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(nicID),
	})
	if err != nil {
		l.b.l.Success(l.b.ctx, "skipped already deleted the nic", "id", nicID)
		return nil
	}

	return nil
}

func (l *AwsClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	_, err := l.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGrpID),
	})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "Error Deleting Security Group", "Reason", err),
		)
	}
	return nil
}

func (l *AwsClient) BeginDeleteSubNet(ctx context.Context, subnetID string) error {

	_, err := l.ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "Error Deleting Subnet", "Reason", err),
		)
	}

	return nil

}

func (l *AwsClient) BeginDeleteVM(instanceID string) error {

	_, err := l.ec2Client.TerminateInstances(l.b.ctx, &ec2.TerminateInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "failed to delete instance", "Reason", err),
		)
	}

	ec2TerminatedWaiter := ec2.NewInstanceTerminatedWaiter(l.ec2Client, func(itwo *ec2.InstanceTerminatedWaiterOptions) {
		itwo.MinDelay = initialInstanceMinDelay
		itwo.MaxDelay = initialInstanceMaxDelay
	})

	describeEc2Inp := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	err = ec2TerminatedWaiter.Wait(l.b.ctx, describeEc2Inp, instanceInitialTerminationTime)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "failed to wait for instance to terminate", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) InstanceInitialWaiter(ctx context.Context, instanceID string) error {

	instanceExistsWaiter := ec2.NewInstanceStatusOkWaiter(l.ec2Client, func(instancewaiter *ec2.InstanceStatusOkWaiterOptions) {
		instancewaiter.MinDelay = initialInstanceMinDelay
		instancewaiter.MaxDelay = initialInstanceMaxDelay
	})

	describeInstanceInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{instanceID},
	}

	err := instanceExistsWaiter.Wait(ctx, describeInstanceInput, instanceInitialWaiterTime)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			l.b.l.NewError(l.b.ctx, "Error Waiting for Instance", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

	naclresp, err := l.ec2Client.CreateNetworkAcl(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Network ACL", "Reason", err),
		)
	}

	_, err = l.ec2Client.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
		NetworkAclId: aws.String(*naclresp.NetworkAcl.NetworkAclId),
		RuleNumber:   aws.Int32(100),
		Protocol:     aws.String("-1"),
		RuleAction:   types.RuleActionAllow,
		CidrBlock:    aws.String("0.0.0.0/0"),
		Egress:       aws.Bool(true),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Network ACL Entry", "Reason", err),
		)
	}

	return naclresp, nil
}

func (l *AwsClient) BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup, err := l.ec2Client.CreateSecurityGroup(ctx, &parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Creating Security Group", "Reason", err),
		)
	}

	return securitygroup, nil
}

func (l *AwsClient) BeginDeleteVirtNet(ctx context.Context) error {

	if l.b.state.CloudInfra.Aws.RouteTableID == "" {
		l.b.l.Success(l.b.ctx, "skipped already deleted the route table")
	} else {
		_, err := l.ec2Client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: aws.String(l.b.state.CloudInfra.Aws.RouteTableID),
		})
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				l.b.l.NewError(l.b.ctx, "Error Deleting Route Table", "Reason", err),
			)
		}
		l.b.l.Success(l.b.ctx, "deleted the route table", "id", l.b.state.CloudInfra.Aws.RouteTableID)
		l.b.state.CloudInfra.Aws.RouteTableID = ""
		err = l.b.store.Write(l.b.state)
		if err != nil {
			return err
		}
	}

	if l.b.state.CloudInfra.Aws.GatewayID == "" {
		l.b.l.Success(l.b.ctx, "skipped already deleted the internet gateway")
	} else {
		_, err := l.ec2Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(l.b.state.CloudInfra.Aws.GatewayID),
			VpcId:             aws.String(l.b.state.CloudInfra.Aws.VpcId),
		})

		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				l.b.l.NewError(l.b.ctx, "Error Detaching Internet Gateway", "Reason", err),
			)
		}

		_, err = l.ec2Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(l.b.state.CloudInfra.Aws.GatewayID),
		})
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				l.b.l.NewError(l.b.ctx, "Error Deleting Internet Gateway", "Reason", err),
			)
		}
		id := l.b.state.CloudInfra.Aws.GatewayID
		l.b.state.CloudInfra.Aws.GatewayID = ""
		err = l.b.store.Write(l.b.state)
		if err != nil {
			return err
		}

		l.b.l.Success(l.b.ctx, "deleted the internet gateway", "id", id)

	}

	if l.b.state.CloudInfra.Aws.NetworkAclID == "" {
		// TODO(praful)!: resolve this empty branch
	} else {

		_, err := l.ec2Client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: aws.String(l.b.state.CloudInfra.Aws.NetworkAclID),
		})

		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				l.b.l.NewError(l.b.ctx, "Error Deleting Network ACL", "Reason", err),
			)
		}

		id := l.b.state.CloudInfra.Aws.NetworkAclID
		l.b.state.CloudInfra.Aws.NetworkAclID = ""
		err = l.b.store.Write(l.b.state)
		if err != nil {
			return err
		}
		l.b.l.Success(l.b.ctx, "deleted the network acl", "id", id)

	}
	return nil
}

func (l *AwsClient) DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error) {

	instanceipinput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}

	instanceinforesponse, err := l.ec2Client.DescribeInstances(ctx, instanceipinput)
	if err != nil {
		return instanceinforesponse, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Describing Instances", "Reason", err),
		)
	}

	return instanceinforesponse, nil
}

func (l *AwsClient) AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error {

	_, err := l.ec2Client.AuthorizeSecurityGroupEgress(ctx, &parameter)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Authorizing Security Group Egress", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) DeleteSSHKey(ctx context.Context, name string) error {

	_, err := l.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Deleting Key Pair", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) InitClient(b *Provider) error {
	l.b = b

	err := l.setRequiredENVVAR(l.b.ctx)
	if err != nil {
		return err
	}

	session, err := l.newEC2Client(l.b.Region)
	if err != nil {
		return err
	}

	l.config = &session
	l.stsClient = sts.NewFromConfig(session)
	l.ec2Client = ec2.NewFromConfig(session)
	l.eksClient = eks.NewFromConfig(session)
	l.iam = iam.NewFromConfig(session)
	return nil
}

func (l *AwsClient) ImportKeyPair(ctx context.Context, input *ec2.ImportKeyPairInput) error {

	if _, err := l.ec2Client.ImportKeyPair(ctx, input); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Importing Key Pair", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) setRequiredENVVAR(_ context.Context) error {

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

	l.b.l.Debug(l.b.ctx, msg)

	credentialsDocument, err := l.b.store.ReadCredentials(consts.CloudAws)
	if err != nil {
		return err
	}

	if credentialsDocument.Aws == nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNilCredentials,
			l.b.l.NewError(l.b.ctx, "no entry for aws present"),
		)
	}

	err = os.Setenv("AWS_ACCESS_KEY_ID", credentialsDocument.Aws.AccessKeyId)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			l.b.l.NewError(l.b.ctx, "failed to set environmenet variable", "Reason", err),
		)
	}
	err = os.Setenv("AWS_SECRET_ACCESS_KEY", credentialsDocument.Aws.SecretAccessKey)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			l.b.l.NewError(l.b.ctx, "failed to set environmenet variable", "Reason", err),
		)
	}
	return nil
}

func (l *AwsClient) ListLocations() ([]string, error) {

	parameter := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	result, err := l.ec2Client.DescribeRegions(l.b.ctx, parameter)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudRegion,
			l.b.l.NewError(l.b.ctx, "Error Describing Regions", "Reason", err),
		)
	}

	var validRegion []string
	for _, region := range result.Regions {
		validRegion = append(validRegion, *region.RegionName)
	}
	return validRegion, nil
}

func (l *AwsClient) ListVMTypes() (ec2.DescribeInstanceTypesOutput, error) {
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
		output, err := l.ec2Client.DescribeInstanceTypes(l.b.ctx, input)
		if err != nil {
			return vmTypes, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidCloudVMSize,
				l.b.l.NewError(l.b.ctx, "Error Describing Instance Types", "Reason", err),
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

func (l *AwsClient) ModifyVpcAttribute(ctx context.Context) error {

	modifyvpcinput := &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(l.b.state.CloudInfra.Aws.VpcId),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := l.ec2Client.ModifyVpcAttribute(l.b.ctx, modifyvpcinput)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Modifying VPC Attribute", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) ModifySubnetAttribute(ctx context.Context, i int) error {

	modifyusbnetinput := &ec2.ModifySubnetAttributeInput{
		SubnetId: aws.String(l.b.state.CloudInfra.Aws.SubnetIDs[i]),
		MapPublicIpOnLaunch: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := l.ec2Client.ModifySubnetAttribute(l.b.ctx, modifyusbnetinput)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Modifying Subnet Attribute", "Reason", err),
		)
	}

	return nil
}

func (l *AwsClient) BeginCreateEKS(ctx context.Context, paramter *eks.CreateClusterInput) (*eks.CreateClusterOutput, error) {

	resp, err := l.eksClient.CreateCluster(ctx, paramter)
	if err != nil {
		return nil, err
	}

	l.b.state.CloudInfra.Aws.ManagedClusterName = *resp.Cluster.Name
	l.b.state.CloudInfra.Aws.ManagedClusterArn = *resp.Cluster.Arn
	if err := l.b.store.Write(l.b.state); err != nil {
		return nil, err
	}

	waiter := eks.NewClusterActiveWaiter(l.eksClient)

	describeCluster := eks.DescribeClusterInput{
		Name: resp.Cluster.Name,
	}
	err = waiter.Wait(ctx, &describeCluster, managedClusterActiveWaiter)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (l *AwsClient) BeginCreateNodeGroup(ctx context.Context, paramter *eks.CreateNodegroupInput) (*eks.CreateNodegroupOutput, error) {
	resp, err := l.eksClient.CreateNodegroup(ctx, paramter)
	if err != nil {
		return nil, err
	}
	l.b.state.CloudInfra.Aws.ManagedNodeGroupName = *resp.Nodegroup.NodegroupName
	l.b.state.CloudInfra.Aws.ManagedNodeGroupArn = *resp.Nodegroup.NodegroupArn
	if err := l.b.store.Write(l.b.state); err != nil {
		return nil, err
	}

	waiter := eks.NewNodegroupActiveWaiter(l.eksClient)

	describeNodeGroup := &eks.DescribeNodegroupInput{
		NodegroupName: resp.Nodegroup.NodegroupName,
		ClusterName:   aws.String(l.b.state.CloudInfra.Aws.ManagedClusterName),
	}
	err = waiter.Wait(ctx, describeNodeGroup, managedNodeGroupActiveWaiter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *AwsClient) BeginDeleteNodeGroup(ctx context.Context, parameter *eks.DeleteNodegroupInput) (*eks.DeleteNodegroupOutput, error) {

	resp, err := l.eksClient.DeleteNodegroup(ctx, parameter)
	if err != nil {
		return nil, err
	}

	waiter := eks.NewNodegroupDeletedWaiter(l.eksClient, func(ndwo *eks.NodegroupDeletedWaiterOptions) {
		ndwo.MinDelay = 15 * time.Second
		ndwo.MaxDelay = 30 * time.Second
	})

	describeNodeGroup := eks.DescribeNodegroupInput{
		NodegroupName: aws.String(*resp.Nodegroup.NodegroupName),
		ClusterName:   aws.String(l.b.state.CloudInfra.Aws.ManagedClusterName),
	}

	err = waiter.Wait(ctx, &describeNodeGroup, managedNodeGroupDeletionWaiter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *AwsClient) BeginDeleteManagedCluster(ctx context.Context, parameter *eks.DeleteClusterInput) (*eks.DeleteClusterOutput, error) {

	resp, err := l.eksClient.DeleteCluster(ctx, parameter)
	if err != nil {
		return nil, err
	}

	waiter := eks.NewClusterDeletedWaiter(l.eksClient, func(cdwo *eks.ClusterDeletedWaiterOptions) {
		cdwo.MinDelay = 15 * time.Second
		cdwo.MaxDelay = 30 * time.Second
	})

	describeCluster := eks.DescribeClusterInput{
		Name: aws.String(*resp.Cluster.Name),
	}

	err = waiter.Wait(ctx, &describeCluster, managedClusterDeletionWaiter)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (l *AwsClient) DescribeCluster(ctx context.Context, parameter *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error) {
	resp, err := l.eksClient.DescribeCluster(ctx, parameter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *AwsClient) BeginCreateIAM(ctx context.Context, node string, parameter *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	createRoleResp, err := l.iam.CreateRole(ctx, parameter)
	if err != nil {
		return nil, err
	}

	switch node {
	case "controlplane":
		attachClusterPolicyInput := &iam.AttachRolePolicyInput{
			RoleName:  aws.String(*createRoleResp.Role.RoleName),
			PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"),
		}

		_, err = l.iam.AttachRolePolicy(ctx, attachClusterPolicyInput)
		if err != nil {
			return nil, err
		}
	case "worker":
		policyArn := []string{"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy", "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy", "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"}
		for _, policy := range policyArn {
			attachWorkerPolicyInput := &iam.AttachRolePolicyInput{
				RoleName:  aws.String(*createRoleResp.Role.RoleName),
				PolicyArn: aws.String(policy),
			}

			_, err = l.iam.AttachRolePolicy(ctx, attachWorkerPolicyInput)
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
func (s *STSTokenRetriever) GetToken(ctx context.Context, b *Provider, clusterName string, cfg aws.Config) (string, error) {
	out, err := s.PresignClient.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(opt *sts.PresignOptions) {
		k8sHeader := "x-k8s-aws-id"
		opt.Presigner = newCustomHTTPPresignerV4(opt.Presigner, map[string]string{
			k8sHeader:       clusterName,
			"X-Amz-Expires": "60",
		})
	})
	if err != nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrKubeconfigOperations,
			b.l.NewError(ctx,
				"unable to generate sst token for kubeconfig to auth with aws eks",
				"err", err),
		)
	}

	tokenPrefix := "k8s-aws-v1."
	token := tokenPrefix + base64.RawURLEncoding.EncodeToString([]byte(out.URL))
	return token, nil
}

func (l *AwsClient) GetKubeConfig(ctx context.Context, clusterName string) (string, error) {
	clusterData, err := l.eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})

	if err != nil {
		return "", err
	}

	preSignClient := sts.NewPresignClient(l.stsClient)
	tokenRetriver := NewSTSTokenRetriver(preSignClient)
	token, errToken := tokenRetriver.GetToken(context.Background(), l.b, clusterName, *l.config)
	if errToken != nil {
		return "", errToken
	}

	data := KubeConfigData{
		ClusterEndpoint:          *clusterData.Cluster.Endpoint,
		CertificateAuthorityData: *clusterData.Cluster.CertificateAuthority.Data,
		ClusterName:              *clusterData.Cluster.Name,
		Token:                    token,
	}

	tmpl, err := template.New("kubeconfig").Parse(kubeconfigTemplate)
	if err != nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			l.b.l.NewError(ctx, "failed to create kubeconfig templ", "err", err),
		)
	}

	var kubeconfig bytes.Buffer
	err = tmpl.Execute(&kubeconfig, data)
	if err != nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			l.b.l.NewError(ctx, "failed to generate kubeconfig", "err", err),
		)
	}

	return kubeconfig.String(), nil
}

func (l *AwsClient) BeginDeleteIAM(ctx context.Context, parameter *iam.DeleteRoleInput, node string) (*iam.DeleteRoleOutput, error) {

	switch node {
	case "controlplane":
		detachClusterPolicyInput := &iam.DetachRolePolicyInput{
			RoleName:  parameter.RoleName,
			PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"),
		}

		_, err := l.iam.DetachRolePolicy(ctx, detachClusterPolicyInput)
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

			_, err := l.iam.DetachRolePolicy(ctx, detachWorkerPolicyInput)
			if err != nil {
				return nil, err
			}

		}
	}

	resp, err := l.iam.DeleteRole(ctx, parameter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *AwsClient) CreateAddons(ctx context.Context, input *eks.CreateAddonInput) error {

	_, err := l.eksClient.CreateAddon(ctx, input)
	if err != nil {
		return l.b.l.NewError(l.b.ctx, "Error Creating Addon", "Reason", err)
	}

	return nil
}

func (l *AwsClient) DeleteAddons(ctx context.Context, input *eks.DeleteAddonInput) error {

	_, err := l.eksClient.DeleteAddon(ctx, input)
	if err != nil {
		return l.b.l.NewError(l.b.ctx, "Error Deleting Addon", "Reason", err)
	}

	return nil
}

func (l *AwsClient) ListK8sVersions(ctx context.Context) ([]string, error) {

	input := &eks.DescribeAddonVersionsInput{
		AddonName:         aws.String("vpc-cni"),
		KubernetesVersion: aws.String(""),
	}

	resp, err := l.eksClient.DescribeAddonVersions(ctx, input)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			l.b.l.NewError(l.b.ctx, "Error Describing Addon Versions", "Reason", err),
		)
	}

	versions := make(map[string]struct{})
	for _, addon := range resp.Addons {
		for _, addonVersion := range addon.AddonVersions {
			for _, k8sVersion := range addonVersion.Compatibilities {
				if k8sVersion.ClusterVersion != nil {
					versions[*k8sVersion.ClusterVersion] = struct{}{}
				}
			}
		}
	}

	var s []string
	for k := range versions {
		s = append(s, k)
	}
	return s, nil
}
