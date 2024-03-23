package aws

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

var printcounter sync.Once

const (
	initialWait    = time.Second * 5
	waiterMinDelay = time.Second * 5
	waiterMaxDelay = time.Second * 10
)

func ProvideClient() AwsGo {
	return &AwsGoClient{}
}

func ProvideMockClient() AwsGo {
	return &AwsGoMockClient{}
}

type AwsGo interface {
	AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error

	InitClient(storage resources.StorageFactory) error

	ListLocations() (*string, error)

	ListVMTypes() (ec2.DescribeInstanceTypesOutput, error)

	BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error)

	BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory) error

	BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string) error

	CreateSSHKey() error

	DeleteSSHKey(ctx context.Context, name string) error

	BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)

	BeginDeleteVM(vmname string) error

	BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error)

	BeginDeleteNIC(nicID string) error

	BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory) error

	BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error)

	BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)

	BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error
	BeginCreateSecurityGrp() error

	DescribeInstanceState(ctx context.Context, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)

	GetLatestAMI(filter *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)

	AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error

	ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error

	ModifyVpcAttribute(ctx context.Context) error
	ModifySubnetAttribute(ctx context.Context) error

	setRequiredENV_VAR(storage resources.StorageFactory, ctx context.Context) error

	SetRegion(string) string
	SetVpc(string) string
}

type AwsGoClient struct {
	acessKeyID     string
	acessKeySecret string
	region         string
	vpc            string
	ec2Client      *ec2.Client
}

func newclient(region string) aws.Config {
	NewSession, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")),
	)
	if err != nil {
		log.Error(err.Error())
	}
	log.Success("AWS Session created successfully")

	return NewSession
}

func (awsclient *AwsGoClient) AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error {

	_, err := awsclient.ec2Client.AuthorizeSecurityGroupIngress(ctx, &parameter)
	if err != nil {
		log.Error("Error Authorizing Security Group Ingress", "error", err)
	}

	return nil
}

func (awsclient *AwsGoClient) GetLatestAMI(filter *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	resp, err := awsclient.ec2Client.DescribeImages(context.TODO(), filter)
	if err != nil {
		return resp, fmt.Errorf("failed to describe images: %w", err)
	}
	if len(resp.Images) == 0 {
		return resp, fmt.Errorf("no images found")
	}

	return resp, nil
}

func (awsclient *AwsGoClient) BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic, err := awsclient.ec2Client.CreateNetworkInterface(ctx, parameter)
	if err != nil {
		log.Error("Error Creating Network Interface", "error", err)
	}

	nicExistsWaiter := ec2.NewNetworkInterfaceAvailableWaiter(awsclient.ec2Client, func(nicwaiter *ec2.NetworkInterfaceAvailableWaiterOptions) {
		nicwaiter.MinDelay = waiterMinDelay
		nicwaiter.MaxDelay = waiterMaxDelay
	})

	describeNICInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{*nic.NetworkInterface.NetworkInterfaceId},
	}

	err = nicExistsWaiter.Wait(context.Background(), describeNICInput, 60*time.Second)

	if err != nil {
		log.Error("Error Waiting for Network Interface", "error", err)
	}

	return nic, err
}

func (awsclient *AwsGoClient) BeginCreateSecurityGrp() error {
	panic("unimplemented")
}

func (awsclient *AwsGoClient) BeginCreateSubNet(ctx context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet, err := awsclient.ec2Client.CreateSubnet(ctx, &parameter)
	if err != nil {
		log.Error("Error Creating Subnet", "error", err)
	}

	subnetExistsWaiter := ec2.NewSubnetAvailableWaiter(awsclient.ec2Client, func(subnetwaiter *ec2.SubnetAvailableWaiterOptions) {
		subnetwaiter.MinDelay = waiterMinDelay
		subnetwaiter.MaxDelay = waiterMaxDelay
	})

	describeSubnetInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{*subnet.Subnet.SubnetId},
	}

	err = subnetExistsWaiter.Wait(ctx, describeSubnetInput, 60*time.Second)
	if err != nil {
		log.Error("Error Waiting for Subnet", "error", err)
	}

	return subnet, err
}

func (awsclient *AwsGoClient) BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

	printcounter.Do(func() {
		time.Sleep(1 * time.Second)
		log.Note("Creating ec2 instances......")
	})

	runResult, err := awsclient.ec2Client.RunInstances(ctx, parameter)
	if err != nil {
		log.Error("Error Creating Instance", "error", err)
	}

	instanceExistsWaiter := ec2.NewInstanceStatusOkWaiter(awsclient.ec2Client, func(instancewaiter *ec2.InstanceStatusOkWaiterOptions) {
		instancewaiter.MinDelay = waiterMinDelay
		instancewaiter.MaxDelay = waiterMaxDelay
	})

	describeInstanceInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{*runResult.Instances[0].InstanceId},
	}

	err = instanceExistsWaiter.Wait(context.Background(), describeInstanceInput, 5*time.Minute)
	if err != nil {
		log.Error("Error Waiting for Instance", "error", err)
	}

	return runResult, err
}

