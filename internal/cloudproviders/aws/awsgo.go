package aws

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func ProvideClient() AwsGo {
	return &AwsGoClient{}
}

func ProvideMockClient() AwsGo {
	return &AwsGoMockClient{}
}

/* TODO figure out about pull until done funtions */

type AwsGo interface {
	//SetRegion(string)

	InitClient(storage resources.StorageFactory) error

	ListLocations(ec2client *ec2.Client) (*string, error)

	ListVMTypes() ([]string, error)

	BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error)

	BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error

	BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string, ec2client *ec2.Client) error

	CreateSSHKey() error

	DeleteSSHKey(ctx context.Context, ec2Client *ec2.Client, name string) error

	BeginCreateVM() error

	BeginDeleteVM(vmname string, ec2client *ec2.Client) error

	BeginCreatePubIP() error

	BeginDeletePubIP() error

	BeginCreateNIC() error

	BeginDeleteNIC(nicname string, ec2Client *ec2.Client) error

	BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error

	BeginDeleteSecurityGrp(ctx context.Context, ec2Client *ec2.Client, securityGrpID string) error
	BeginCreateSecurityGrp() error

	SetRegion(string)
	SetVpc(string)
}

type AwsGoClient struct {
	ACESSKEYID     string
	ACESSKEYSECRET string
	Region         string
	Vpc            string
}

// BeginCreateNIC implements AwsGo.
func (*AwsGoClient) BeginCreateNIC() error {
	panic("unimplemented")
}

// BeginCreatePubIP implements AwsGo.
func (*AwsGoClient) BeginCreatePubIP() error {
	panic("unimplemented")
}

// BeginCreateSecurityGrp implements AwsGo.
func (*AwsGoClient) BeginCreateSecurityGrp() error {
	panic("unimplemented")
}

// BeginCreateSubNet implements AwsGo.
func (awsclient *AwsGoClient) BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := ec2client.CreateSubnet(context, &parameter)
	if err != nil {
		log.Debug("Error Creating Subnet", "error", err)
	}

	_, err = ec2client.CreateTags(context, &ec2.CreateTagsInput{

		Resources: []string{*subnet.Subnet.SubnetId},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(subnetName),
			},
		},
	})

	return subnet, err
}

// BeginCreateVM implements AwsGo.
func (*AwsGoClient) BeginCreateVM() error {
	panic("unimplemented")
}

// BeginCreateVirtNet implements AwsGo.
func (*AwsGoClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := ec2client.CreateInternetGateway(context.TODO(), &gatewayparameter)
	if err != nil {
		log.Debug("Error Creating Internet Gateway", "error", err)
	}

	_, err = ec2client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		log.Debug("Error Attaching Internet Gateway", "error", err)
	}

	awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	////////////////////////////////////////
	routeTable, err := ec2client.CreateRouteTable(context.TODO(), &routeTableparameter)
	if err != nil {
		log.Debug("Error Creating Route Table", "error", err)
	}

	awsCloudState.RouteTableID = *routeTable.RouteTable.RouteTableId

	/*        create route		*/
	_, err = ec2client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(awsCloudState.GatewayID),
		RouteTableId:         aws.String(awsCloudState.RouteTableID),
	})
	if err != nil {
		log.Debug("Error Creating Route", "error", err)
	}

	return routeTable, createInternetGateway, err
}

// BeginCreateVpc implements AwsGo.
func (*AwsGoClient) BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := ec2client.CreateVpc(context.TODO(), &parameter)
	if err != nil {
		log.Debug("Error Creating VPC", "error", err)
	}

	return vpc, err
}

func (*AwsGoClient) BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	_, err := ec2client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(awsCloudState.VPCID),
	})
	if err != nil {
		log.Error("Error Deleting VPC", "error", err)
	}

	awsCloudState.VPCID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the vpc ", awsCloudState.VPCID)

	return nil

}

