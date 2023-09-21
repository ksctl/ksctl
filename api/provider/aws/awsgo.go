package aws

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/api/resources"
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

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client , vpcid string) (string, error)

	BeginDeleteVirtNet() error

	BeginCreateSubNet(context context.Context, subnetName string, ec2client *ec2.Client, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteSubNet() error

	CreateSSHKey() error

	DeleteSSHKey() error

	BeginCreateVM() error

	BeginDeleteVM() error

	BeginCreatePubIP() error

	BeginDeletePubIP() error

	BeginCreateNIC() error

	BeginDeleteNIC() error

	BeginDeleteSecurityGrp() error
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
	
	return nil, nil
}

// BeginCreateVM implements AwsGo.
func (*AwsGoClient) BeginCreateVM() error {
	panic("unimplemented")
}

// BeginCreateVirtNet implements AwsGo.
func (*AwsGoClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client, vpcid string) (string, error) {

	createInternetGateway, err := ec2Client.CreateInternetGateway(context.TODO(), &gatewayparameter)
	if err != nil {
		log.Println(err)
	}

	_, err = ec2Client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(vpcid),
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println(*createInternetGateway.InternetGateway.InternetGatewayId)
	awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	fmt.Print("Internet Gateway Created Successfully: ")

	awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	
	routeTable, err := ec2Client.CreateRouteTable(context.TODO(), &routeTableparameter)
	if err != nil {
		log.Println(err)
	}

	fmt.Print("Route Table Created Successfully: ")
	fmt.Println(*routeTable.RouteTable.RouteTableId)
	RouteTableID = *routeTable.RouteTable.RouteTableId

	for _, subnet := range awsCloudState.SubnetID {
		_, err = ec2Client.AssociateRouteTable(context.TODO(), &ec2.AssociateRouteTableInput{
			RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
			SubnetId:     aws.String(subnet),
		})
		if err != nil {
			log.Println(err)
		}
	}
	fmt.Println("Route Table Associated Successfully....")

	/*        create route		*/
	_, err = ec2Client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(awsCloudState.GatewayID),
		RouteTableId:         aws.String(*routeTable.RouteTable.RouteTableId),
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Route Created Successfully....")

	return "", nil
}

// BeginCreateVpc implements AwsGo.
func (*AwsGoClient) BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	panic("unimplemented")
}

// BeginDeleteNIC implements AwsGo.
func (*AwsGoClient) BeginDeleteNIC() error {
	panic("unimplemented")
}

// BeginDeletePubIP implements AwsGo.
func (*AwsGoClient) BeginDeletePubIP() error {
	panic("unimplemented")
}

// BeginDeleteSecurityGrp implements AwsGo.
func (*AwsGoClient) BeginDeleteSecurityGrp() error {
	panic("unimplemented")
}

// BeginDeleteSubNet implements AwsGo.
func (*AwsGoClient) BeginDeleteSubNet() error {
	panic("unimplemented")
}

// BeginDeleteVM implements AwsGo.
func (*AwsGoClient) BeginDeleteVM() error {
	panic("unimplemented")
}

// BeginDeleteVirtNet implements AwsGo.
func (*AwsGoClient) BeginDeleteVirtNet() error {
	panic("unimplemented")
}

// CreateSSHKey implements AwsGo.
func (*AwsGoClient) CreateSSHKey() error {
	panic("unimplemented")
}

// DeleteSSHKey implements AwsGo.
func (*AwsGoClient) DeleteSSHKey() error {
	panic("unimplemented")
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

	storage.Logger().Warn(msg)

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
	panic("unimplemented")
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
		log.Println(err)
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
func (*AwsGoMockClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, ec2client *ec2.Client) (string, error) {

	return "", nil
}

// BeginCreateVpc implements AwsGo.
func (*AwsGoMockClient) BeginCreateVpc(ec2client *ec2.Client, parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc, err := ec2client.CreateVpc(context.TODO(), &parameter)
	if err != nil {
		fmt.Println("Error Creating VPC")
		log.Println(err)
	}
	_ , err := ec2client.

	VPCID = *vpc.Vpc.VpcId
	fmt.Print("VPC Created Successfully: ")
	fmt.Println(*vpc.Vpc.VpcId)

	return vpc, err
}

// BeginDeleteNIC implements AwsGo.
func (*AwsGoMockClient) BeginDeleteNIC() error {
	panic("unimplemented")
}

// BeginDeletePubIP implements AwsGo.
func (*AwsGoMockClient) BeginDeletePubIP() error {
	panic("unimplemented")
}

// BeginDeleteSecurityGrp implements AwsGo.
func (*AwsGoMockClient) BeginDeleteSecurityGrp() error {
	panic("unimplemented")
}

// BeginDeleteSubNet implements AwsGo.
func (*AwsGoMockClient) BeginDeleteSubNet() error {
	panic("unimplemented")
}

// BeginDeleteVM implements AwsGo.
func (*AwsGoMockClient) BeginDeleteVM() error {
	panic("unimplemented")
}

// BeginDeleteVirtNet implements AwsGo.
func (*AwsGoMockClient) BeginDeleteVirtNet() error {
	panic("unimplemented")
}

// CreateSSHKey implements AwsGo.
func (*AwsGoMockClient) CreateSSHKey() error {
	panic("unimplemented")
}

// DeleteSSHKey implements AwsGo.
func (*AwsGoMockClient) DeleteSSHKey() error {
	panic("unimplemented")
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
