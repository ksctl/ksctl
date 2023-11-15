package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

func (obj *AwsProvider) DelNetwork(storage resources.StorageFactory) error {

	_ = obj.metadata.resName
	obj.mxName.Unlock()

	if len(awsCloudState.VPCID) == 0 {
		log.Debug("[skip] already deleted the vpc", awsCloudState.VPCNAME)
		return nil
	}

	//ec2client := ec2.NewFromConfig(obj.session)

	err := obj.DeleteSubnet(context.Background(), storage, awsCloudState.SubnetID)
	if err != nil {
		return err
	}

	err = obj.client.BeginDeleteVirtNet(context.Background(), storage, obj.ec2Client())
	if err != nil {
		return err
	}

	err = obj.DeleteVpc(context.Background(), storage, awsCloudState.VPCID)
	if err != nil {
		return err
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the vpc ", awsCloudState.VPCNAME)

	return nil
}

func (obj *AwsProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(awsCloudState.SubnetID) == 0 {
		log.Debug("[skip] already deleted the subnet", awsCloudState.SubnetID)
		return nil
	}

	err := obj.client.BeginDeleteSubNet(ctx, storage, subnetName, obj.ec2Client())
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

func (obj *AwsProvider) DeleteVpc(ctx context.Context, storage resources.StorageFactory, resName string) error {

	if len(awsCloudState.VPCID) == 0 {
		log.Debug("[skip] already deleted the vpc", awsCloudState.VPCID)
		return nil
	}

	err := obj.client.BeginDeleteVpc(ctx, storage, obj.ec2Client())
	if err != nil {
		return err
	}

	awsCloudState.VPCID = ""
	awsCloudState.VPCNAME = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	return nil
}

func (obj *AwsProvider) NewNetwork(storage resources.StorageFactory) error {
	_ = obj.metadata.resName
	obj.mxName.Unlock()

	if len(awsCloudState.VPCID) != 0 {
		log.Debug("[skip] already created the vpc", awsCloudState.VPCNAME)
	}
	ec2client := ec2.NewFromConfig(obj.session)
	vpcclient := ec2.CreateVpcInput{
		// the subnet cidr block should be in the range of vpc cidr block
		CidrBlock: aws.String("172.31.0.0/16"),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("vpc"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.clusterName + "-vpc"),
					},
				},
			},
		},
	}

	vpc, err := obj.client.BeginCreateVpc(ec2client, vpcclient)
	if err != nil {
		return err
	}

	awsCloudState.VPCID = *vpc.Vpc.VpcId
	awsCloudState.VPCNAME = *vpc.Vpc.Tags[0].Value

	// now edit the vpc configuration

	// enable dns hostnames
	modifyvpcinput := &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(awsCloudState.VPCID),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err = ec2client.ModifyVpcAttribute(context.Background(), modifyvpcinput)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// if err := storage.Path(generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)).
	// 	Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
	// 	return err
	// }

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] created the vpc ", *vpc.Vpc.VpcId)

	ctx := context.TODO()

	if obj.haCluster {
		virtNet := obj.clusterName + "-vnet"
		subNet := obj.clusterName + "-subnet"

		if err := obj.CreateSubnet(ctx, storage, subNet); err != nil {
			return err
		}

		if err := obj.CreateVirtualNetwork(ctx, storage, virtNet); err != nil {
			return err
		}

	}
	return nil
}

func (obj *AwsProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(awsCloudState.SubnetID) != 0 {
		log.Debug("[skip] already created the subnet", awsCloudState.SubnetID)
	}

	client := obj.ec2Client()

	parameter := ec2.CreateSubnetInput{
		CidrBlock: aws.String("172.31.32.0/20"),
		VpcId:     aws.String(awsCloudState.VPCID),

		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("subnet"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.clusterName + "-subnet"),
					},
				},
			},
		},
		AvailabilityZone: aws.String("ap-south-1a"),

		// TODO: Add the following parameters
		// AvailabilityZoneId: aws.String(obj.AvailabilityZoneID),
	}
	response, err := obj.client.BeginCreateSubNet(ctx, subnetName, client, parameter)
	if err != nil {
		return err
	}

	awsCloudState.SubnetID = *response.Subnet.SubnetId
	awsCloudState.SubnetName = *response.Subnet.Tags[0].Value

	modifyusbnetinput := &ec2.ModifySubnetAttributeInput{
		SubnetId: aws.String(awsCloudState.SubnetID),
		MapPublicIpOnLaunch: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}
	_, err = client.ModifySubnetAttribute(ctx, modifyusbnetinput)
	if err != nil {
		return err
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	log.Success("[aws] created the subnet ", *response.Subnet.Tags[0].Value)

	naclinput := ec2.CreateNetworkAclInput{
		VpcId: aws.String(awsCloudState.VPCID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("network-acl"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.clusterName + "-nacl"),
					},
				},
			},
		},
	}

	naclresp, err := obj.ec2Client().CreateNetworkAcl(ctx, &naclinput)
	if err != nil {
		return err
	}
	NACLID = *naclresp.NetworkAcl.NetworkAclId
	log.Success("[aws] created the network acl ", *naclresp.NetworkAcl.NetworkAclId)

	_, err = obj.ec2Client().CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
		// ALLOW ALL TRAFFIC
		NetworkAclId: aws.String(NACLID),
		RuleNumber:   aws.Int32(100),
		Protocol:     aws.String("-1"),
		RuleAction:   types.RuleActionAllow,
		CidrBlock:    aws.String("0.0.0.0/0"),
		Egress:       aws.Bool(true),
	})
	if err != nil {
		return err
	}

	return nil
}

var NACLID string

func saveStateHelper(storage resources.StorageFactory) error {

	return nil
}

// Implements internetgateway, route table
func (obj *AwsProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, resName string) error {

	ec2Client := obj.ec2Client()

	internetGateway := ec2.CreateInternetGatewayInput{
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("internet-gateway"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.clusterName + "-ig"),
					},
				},
			},
		},
	}

	routeTableClient := ec2.CreateRouteTableInput{
		VpcId: aws.String(awsCloudState.VPCID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("route-table"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("DEMO" + "-rt"),
					},
				},
			},
		},
	}

	routeresponce, gatewayresp, err := obj.client.BeginCreateVirtNet(internetGateway, routeTableClient, ec2Client, awsCloudState.VPCID)
	if err != nil {
		return err
	}

	_, err = obj.ec2Client().AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(*routeresponce.RouteTable.RouteTableId),
		SubnetId:     aws.String(awsCloudState.SubnetID),
	})

	log.Success("[aws] created the internet gateway ", *gatewayresp.InternetGateway.InternetGatewayId)
	log.Success("[aws] created the route table ", *routeresponce.RouteTable.RouteTableId)

	return nil
}