func (awsclient *AwsGoClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	createInternetGateway, err := awsclient.ec2Client.CreateInternetGateway(context.TODO(), &gatewayparameter)
	if err != nil {
		log.Error("Error Creating Internet Gateway", "error", err)
	}

	_, err = awsclient.ec2Client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		log.Error("Error Attaching Internet Gateway", "error", err)
	}

	mainStateDocument.CloudInfra.Aws.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId

	routeTable, err := awsclient.ec2Client.CreateRouteTable(context.TODO(), &routeTableparameter)
	if err != nil {
		log.Error("Error Creating Route Table", "error", err)
	}

	mainStateDocument.CloudInfra.Aws.RouteTableID = *routeTable.RouteTable.RouteTableId

	/*        create route		*/
	_, err = awsclient.ec2Client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
		RouteTableId:         aws.String(mainStateDocument.CloudInfra.Aws.RouteTableID),
	})
	if err != nil {
		log.Error("Error Creating Route", "error", err)
	}

	_, err = awsclient.ec2Client.AssociateRouteTable(context.Background(), &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
		SubnetId:     aws.String(mainStateDocument.CloudInfra.Aws.SubnetID),
	})

	return routeTable, createInternetGateway, err
}

func (awsclient *AwsGoClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := awsclient.ec2Client.CreateVpc(context.TODO(), &parameter)
	if err != nil {
		log.Error("Error Creating VPC", "error", err)
	}

	vpcExistsWaiter := ec2.NewVpcExistsWaiter(awsclient.ec2Client, func(vpcwaiter *ec2.VpcExistsWaiterOptions) {
		vpcwaiter.MinDelay = 1
		vpcwaiter.MaxDelay = 5
	})

	describeVpcInput := &ec2.DescribeVpcsInput{
		VpcIds: []string{*vpc.Vpc.VpcId},
	}

	err = vpcExistsWaiter.Wait(context.Background(), describeVpcInput, initialWait)
	if err != nil {
		log.Error("Error Waiting for VPC", "error", err)
	}

	return vpc, err
}

func (obj *AwsGoClient) setRequiredENV_VAR(storage resources.StorageFactory, ctx context.Context) error {

	envAcessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	envKeySecret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if len(envAcessKeyID) != 0 && len(envKeySecret) != 0 {
		return nil
	}

	msg := "environment vars not set:"
	if len(envAcessKeyID) == 0 {
		msg = msg + " AWS_ACCESS_KEY_ID"
	}

	if len(envKeySecret) == 0 {
		msg = msg + " AWS_SECRET_ACCESS_KEY"
	}

	log.Warn(msg)

	credentials, err := storage.ReadCredentials(consts.CloudAws)
	if err != nil {
		return err
	}

	obj.acessKeyID = credentials.Aws.AcessKeyID

	err = os.Setenv("AWS_ACCESS_KEY_ID", obj.acessKeyID)
	if err != nil {
		return err
	}

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", credentials.Aws.AcessKeySecret)
	if err != nil {
		return err
	}

	return nil
}

func (awsclient *AwsGoClient) BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory) error {

	_, err := awsclient.ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
	})
	if err != nil {
		log.Error("Error Deleting VPC", "error", err)
	}

	return nil

}

func (awsclient *AwsGoClient) BeginDeleteNIC(nicID string) error {
	initialWater := time.Now()
	for {
		nic, err := awsclient.ec2Client.DescribeNetworkInterfaces(context.Background(), &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{nicID},
		})
		if err != nil {
			log.Error("Error Describing Network Interface", "error", err)
		}
		if nic.NetworkInterfaces[0].Status == "available" {
			break
		}
		if time.Since(initialWater) > 30*time.Second {
			log.NewError("Error Waiting for Network Interface Timeout", "error", err)
			return err
		}
	}
	_, err := awsclient.ec2Client.DeleteNetworkInterface(context.Background(), &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(nicID),
	})
	if err != nil {
		log.Success("[skip] already deleted the nic", nicID)
		return nil
	}

	return nil
}

func (awsclient *AwsGoClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	_, err := awsclient.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGrpID),
	})
	if err != nil {
		log.Error("Error Deleting Security Group", "error", err)
	}
	return nil
}

func (awsclient *AwsGoClient) BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string) error {

	_, err := awsclient.ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnetID),
	})
	if err != nil {
		return err
	}

	return nil

}

