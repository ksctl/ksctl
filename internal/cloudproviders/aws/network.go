package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"
)

func (obj *AwsProvider) DelNetwork(storage ksctlTypes.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Aws.SubnetID) == 0 {
		log.Print(awsCtx, "skipped already deleted the vpc", "name", mainStateDocument.CloudInfra.Aws.VpcName)
	} else {
		err := obj.DeleteSubnet(awsCtx, storage, mainStateDocument.CloudInfra.Aws.SubnetID)
		if err != nil {
			return err
		}
	}

	err := obj.client.BeginDeleteVirtNet(awsCtx, storage)
	if err != nil {
		return err
	}

	if mainStateDocument.CloudInfra.Aws.VpcId == "" {
		log.Success(awsCtx, "Deleted the vpc", "id", mainStateDocument.CloudInfra.Aws.VpcName)
	} else {
		err = obj.DeleteVpc(awsCtx, storage, mainStateDocument.CloudInfra.Aws.VpcId)
		if err != nil {
			return err
		}
	}

	log.Success(awsCtx, "Deleted the vpc", "name", mainStateDocument.CloudInfra.Aws.VpcName)

	if err := storage.DeleteCluster(); err != nil {
		return err
	}

	return nil
}

func (obj *AwsProvider) DeleteSubnet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetName string) error {

	err := obj.client.BeginDeleteSubNet(ctx, storage, subnetName)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Aws.SubnetID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(awsCtx, "Deleted the subnet", "id", mainStateDocument.CloudInfra.Aws.SubnetName)

	return nil
}

func (obj *AwsProvider) DeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory, resName string) error {

	err := obj.client.BeginDeleteVpc(ctx, storage)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Aws.VpcId = ""
	name := mainStateDocument.CloudInfra.Aws.VpcName
	mainStateDocument.CloudInfra.Aws.VpcName = ""
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(awsCtx, "Deleted the vpc", "name", name)
	return nil
}

func (obj *AwsProvider) NewNetwork(storage ksctlTypes.StorageFactory) error {
	<-obj.chResName

	if len(mainStateDocument.CloudInfra.Aws.VpcId) != 0 {
		log.Print(awsCtx, "skipped already created the vpc", mainStateDocument.CloudInfra.Aws.VpcName)
	} else {
		vpcclient := ec2.CreateVpcInput{
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
		mainStateDocument.CloudInfra.Aws.VpcCidr = "172.31.0.0/16"

		log.Debug(awsCtx, "Printing", "virtualprivatecloud", vpcclient)

		vpc, err := obj.client.BeginCreateVpc(vpcclient)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.VpcId = *vpc.Vpc.VpcId
		mainStateDocument.CloudInfra.Aws.VpcName = *vpc.Vpc.Tags[0].Value

		if err := obj.client.ModifyVpcAttribute(awsCtx); err != nil {
			return err
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return err
		}

		log.Success(awsCtx, "created the vpc", "id", *vpc.Vpc.VpcId)

	}

	if obj.haCluster {
		virtNet := obj.clusterName + "-vnet"
		subNet := obj.clusterName + "-subnet"

		if err := obj.CreateSubnet(awsCtx, storage, subNet); err != nil {
			return err
		}

		if err := obj.CreateVirtualNetwork(awsCtx, storage, virtNet); err != nil {
			return err
		}

	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	return nil
}

func (obj *AwsProvider) CreateSubnet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetName string) error {

	zones, err := obj.client.GetAvailabilityZones()
	if err != nil {
		return err
	}

	log.Print(awsCtx, "Selected availability zone", "zone", *zones.AvailabilityZones[0].ZoneName)

	if len(mainStateDocument.CloudInfra.Aws.SubnetID) != 0 {
		log.Print(awsCtx, "skipped already created the subnet", mainStateDocument.CloudInfra.Aws.SubnetID)
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
			AvailabilityZone: aws.String(*zones.AvailabilityZones[0].ZoneName),
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
			return err
		}

		log.Success(awsCtx, "created the subnet", "id", *response.Subnet.Tags[0].Value)

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
			return err
		}

		log.Success(awsCtx, "created the network acl", "id", *naclresp.NetworkAcl.NetworkAclId)
	}

	return nil
}

func (obj *AwsProvider) CreateVirtualNetwork(ctx context.Context, storage ksctlTypes.StorageFactory, resName string) error {

	if len(mainStateDocument.CloudInfra.Aws.RouteTableID) != 0 {
		log.Success(awsCtx, "skipped already created the route table ", "id", mainStateDocument.CloudInfra.Aws.RouteTableID)
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

		mainStateDocument.CloudInfra.Aws.RouteTableID = *routeresponce.RouteTable.RouteTableId
		mainStateDocument.CloudInfra.Aws.GatewayID = *gatewayresp.InternetGateway.InternetGatewayId

		if err := storage.Write(mainStateDocument); err != nil {
			return err
		}
		log.Success(awsCtx, "created the internet gateway", "id", *gatewayresp.InternetGateway.InternetGatewayId)
		log.Success(awsCtx, "created the route table", "id", *routeresponce.RouteTable.RouteTableId)
	}

	return nil
}
