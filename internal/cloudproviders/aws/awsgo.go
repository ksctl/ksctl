package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"os"
)

func ProvideClient() AwsGo {
	return &AwsGoClient{}
}

func ProvideMockClient() AwsGo {
	return &AwsGoMockClient{}
}

type AwsGo interface {
	AuthorizeSecurityGroupIngress(ctx context.Context, ec2Client *ec2.Client, parameter ec2.AuthorizeSecurityGroupIngressInput) error

	InitClient(storage resources.StorageFactory) error

	ListLocations(ec2client *ec2.Client) (*string, error)

	ListVMTypes(ec2Client *ec2.Client) (ec2.DescribeInstanceTypesOutput, error)

	BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error)

	BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error

	BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string, ec2client *ec2.Client) error

	CreateSSHKey() error

	DeleteSSHKey(ctx context.Context, ec2Client *ec2.Client, name string) error

	BeginCreateVM(ctx context.Context, ec2client *ec2.Client, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)

	BeginDeleteVM(vmname string, ec2client *ec2.Client) error

	BeginCreatePubIP() error

	BeginDeletePubIP() error

	BeginCreateNIC(ctx context.Context, ec2client *ec2.Client, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error)

	BeginDeleteNIC(nicname string, ec2Client *ec2.Client) error

	BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error

	BeginCreateNetworkAcl(ctx context.Context, ec2client *ec2.Client, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error)

	BeginCreateSecurityGroup(ctx context.Context, ec2Client *ec2.Client, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)

	BeginDeleteSecurityGrp(ctx context.Context, ec2Client *ec2.Client, securityGrpID string) error
	BeginCreateSecurityGrp() error

	DescribeInstanceState(ctx context.Context, ec2Client *ec2.Client, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)

	AuthorizeSecurityGroupEgress(ctx context.Context, ec2Client *ec2.Client, parameter ec2.AuthorizeSecurityGroupEgressInput) error

	ImportKeyPair(ctx context.Context, ec2Client *ec2.Client, keypair ec2.ImportKeyPairInput) error

	ModifyVpcAttribute(ctx context.Context, ec2client *ec2.Client) error
	ModifySubnetAttribute(ctx context.Context, ec2client *ec2.Client) error

	SetRegion(string) string
	SetVpc(string) string
}

type AwsGoClient struct {
	ACESSKEYID     string
	ACESSKEYSECRET string
	Region         string
	Vpc            string
}

func (*AwsGoClient) AuthorizeSecurityGroupIngress(ctx context.Context, ec2Client *ec2.Client, parameter ec2.AuthorizeSecurityGroupIngressInput) error {

	_, err := ec2Client.AuthorizeSecurityGroupIngress(ctx, &parameter)
	if err != nil {
		log.Error("Error Authorizing Security Group Ingress", "error", err)
	}

	return nil
}

func (*AwsGoClient) BeginCreateNIC(ctx context.Context, ec2client *ec2.Client, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic, err := ec2client.CreateNetworkInterface(ctx, parameter)
	if err != nil {
		log.Error("Error Creating Network Interface", "error", err)
	}

	return nic, err
}

func (*AwsGoClient) BeginCreatePubIP() error {
	panic("unimplemented")
}

func (*AwsGoClient) BeginCreateSecurityGrp() error {
	panic("unimplemented")
}

func (awsclient *AwsGoClient) BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := ec2client.CreateSubnet(context, &parameter)
	if err != nil {
		log.Error("Error Creating Subnet", "error", err)
	}

	return subnet, err
}

func (*AwsGoClient) BeginCreateVM(ctx context.Context, ec2client *ec2.Client, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

	runResult, err := ec2client.RunInstances(ctx, parameter)
	if err != nil {
		log.Error("Error Creating Instance", "error", err)
	}

	// TODO Wait until the instance is running.
	for {
		instanceinfo := &ec2.DescribeInstancesInput{
			InstanceIds: []string{*runResult.Instances[0].InstanceId},
		}

		instanceinforesponse, err := ec2client.DescribeInstances(context.Background(), instanceinfo)
		if err != nil {
			return nil, err
		}
		if instanceinforesponse.Reservations[0].Instances[0].State.Name == types.InstanceStateNameRunning {
			break
		}
	}

	return runResult, err
}

