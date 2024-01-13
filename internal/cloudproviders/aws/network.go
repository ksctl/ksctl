package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ksctl/ksctl/pkg/resources"
)

func (obj *AwsProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Aws.SubnetID) == 0 {
		log.Print("[skip] already deleted the vpc", mainStateDocument.CloudInfra.Aws.VPCNAME)
	} else {
		err := obj.DeleteSubnet(context.Background(), storage, mainStateDocument.CloudInfra.Aws.SubnetID)
		if err != nil {
			return err
		}
	}

	err := obj.client.BeginDeleteVirtNet(context.Background(), storage, obj.ec2Client())
	if err != nil {
		return err
	}

	if mainStateDocument.CloudInfra.Aws.VPCID == "" {
		log.Success("[aws] deleted the vpc ", "id: ", mainStateDocument.CloudInfra.Aws.VPCNAME)
	} else {
		err = obj.DeleteVpc(context.Background(), storage, mainStateDocument.CloudInfra.Aws.VPCID)
		if err != nil {
			return err
		}
	}

	if err := storage.DeleteCluster(); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("[aws] deleted the vpc ", "id: ", mainStateDocument.CloudInfra.Aws.VPCNAME)

	return nil
}

func (obj *AwsProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	err := obj.client.BeginDeleteSubNet(ctx, storage, subnetName, obj.ec2Client())
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Aws.SubnetID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("[aws] deleted the subnet ", "id: ", mainStateDocument.CloudInfra.Aws.SubnetName)

	return nil
}

func (obj *AwsProvider) DeleteVpc(ctx context.Context, storage resources.StorageFactory, resName string) error {

	err := obj.client.BeginDeleteVpc(ctx, storage, obj.ec2Client())
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Aws.VPCID = ""
	mainStateDocument.CloudInfra.Aws.VPCNAME = ""
	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("[aws] deleted the vpc ", "id", mainStateDocument.CloudInfra.Aws.VPCNAME)
	return nil
}

func (obj *AwsProvider) NewNetwork(storage resources.StorageFactory) error {
	_ = <-obj.chResName

	if len(mainStateDocument.CloudInfra.Aws.VPCID) != 0 {
		log.Print("[skip] already created the vpc", mainStateDocument.CloudInfra.Aws.VPCNAME)
	} else {
		ec2client := obj.ec2Client()
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

		log.Debug("Printing", "virtualprivatecloud", vpcclient)

		vpc, err := obj.client.BeginCreateVpc(ec2client, vpcclient)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.VPCID = *vpc.Vpc.VpcId
		mainStateDocument.CloudInfra.Aws.VPCNAME = *vpc.Vpc.Tags[0].Value

		if err := obj.client.ModifyVpcAttribute(context.Background(), ec2client); err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("[aws] created the vpc ", "id: ", *vpc.Vpc.VpcId)

	}

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

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	return nil
}

func (obj *AwsProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(mainStateDocument.CloudInfra.Aws.SubnetID) != 0 {
		log.Print("[skip] already created the subnet", mainStateDocument.CloudInfra.Aws.SubnetID)
	} else {

		client := obj.ec2Client()

		parameter := ec2.CreateSubnetInput{
			CidrBlock: aws.String("172.31.32.0/20"),
			VpcId:     aws.String(mainStateDocument.CloudInfra.Aws.VPCID),

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
		}
		response, err := obj.client.BeginCreateSubNet(ctx, subnetName, client, parameter)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.SubnetID = *response.Subnet.SubnetId
		mainStateDocument.CloudInfra.Aws.SubnetName = *response.Subnet.Tags[0].Value

		if err := obj.client.ModifySubnetAttribute(ctx, client); err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("[aws] created the subnet ", "id: ", *response.Subnet.Tags[0].Value)

		naclinput := ec2.CreateNetworkAclInput{
			VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VPCID),
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

		naclresp, err := obj.client.BeginCreateNetworkAcl(ctx, client, naclinput)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.NetworkAclID = *naclresp.NetworkAcl.NetworkAclId

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("[aws] created the network acl ", "id", *naclresp.NetworkAcl.NetworkAclId)
	}

	return nil
}

func (obj *AwsProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, resName string) error {

	if len(mainStateDocument.CloudInfra.Aws.RouteTableID) != 0 {
		log.Success("[skip] already created the route table ", "id: ", mainStateDocument.CloudInfra.Aws.RouteTableID)
	} else {
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
			VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VPCID),
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceType("route-table"),
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(obj.clusterName + "-rt"),
						},
					},
				},
			},
		}

		routeresponce, gatewayresp, err := obj.client.BeginCreateVirtNet(internetGateway, routeTableClient, ec2Client, mainStateDocument.CloudInfra.Aws.VPCID)
		if err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}
		log.Success("[aws] created the internet gateway ", "id: ", *gatewayresp.InternetGateway.InternetGatewayId)
		log.Success("[aws] created the route table ", "id: ", *routeresponce.RouteTable.RouteTableId)
	}

	return nil
}