// BeginDeleteNIC implements AwsGo.
func (*AwsGoClient) BeginDeleteNIC(nicname string, ec2Client *ec2.Client) error {

	// check the nic is attached to a vm
	resp, err := ec2Client.DescribeNetworkInterfaces(context.Background(), &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{nicname},
	})
	if err != nil {
		log.Success("[skip] already deleted the nic", nicname)
		return nil
	}

	if err == nil {
		for _, nic := range resp.NetworkInterfaces {
			if nic.Attachment != nil {
				_, err = ec2Client.DetachNetworkInterface(context.Background(), &ec2.DetachNetworkInterfaceInput{
					AttachmentId: nic.Attachment.AttachmentId,
				})
				if err != nil {
					log.Success("[skip] already detached the nic", nicname)
				}
			}
		}
	}
	_, err = ec2Client.DeleteNetworkInterface(context.Background(), &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(nicname),
	})
	// search for the nic if its not found then its already deleted but it will return an error
	// so we ignore the error
	if err != nil {
		log.Success("[skip] already deleted the nic", nicname)
		return nil
	}

	log.Success("[aws] deleted the nic ", nicname)

	return nil
}

// BeginDeletePubIP implements AwsGo.
func (*AwsGoClient) BeginDeletePubIP() error {
	panic("unimplemented")
}

// BeginDeleteSecurityGrp implements AwsGo.
func (*AwsGoClient) BeginDeleteSecurityGrp(ctx context.Context, ec2Client *ec2.Client, securityGrpID string) error {

	_, err := ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGrpID),
	})
	if err != nil {
		log.Error("Error Deleting Security Group", "error", err)
	}

	log.Success("[aws] deleted the security group ", securityGrpID)
	return nil
}

// BeginDeleteSubNet implements AwsGo.
func (*AwsGoClient) BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string, ec2client *ec2.Client) error {

	if len(awsCloudState.SubnetID) == 0 {
		log.Debug("[skip] already deleted the subnet", awsCloudState.SubnetID)
		return nil
	}

	_, err := ec2client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return err
	}

	awsCloudState.SubnetID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the subnet ", awsCloudState.SubnetName)

	return nil

}

// BeginDeleteVM implements AwsGo.
func (*AwsGoClient) BeginDeleteVM(vmname string, ec2client *ec2.Client) error {

	_, err := ec2client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{vmname},
	})
	if err != nil {
		log.Error("Error Terminating Instance", "error", err)
	}

	// initial wait to terminate so nic can be detached and deleted
	time.Sleep(15 * time.Second)

	instanceinput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{vmname},
	}

	if err != nil {
		log.Error("Error Describing Instance", "error", err)
	}

	for {
		responce, err := ec2client.DescribeInstances(context.Background(), instanceinput)
		if err != nil {
			log.Error("Error Describing Instance", "error", err)
		}

		for _, reservation := range responce.Reservations {
			for _, instance := range reservation.Instances {
				if instance.State.Name == "terminated" {
					break
				}
			}
		}
		break
	}

	return nil
}

func (*AwsGoClient) BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	_, err := ec2client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(awsCloudState.RouteTableID),
	})
	if err != nil {
		log.Error("Error Deleting Route Table", "error", err)
	}

	_, err = ec2client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(awsCloudState.GatewayID),
		VpcId:             aws.String(awsCloudState.VPCID),
	})

	if err != nil {
		log.Error("Error Detaching Internet Gateway", "error", err)
	}

	_, err = ec2client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(awsCloudState.GatewayID),
	})

	if err != nil {
		log.Error("Error Deleting Internet Gateway", "error", err)
	}

	_, err = ec2client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
		NetworkAclId: aws.String(awsCloudState.NetworkAclID),
	})

	if err != nil {
		log.Error("Error Deleting Network ACL", "error", err)
	}

	awsCloudState.NetworkAclID = ""
	awsCloudState.RouteTableID = ""
	awsCloudState.GatewayID = ""

	log.Success("[aws] deleted the route table ", awsCloudState.RouteTableID)
	log.Success("[aws] deleted the network acl ", awsCloudState.NetworkAclID)
	log.Success("[aws] deleted the internet gateway ", awsCloudState.GatewayID)

	return nil
}

// CreateSSHKey implements AwsGo.
func (*AwsGoClient) CreateSSHKey() error {
	panic("unimplemented")
}

// DeleteSSHKey implements AwsGo.
func (*AwsGoClient) DeleteSSHKey(ctx context.Context, ec2Client *ec2.Client, name string) error {

	resp, err := ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		log.Error("Error Deleting Key Pair", "error", err)
	}

	log.Success("[aws] deleted the ssh key pair ", resp.KeyPairId)

	return nil
}