func (*AwsGoClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := ec2client.CreateInternetGateway(context.TODO(), &gatewayparameter)
	if err != nil {
		log.Error("Error Creating Internet Gateway", "error", err)
	}

	_, err = ec2client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		log.Error("Error Attaching Internet Gateway", "error", err)
	}

	awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId

	routeTable, err := ec2client.CreateRouteTable(context.TODO(), &routeTableparameter)
	if err != nil {
		log.Error("Error Creating Route Table", "error", err)
	}

	awsCloudState.RouteTableID = *routeTable.RouteTable.RouteTableId

	/*        create route		*/
	_, err = ec2client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(awsCloudState.GatewayID),
		RouteTableId:         aws.String(awsCloudState.RouteTableID),
	})
	if err != nil {
		log.Error("Error Creating Route", "error", err)
	}

	_, err = ec2client.AssociateRouteTable(context.Background(), &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
		SubnetId:     aws.String(awsCloudState.SubnetID),
	})

	return routeTable, createInternetGateway, err
}

func (*AwsGoClient) BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := ec2client.CreateVpc(context.TODO(), &parameter)
	if err != nil {
		log.Error("Error Creating VPC", "error", err)
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

	return nil

}

func (*AwsGoClient) BeginDeleteNIC(nicname string, ec2Client *ec2.Client) error {
	for {
		// describe nic
		nic, err := ec2Client.DescribeNetworkInterfaces(context.Background(), &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{nicname},
		})
		if err != nil {
			log.Error("Error Describing Network Interface", "error", err)
		}
		if nic.NetworkInterfaces[0].Status == "available" {
			break
		}
	}
	_, err := ec2Client.DeleteNetworkInterface(context.Background(), &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(nicname),
	})
	if err != nil {
		log.Success("[skip] already deleted the nic", nicname)
		return nil
	}

	return nil
}

func (*AwsGoClient) BeginDeletePubIP() error {
	panic("unimplemented")
}

func (*AwsGoClient) BeginDeleteSecurityGrp(ctx context.Context, ec2Client *ec2.Client, securityGrpID string) error {

	_, err := ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGrpID),
	})
	if err != nil {
		log.Error("Error Deleting Security Group", "error", err)
	}
	return nil
}

func (*AwsGoClient) BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string, ec2client *ec2.Client) error {

	_, err := ec2client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return err
	}

	return nil

}

func (*AwsGoClient) BeginDeleteVM(vmname string, ec2client *ec2.Client) error {

	_, err := ec2client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{vmname},
	})
	if err != nil {
		log.Error("Error Terminating Instance", "error", err)
	}

	for {
		resp, err := ec2client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{vmname},
		})
		if err != nil {
			log.Error("Error Describing Instance", "error", err)
		}
		if resp.Reservations[0].Instances[0].State.Name == types.InstanceStateNameTerminated {
			break
		}
	}

	// initial wait to terminate so nic can be detached and deleted
	return nil
}

func (*AwsGoClient) BeginCreateNetworkAcl(ctx context.Context, ec2client *ec2.Client, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

	naclresp, err := ec2client.CreateNetworkAcl(ctx, &parameter)
	if err != nil {
		return nil, err
	}

	_, err = ec2client.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
		// ALLOW ALL TRAFFIC
		NetworkAclId: aws.String(*naclresp.NetworkAcl.NetworkAclId),
		RuleNumber:   aws.Int32(100),
		Protocol:     aws.String("-1"),
		RuleAction:   types.RuleActionAllow,
		CidrBlock:    aws.String("0.0.0.0/0"),
		Egress:       aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return naclresp, nil
}

func (*AwsGoClient) BeginCreateSecurityGroup(ctx context.Context, ec2Client *ec2.Client, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup, err := ec2Client.CreateSecurityGroup(ctx, &parameter)
	if err != nil {
		log.Error("Error Creating Security Group", "error", err)
	}

	return securitygroup, nil
}

func (*AwsGoClient) BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	if awsCloudState.RouteTableID == "" {
		log.Success("[skip] already deleted the route table")
	} else {
		_, err := ec2client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: aws.String(awsCloudState.RouteTableID),
		})
		if err != nil {
			log.Error("Error Deleting Route Table", "error", err)
		}
		awsCloudState.RouteTableID = ""
		log.Success("[aws] deleted the route table ", awsCloudState.RouteTableID)
		err = saveStateHelper(storage)
		if err != nil {
			log.Error("Error Saving State", "error", err)
		}
	}

	if awsCloudState.GatewayID == "" {
		log.Success("[skip] already deleted the internet gateway")
	} else {
		_, err := ec2client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
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
		awsCloudState.GatewayID = ""
		err = saveStateHelper(storage)

		log.Success("[aws] deleted the internet gateway ", awsCloudState.GatewayID)

	}

	if awsCloudState.NetworkAclID == "" {

	} else {

		_, err := ec2client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: aws.String(awsCloudState.NetworkAclID),
		})

		if err != nil {
			log.Error("Error Deleting Network ACL", "error", err)
		}

		awsCloudState.NetworkAclID = ""
		err = saveStateHelper(storage)
		log.Success("[aws] deleted the network acl ", awsCloudState.NetworkAclID)

	}
	return nil
}