func (awsgo *AwsGoClient) BeginDeleteVM(instanceID string) error {

	_, err := awsgo.ec2Client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		log.NewError("failed to delete instance, %v", err)
		return err
	}

	ec2TerminatedWaiter := ec2.NewInstanceTerminatedWaiter(awsgo.ec2Client, func(itwo *ec2.InstanceTerminatedWaiterOptions) {
		itwo.MinDelay = waiterMinDelay
		itwo.MaxDelay = waiterMaxDelay
	})

	describeEc2Inp := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	err = ec2TerminatedWaiter.Wait(context.TODO(), describeEc2Inp, 300*time.Second)
	if err != nil {
		log.NewError("failed to wait for instance to terminate, %v", err)
		return err
	}

	return nil
}

func (awsclient *AwsGoClient) BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

	naclresp, err := awsclient.ec2Client.CreateNetworkAcl(ctx, &parameter)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return naclresp, nil
}

func (awsclient *AwsGoClient) BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup, err := awsclient.ec2Client.CreateSecurityGroup(ctx, &parameter)
	if err != nil {
		log.Error("Error Creating Security Group", "error", err)
	}

	return securitygroup, nil
}

func (awsclient *AwsGoClient) BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory) error {

	if mainStateDocument.CloudInfra.Aws.RouteTableID == "" {
		log.Success("[skip] already deleted the route table")
	} else {
		_, err := awsclient.ec2Client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: aws.String(mainStateDocument.CloudInfra.Aws.RouteTableID),
		})
		if err != nil {
			log.Error("Error Deleting Route Table", "error", err)
		}
		mainStateDocument.CloudInfra.Aws.RouteTableID = ""
		log.Success("[aws] deleted the route table ", mainStateDocument.CloudInfra.Aws.RouteTableID)
		err = storage.Write(mainStateDocument)
		if err != nil {
			log.Error("Error Writing State File", "error", err)
		}
	}

	if mainStateDocument.CloudInfra.Aws.GatewayID == "" {
		log.Success("[skip] already deleted the internet gateway")
	} else {
		_, err := awsclient.ec2Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
			VpcId:             aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
		})

		if err != nil {
			log.Error("Error Detaching Internet Gateway", "error", err)
		}

		_, err = awsclient.ec2Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(mainStateDocument.CloudInfra.Aws.GatewayID),
		})
		if err != nil {
			log.Error("Error Deleting Internet Gateway", "error", err)
		}
		mainStateDocument.CloudInfra.Aws.GatewayID = ""
		err = storage.Write(mainStateDocument)
		if err != nil {
			log.Error("Error Writing State File", "error", err)
		}

		log.Success("[aws] deleted the internet gateway ", mainStateDocument.CloudInfra.Aws.GatewayID)

	}

	if mainStateDocument.CloudInfra.Aws.NetworkAclID == "" {

	} else {

		_, err := awsclient.ec2Client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: aws.String(mainStateDocument.CloudInfra.Aws.NetworkAclID),
		})

		if err != nil {
			log.Error("Error Deleting Network ACL", "error", err)
		}

		mainStateDocument.CloudInfra.Aws.NetworkAclID = ""
		err = storage.Write(mainStateDocument)
		if err != nil {
			log.Error("Error Writing State File", "error", err)
		}
		log.Success("[aws] deleted the network acl ", mainStateDocument.CloudInfra.Aws.NetworkAclID)

	}
	return nil
}

func (awsclient *AwsGoClient) CreateSSHKey() error {
	panic("unimplemented")
}

func (awsclient *AwsGoClient) DescribeInstanceState(ctx context.Context, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

	instanceinforesponse, err := awsclient.ec2Client.DescribeInstances(ctx, instanceInput)
	if err != nil {
		return instanceinforesponse, err
	}

	return instanceinforesponse, nil
}

func (awsclient *AwsGoClient) AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error {

	_, err := awsclient.ec2Client.AuthorizeSecurityGroupEgress(ctx, &parameter)
	if err != nil {
		log.Error("Error Authorizing Security Group Egress", "error", err)
	}

	return nil
}

func (awsclient *AwsGoClient) DeleteSSHKey(ctx context.Context, name string) error {

	_, err := awsclient.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		log.Error("Error Deleting Key Pair", "error", err)
	}

	return nil
}

func (awsclient *AwsGoClient) InitClient(storage resources.StorageFactory) error {
	err := awsclient.setRequiredENV_VAR(storage, ctx)

	err = awsclient.setRequiredENVVAR(storage, context.Background())
	if err != nil {
		return err
	}

	awsclient.ec2Client = ec2.NewFromConfig(newclient(awsclient.region))
	return nil
}

func (awsclient *AwsGoClient) ImportKeyPair(ctx context.Context, input *ec2.ImportKeyPairInput) error {

	if _, err := awsclient.ec2Client.ImportKeyPair(ctx, input); err != nil {
		return err
	}

	return nil
}