// InitClient implements AwsGo.
func (awsclient *AwsGoClient) InitClient(storage resources.StorageFactory) error {
	err := awsclient.setRequiredENVVAR(storage, context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (obj *AwsGoClient) setRequiredENVVAR(storage resources.StorageFactory, ctx context.Context) error {

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

	log.Debug(msg)

	return nil

}

// ListLocations implements AwsGo.
func (awsclient *AwsGoClient) ListLocations(ec2client *ec2.Client) (*string, error) {

	parameter := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	result, err := ec2client.DescribeRegions(context.Background(), parameter)
	if err != nil {
		fmt.Println("Error describing regions:", err)
		return nil, err
	}

	for _, region := range result.Regions {
		if *region.RegionName == awsclient.Region {
			fmt.Printf("Region: %s RegionEndpoint: %s\n", *region.RegionName, *region.Endpoint)
			return region.RegionName, nil
		}
	}

	return nil, fmt.Errorf("region not found")
}

// ListVMTypes implements AwsGo.
func (*AwsGoClient) ListVMTypes() ([]string, error) {

	//ec2Client := obj.ec2Client()
	//
	//result, err := ec2Client.DescribeInstanceTypes(context.Background(), *ec2.DescribeInstancesInput{})
	//if err != nil {
	//	fmt.Println(err)
	//	return nil, err
	//}
	//
	//var avalableinstance []string
	//for _, instancetype := range result.InstanceTypes {
	//	avalableinstance = append(avalableinstance, *instancetype.InstanceType)
	//}

	return nil, nil
}

// SetRegion implements AwsGo.
func (awsclient *AwsGoClient) SetRegion(region string) {
	awsclient.Region = region
	fmt.Println("region set to: ", awsclient.Region)
}

// SetVpc implements AwsGo.
func (awsclient *AwsGoClient) SetVpc(vpc string) {
	awsclient.Vpc = vpc
	fmt.Println("vpc set to: ", awsclient.Vpc)
}

type AwsGoMockClient struct {
	ACESSKEYID     string
	ACESSKEYSECRET string
	Region         string
	Vpc            string
}

// BeginCreateNIC implements AwsGo.
func (*AwsGoMockClient) BeginCreateNIC() error {
	panic("unimplemented")
}

// BeginCreatePubIP implements AwsGo.
func (*AwsGoMockClient) BeginCreatePubIP() error {
	panic("unimplemented")
}

// BeginCreateSecurityGrp implements AwsGo.
func (*AwsGoMockClient) BeginCreateSecurityGrp() error {
	panic("unimplemented")
}

// BeginCreateSubNet implements AwsGo.
func (*AwsGoMockClient) BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := ec2client.CreateSubnet(context, &parameter)
	if err != nil {
		log.Debug("Error Creating Subnet", "error", err)
	}

	_, err = ec2client.CreateTags(context, &ec2.CreateTagsInput{
		Resources: []string{*subnet.Subnet.SubnetId},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(subnetName),
			},
		},
	})

	return subnet, err
}

// BeginCreateVM implements AwsGo.
func (*AwsGoMockClient) BeginCreateVM() error {
	panic("unimplemented")
}

// BeginCreateVirtNet implements AwsGo.
func (*AwsGoMockClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := ec2client.CreateInternetGateway(context.TODO(), &gatewayparameter)
	if err != nil {
		log.Debug("Error Creating Internet Gateway", "error", err)
	}

	_, err = ec2client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		log.Debug("Error Attaching Internet Gateway", "error", err)
	}

	awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	routeTable, err := ec2client.CreateRouteTable(context.TODO(), &routeTableparameter)
	if err != nil {
		log.Debug("Error Creating Route Table", "error", err)
	}

	awsCloudState.RouteTableID = *routeTable.RouteTable.RouteTableId

	/*        create route		*/
	_, err = ec2client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(awsCloudState.GatewayID),
		RouteTableId:         aws.String(awsCloudState.RouteTableID),
	})
	if err != nil {
		log.Debug("Error Creating Route", "error", err)
	}

	return routeTable, createInternetGateway, err
}

// BeginCreateVpc implements AwsGo.
func (*AwsGoMockClient) BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := ec2client.CreateVpc(context.TODO(), &parameter)
	if err != nil {
		log.Debug("Error Creating VPC", "error", err)
	}

	fmt.Print("VPC Created Successfully: ")
	fmt.Println(*vpc.Vpc.VpcId)

	return vpc, err
}

