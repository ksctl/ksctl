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
		log.Print("[skip] already deleted the vpc", mainStateDocument.CloudInfra.Aws.VpcName)
	} else {
		err := obj.DeleteSubnet(context.Background(), storage, mainStateDocument.CloudInfra.Aws.SubnetID)
		if err != nil {
			return err
		}
	}

	err := obj.client.BeginDeleteVirtNet(context.Background(), storage)
	if err != nil {
		return err
	}

	if mainStateDocument.CloudInfra.Aws.VpcId == "" {
		log.Success("[aws] deleted the vpc ", "id", mainStateDocument.CloudInfra.Aws.VpcName)
	} else {
		err = obj.DeleteVpc(context.Background(), storage, mainStateDocument.CloudInfra.Aws.VpcId)
		if err != nil {
			return err
		}
	}

	if err := storage.DeleteCluster(); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("[aws] deleted the vpc ", "id", mainStateDocument.CloudInfra.Aws.VpcName)

	return nil
}

func (obj *AwsProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	err := obj.client.BeginDeleteSubNet(ctx, storage, subnetName)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Aws.SubnetID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("[aws] deleted the subnet ", "id", mainStateDocument.CloudInfra.Aws.SubnetName)

	return nil
}

func (obj *AwsProvider) DeleteVpc(ctx context.Context, storage resources.StorageFactory, resName string) error {

	err := obj.client.BeginDeleteVpc(ctx, storage)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Aws.VpcId = ""
	mainStateDocument.CloudInfra.Aws.VpcName = ""
	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("deleted the vpc ", "id", mainStateDocument.CloudInfra.Aws.VpcName)
	return nil
}

func (obj *AwsProvider) NewNetwork(storage resources.StorageFactory) error {
	_ = <-obj.chResName

	if len(mainStateDocument.CloudInfra.Aws.VpcId) != 0 {
		log.Print("[skip] already created the vpc", mainStateDocument.CloudInfra.Aws.VpcName)
	} else {
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

		vpc, err := obj.client.BeginCreateVpc(vpcclient)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.VpcId = *vpc.Vpc.VpcId
		mainStateDocument.CloudInfra.Aws.VpcName = *vpc.Vpc.Tags[0].Value

		if err := obj.client.ModifyVpcAttribute(context.Background()); err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("created the vpc ", "id", *vpc.Vpc.VpcId)

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

		parameter := ec2.CreateSubnetInput{
			CidrBlock: aws.String("172.31.32.0/20"),
			VpcId:     aws.String(mainStateDocument.CloudInfra.Aws.VpcId),

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
		response, err := obj.client.BeginCreateSubNet(ctx, subnetName, parameter)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.SubnetID = *response.Subnet.SubnetId
		mainStateDocument.CloudInfra.Aws.SubnetName = *response.Subnet.Tags[0].Value

		if err := obj.client.ModifySubnetAttribute(ctx); err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("created the subnet ", "id", *response.Subnet.Tags[0].Value)

		naclinput := ec2.CreateNetworkAclInput{
			VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
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

		naclresp, err := obj.client.BeginCreateNetworkAcl(ctx, naclinput)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.NetworkAclID = *naclresp.NetworkAcl.NetworkAclId

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("created the network acl ", "id", *naclresp.NetworkAcl.NetworkAclId)
	}

	return nil
}

func (obj *AwsProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, resName string) error {

	if len(mainStateDocument.CloudInfra.Aws.RouteTableID) != 0 {
		log.Success("[skip] already created the route table ", "id", mainStateDocument.CloudInfra.Aws.RouteTableID)
	} else {
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
			VpcId: aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
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

		routeresponce, gatewayresp, err := obj.client.BeginCreateVirtNet(internetGateway, routeTableClient, mainStateDocument.CloudInfra.Aws.VpcId)
		if err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}
		log.Success("[aws] created the internet gateway ", "id", *gatewayresp.InternetGateway.InternetGatewayId)
		log.Success("[aws] created the route table ", "id", *routeresponce.RouteTable.RouteTableId)
	}

	return nil
}