func (*AwsGoClient) CreateSSHKey() error {
	panic("unimplemented")
}

func (*AwsGoClient) DescribeInstanceState(ctx context.Context, ec2Client *ec2.Client, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

	instanceinforesponse, err := ec2Client.DescribeInstances(ctx, instanceInput)
	if err != nil {
		return instanceinforesponse, err
	}

	return instanceinforesponse, nil
}

func (*AwsGoClient) AuthorizeSecurityGroupEgress(ctx context.Context, ec2Client *ec2.Client, parameter ec2.AuthorizeSecurityGroupEgressInput) error {

	_, err := ec2Client.AuthorizeSecurityGroupEgress(ctx, &parameter)
	if err != nil {
		log.Error("Error Authorizing Security Group Egress", "error", err)
	}

	return nil
}

func (*AwsGoClient) DeleteSSHKey(ctx context.Context, ec2Client *ec2.Client, name string) error {

	_, err := ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		log.Error("Error Deleting Key Pair", "error", err)
	}

	return nil
}

func (awsclient *AwsGoClient) InitClient(storage resources.StorageFactory) error {
	err := awsclient.setRequiredENVVAR(storage, context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (AwsGoClient *AwsGoClient) ImportKeyPair(ctx context.Context, ec2client *ec2.Client, input ec2.ImportKeyPairInput) error {

	if _, err := ec2client.ImportKeyPair(ctx, &input); err != nil {
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

	return nil

}

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

func (*AwsGoClient) ListVMTypes(ec2Client *ec2.Client) (ec2.DescribeInstanceTypesOutput, error) {

	var vmTypes ec2.DescribeInstanceTypesOutput

	parameter, err := ec2Client.DescribeInstanceTypes(context.Background(), &ec2.DescribeInstanceTypesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("current-generation"),
				Values: []string{"true"},
			},
		},
		InstanceTypes: []types.InstanceType{"t2.micro"},
	})
	if err != nil {
		log.Error("Error Describing Instance Types", "error", err)
		return vmTypes, err
	}

	for _, instanceType := range parameter.InstanceTypes {
		vmTypes.InstanceTypes = append(vmTypes.InstanceTypes, instanceType)
	}

	return vmTypes, nil

}
func (awsclient *AwsGoClient) ModifyVpcAttribute(ctx context.Context, ec2client *ec2.Client) error {

	modifyvpcinput := &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(awsCloudState.VPCID),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := ec2client.ModifyVpcAttribute(context.Background(), modifyvpcinput)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (awsclient *AwsGoClient) ModifySubnetAttribute(ctx context.Context, ec2client *ec2.Client) error {

	modifyusbnetinput := &ec2.ModifySubnetAttributeInput{
		SubnetId: aws.String(awsCloudState.SubnetID),
		MapPublicIpOnLaunch: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := ec2client.ModifySubnetAttribute(context.Background(), modifyusbnetinput)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (awsclient *AwsGoClient) SetRegion(region string) string {
	awsclient.Region = region
	fmt.Println("region set to: ", awsclient.Region)

	return awsclient.Region
}

func (awsclient *AwsGoClient) SetVpc(vpc string) string {
	awsclient.Vpc = vpc
	fmt.Println("vpc set to: ", awsclient.Vpc)

	return awsclient.Vpc
}

type AwsGoMockClient struct {
	ACESSKEYID     string
	ACESSKEYSECRET string
	Region         string
	Vpc            string
}

func (*AwsGoMockClient) AuthorizeSecurityGroupIngress(ctx context.Context, ec2Client *ec2.Client, parameter ec2.AuthorizeSecurityGroupIngressInput) error {
	return nil
}

func (*AwsGoMockClient) AuthorizeSecurityGroupEgress(ctx context.Context, ec2Client *ec2.Client, parameter ec2.AuthorizeSecurityGroupEgressInput) error {
	return nil
}

func (*AwsGoMockClient) BeginCreateNIC(ctx context.Context, ec2client *ec2.Client, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic := &ec2.CreateNetworkInterfaceOutput{
		NetworkInterface: &types.NetworkInterface{
			NetworkInterfaceId: aws.String("test-nic-1234567890"),
			TagSet: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-nic"),
				},
			},
		},
	}

	return nic, nil

}

func (*AwsGoMockClient) BeginCreatePubIP() error {
	return nil
}

func (*AwsGoMockClient) BeginCreateSecurityGrp() error {
	return nil
}

func (*AwsGoMockClient) BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet := &ec2.CreateSubnetOutput{
		Subnet: &types.Subnet{
			SubnetId: aws.String("test-subnet-1234567890"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-subnet"),
				},
			},
		},
	}

	return subnet, nil
}