func (awsclient *AwsGoClient) setRequiredENVVAR(storage resources.StorageFactory, ctx context.Context) error {

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

func (awsclient *AwsGoClient) ListLocations() (*string, error) {

	parameter := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	result, err := awsclient.ec2Client.DescribeRegions(context.Background(), parameter)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, region := range result.Regions {
		if *region.RegionName == awsclient.region {
			fmt.Printf("Region: %s RegionEndpoint: %s\n", *region.RegionName, *region.Endpoint)
			return region.RegionName, nil
		}
	}

	return nil, fmt.Errorf("region not found")
}

func (awsclient *AwsGoClient) ListVMTypes() (ec2.DescribeInstanceTypesOutput, error) {

	var vmTypes ec2.DescribeInstanceTypesOutput

	parameter, err := awsclient.ec2Client.DescribeInstanceTypes(context.Background(), &ec2.DescribeInstanceTypesInput{
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
func (awsclient *AwsGoClient) ModifyVpcAttribute(ctx context.Context) error {

	modifyvpcinput := &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err := awsclient.ec2Client.ModifyVpcAttribute(context.Background(), modifyvpcinput)
	if err != nil {
		fmt.Println(err)
		return err
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
	_, err := awsclient.ec2Client.ModifySubnetAttribute(context.Background(), modifyusbnetinput)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (awsclient *AwsGoClient) SetRegion(region string) string {
	awsclient.region = region
	fmt.Println("region set to: ", awsclient.region)

	return awsclient.region
}

func (awsclient *AwsGoClient) SetVpc(vpc string) string {
	awsclient.vpc = vpc
	fmt.Println("vpc set to: ", awsclient.vpc)

	return awsclient.vpc
}

type AwsGoMockClient struct {
	region string
	vpc    string
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
					Value: aws.String("test-nic"),
				},
			},
		},
	}

	return nic, nil

}

func (*AwsGoMockClient) BeginCreateSecurityGrp() error {
	return nil
}

func (*AwsGoMockClient) BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet := &ec2.CreateSubnetOutput{
		Subnet: &types.Subnet{
			SubnetId: aws.String("demo-ha-subnet"),
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

func (*AwsGoMockClient) setRequiredENV_VAR(storage resources.StorageFactory, ctx context.Context) error {
	return nil
}

func (*AwsGoMockClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

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

func (awsclient *AwsGoMockClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc := &ec2.CreateVpcOutput{
		Vpc: &types.Vpc{
			VpcId: aws.String("demo-ha-vpc"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("demo-ha-vpc"),
				},
			},
		},
	}
	return vpc, nil
}

func (*AwsGoMockClient) BeginDeleteVpc(ctx context.Context, storage resources.StorageFactory) error {

	mainStateDocument.CloudInfra.Aws.VpcId = ""

	if err := storage.Write(mainStateDocument); err != nil {
		log.Error("Error Writing State File", "error", err)
	}

	log.Success("[aws] deleted the vpc ", mainStateDocument.CloudInfra.Aws.VpcId)

	return nil

}

func (*AwsGoMockClient) BeginDeleteNIC(nicID string) error {

	return nil
}

func (*AwsGoMockClient) GetLatestAMI(filter *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	return nil, nil
}

func (*AwsGoMockClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	return nil
}

func (*AwsGoMockClient) BeginDeleteSubNet(ctx context.Context, storage resources.StorageFactory, subnetID string) error {

	mainStateDocument.CloudInfra.Aws.SubnetID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		log.Error("Error Writing State File", "error", err)
	}

	log.Success("[aws] deleted the subnet ", mainStateDocument.CloudInfra.Aws.SubnetName)

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

func (*AwsGoMockClient) BeginDeleteVirtNet(ctx context.Context, storage resources.StorageFactory) error {

	return nil
}

func (*AwsGoMockClient) CreateSSHKey() error {
	return nil
}

func (*AwsGoMockClient) DescribeInstanceState(ctx context.Context, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

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

func (*AwsGoMockClient) DeleteSSHKey(ctx context.Context, name string) error {

	return nil
}

func (*AwsGoMockClient) InitClient(storage resources.StorageFactory) error {
	return nil
}

func (*AwsGoMockClient) ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error {
	return nil
}

func (awsclient *AwsGoMockClient) ListLocations() (*string, error) {

	op := &ec2.DescribeRegionsOutput{
		Regions: []types.Region{
			{
				Endpoint:   aws.String("fake-endpoint"),
				RegionName: aws.String(awsclient.region),
			},
		},
	}

	region := op.Regions[0].RegionName

	return region, nil
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
func (*AwsGoMockClient) SetRegion(string) string {
	mainStateDocument.CloudInfra.Aws.Region = "fakeregion"
	return "fake-region"
}

func (*AwsGoMockClient) SetVpc(string) string {
	return "fake-vpc"
}