func (*AwsGoMockClient) BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	_, err := ec2client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(awsCloudState.VPCID),
	})
	if err != nil {
		log.Error("Error Deleting VPC", "error", err)
	}

	awsCloudState.VPCID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the vpc ", awsCloudState.VPCID)

	return nil

}

// BeginDeleteNIC implements AwsGo.
func (*AwsGoMockClient) BeginDeleteNIC(nicname string, ec2Client *ec2.Client) error {
	panic("unimplemented")
}

// BeginDeletePubIP implements AwsGo.
func (*AwsGoMockClient) BeginDeletePubIP() error {
	panic("unimplemented")
}

// BeginDeleteSecurityGrp implements AwsGo.
func (*AwsGoMockClient) BeginDeleteSecurityGrp(ctx context.Context, ec2Client *ec2.Client, securityGrpID string) error {
	panic("unimplemented")
}

// BeginDeleteSubNet implements AwsGo.
func (*AwsGoMockClient) BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string, ec2client *ec2.Client) error {

	if len(awsCloudState.SubnetID) == 0 {
		log.Debug("[skip] already deleted the subnet", awsCloudState.SubnetID)
		return nil
	}

	_, err := ec2client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return err
	}

	awsCloudState.SubnetID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the subnet ", awsCloudState.SubnetName)

	return nil

}

// BeginDeleteVM implements AwsGo.
func (*AwsGoMockClient) BeginDeleteVM(vmname string, ec2client *ec2.Client) error {

	_, err := ec2client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{vmname},
	})
	if err != nil {
		log.Debug("Error Terminating Instance", "error", err)
	}

	responce, err := ec2client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{vmname},
	})
	if err != nil {
		log.Debug("Error Describing Instance", "error", err)
	}

	time.Sleep(20 * time.Second)

	for {
		for _, reservation := range responce.Reservations {
			for _, instance := range reservation.Instances {
				if instance.State.Name == "terminated" {
					log.Success("Instance Terminated", "instance", *instance.InstanceId)
				}
			}
		}
		break
	}

	return nil

}

// BeginDeleteVirtNet implements AwsGo.
func (*AwsGoMockClient) BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	_, err := ec2client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(awsCloudState.RouteTableID),
	})
	if err != nil {
		log.Error("Error Deleting Route Table", "error", err)
	}

	_, err = ec2client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(awsCloudState.GatewayID),
		VpcId:             aws.String(awsCloudState.VPCID),
	})

	if err != nil {
		log.Error("Error Detaching Internet Gateway", "error", err)
	}

	_, err = ec2client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(awsCloudState.GatewayID),
	})

	if err != nil {
		log.Error("Error Deleting Internet Gateway", "error", err)
	}

	_, err = ec2client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
		NetworkAclId: aws.String(awsCloudState.NetworkAclID),
	})

	if err != nil {
		log.Error("Error Deleting Network ACL", "error", err)
	}

	awsCloudState.NetworkAclID = ""
	awsCloudState.RouteTableID = ""
	awsCloudState.GatewayID = ""

	log.Success("[aws] deleted the route table ", awsCloudState.RouteTableID)
	log.Success("[aws] deleted the network acl ", awsCloudState.NetworkAclID)
	log.Success("[aws] deleted the internet gateway ", awsCloudState.GatewayID)

	return nil
}

// CreateSSHKey implements AwsGo.
func (*AwsGoMockClient) CreateSSHKey() error {
	panic("unimplemented")
}

// DeleteSSHKey implements AwsGo.
func (*AwsGoMockClient) DeleteSSHKey(ctx context.Context, ec2Client *ec2.Client, name string) error {

	return nil
}

// InitClient implements AwsGo.
func (*AwsGoMockClient) InitClient(storage resources.StorageFactory) error {
	panic("unimplemented")
}

// ListLocations implements AwsGo.
func (*AwsGoMockClient) ListLocations(ec2client *ec2.Client) (*string, error) {
	panic("unimplemented")
}

// ListVMTypes implements AwsGo.
func (*AwsGoMockClient) ListVMTypes() ([]string, error) {
	panic("unimplemented")
}

// SetRegion implements AwsGo.
func (*AwsGoMockClient) SetRegion(string) {
	panic("unimplemented")
}

// SetVpc implements AwsGo.
func (*AwsGoMockClient) SetVpc(string) {
	panic("unimplemented")
}