func (*AwsGoMockClient) BeginCreateVM(ctx context.Context, ec2client *ec2.Client, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

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

func (*AwsGoMockClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	routeTable := &ec2.CreateRouteTableOutput{
		RouteTable: &types.RouteTable{
			RouteTableId: aws.String("test-route-table-1234567890"),
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
			InternetGatewayId: aws.String("test-internet-gateway-1234567890"),
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

func (*AwsGoMockClient) BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc := &ec2.CreateVpcOutput{
		Vpc: &types.Vpc{
			VpcId: aws.String("test-vpc-1234567890"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test-vpc"),
				},
			},
		},
	}
	return vpc, nil
}

func (*AwsGoMockClient) BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	awsCloudState.VPCID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the vpc ", awsCloudState.VPCID)

	return nil

}

func (*AwsGoMockClient) BeginDeleteNIC(nicname string, ec2Client *ec2.Client) error {

	return nil
}

func (*AwsGoMockClient) BeginDeletePubIP() error {

	return nil
}

func (*AwsGoMockClient) BeginDeleteSecurityGrp(ctx context.Context, ec2Client *ec2.Client, securityGrpID string) error {

	return nil
}

func (*AwsGoMockClient) BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string, ec2client *ec2.Client) error {

	awsCloudState.SubnetID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the subnet ", awsCloudState.SubnetName)

	return nil

}

func (*AwsGoMockClient) BeginCreateNetworkAcl(ctx context.Context, ec2client *ec2.Client, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

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

func (*AwsGoMockClient) BeginCreateSecurityGroup(ctx context.Context, ec2Client *ec2.Client, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup := &ec2.CreateSecurityGroupOutput{
		GroupId: aws.String("test-security-group-1234567890"),
	}

	return securitygroup, nil
}

func (*AwsGoMockClient) BeginDeleteVM(vmname string, ec2client *ec2.Client) error {
	return nil
}

func (*AwsGoMockClient) BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory, ec2client *ec2.Client) error {

	return nil
}

func (*AwsGoMockClient) CreateSSHKey() error {
	return nil
}

func (*AwsGoMockClient) DescribeInstanceState(ctx context.Context, ec2Client *ec2.Client, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

	instanceinforesponse := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
						PublicIpAddress:  aws.String("fake-ip"),
						PrivateIpAddress: aws.String("fake-ip"),
					},
				},
			},
		},
	}

	return instanceinforesponse, nil
}

func (*AwsGoMockClient) DeleteSSHKey(ctx context.Context, ec2Client *ec2.Client, name string) error {

	return nil
}

func (*AwsGoMockClient) InitClient(storage resources.StorageFactory) error {
	return nil
}

func (*AwsGoMockClient) ImportKeyPair(ctx context.Context, ec2Client *ec2.Client, keypair ec2.ImportKeyPairInput) error {
	return nil
}

func (*AwsGoMockClient) ListLocations(ec2client *ec2.Client) (*string, error) {

	op := &ec2.DescribeRegionsOutput{
		Regions: []types.Region{
			{
				Endpoint:   aws.String("fake-endpoint"),
				RegionName: aws.String("fake-region"),
			},
		},
	}

	region := op.Regions[0].RegionName

	return region, nil
}

func (*AwsGoMockClient) ListVMTypes(ec2Client *ec2.Client) (ec2.DescribeInstanceTypesOutput, error) {
	return ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []types.InstanceTypeInfo{
			{
				InstanceType: "fake",
			},
		},
	}, nil
}

func (*AwsGoMockClient) ModifyVpcAttribute(ctx context.Context, ec2client *ec2.Client) error {
	return nil
}

func (*AwsGoMockClient) ModifySubnetAttribute(ctx context.Context, ec2client *ec2.Client) error {
	return nil
}
func (*AwsGoMockClient) SetRegion(string) string {
	awsCloudState.Region = "fakeregion"
	return "fake-region"
}

func (*AwsGoMockClient) SetVpc(string) string {
	return "fake-vpc"
}
