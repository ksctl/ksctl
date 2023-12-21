package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func (obj *AwsProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(awsCloudState.SubnetID) == 0 {
		log.Print("[skip] already deleted the vpc", awsCloudState.VPCNAME)
	} else {
		err := obj.DeleteSubnet(context.Background(), storage, awsCloudState.SubnetID)
		if err != nil {
			return err
		}
	}

	err := obj.client.BeginDeleteVirtNet(context.Background(), storage, obj.ec2Client())
	if err != nil {
		return err
	}

	if awsCloudState.VPCID == "" {
		log.Success("[aws] deleted the vpc ", "id: ", awsCloudState.VPCNAME)
	} else {
		err = obj.DeleteVpc(context.Background(), storage, awsCloudState.VPCID)
		if err != nil {
			return err
		}
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the vpc ", "id: ", awsCloudState.VPCNAME)

	return nil
}

func (obj *AwsProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	err := obj.client.BeginDeleteSubNet(ctx, storage, subnetName, obj.ec2Client())
	if err != nil {
		return err
	}
	awsCloudState.SubnetID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the subnet ", "id: ", awsCloudState.SubnetName)

	return nil
}

func (obj *AwsProvider) DeleteVpc(ctx context.Context, storage resources.StorageFactory, resName string) error {

	err := obj.client.BeginDeleteVpc(ctx, storage, obj.ec2Client())
	if err != nil {
		return err
	}
	awsCloudState.VPCID = ""
	awsCloudState.VPCNAME = ""
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("[aws] deleted the vpc ", "id", awsCloudState.VPCNAME)
	return nil
}

func (obj *AwsProvider) NewNetwork(storage resources.StorageFactory) error {
	_ = obj.metadata.resName
	obj.mxName.Unlock()

	if err := storage.Path(generatePath(consts.UtilClusterPath, clusterType, clusterDirName)).
		Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		return log.NewError(err.Error())
	}

	if len(awsCloudState.VPCID) != 0 {
		log.Print("[skip] already created the vpc", awsCloudState.VPCNAME)
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

		vpc, err := obj.client.BeginCreateVpc(ec2client, vpcclient)
		if err != nil {
			return err
		}

		awsCloudState.VPCID = *vpc.Vpc.VpcId
		awsCloudState.VPCNAME = *vpc.Vpc.Tags[0].Value

		log.Success("[aws] created the vpc ", "id: ", *vpc.Vpc.VpcId)

		if err := obj.client.ModifyVpcAttribute(context.Background(), ec2client); err != nil {
			return err
		}

		// if err := storage.Path(generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)).
		// 	Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		// 	return err
		// }

		if err := saveStateHelper(storage); err != nil {
			return err
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

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	return nil
}

func (obj *AwsProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(awsCloudState.SubnetID) != 0 {
		log.Print("[skip] already created the subnet", awsCloudState.SubnetID)
	} else {

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
		}
		response, err := obj.client.BeginCreateSubNet(ctx, subnetName, client, parameter)
		if err != nil {
			return err
		}

		awsCloudState.SubnetID = *response.Subnet.SubnetId
		awsCloudState.SubnetName = *response.Subnet.Tags[0].Value

		if err := obj.client.ModifySubnetAttribute(ctx, client); err != nil {
			return err
		}

		if err := saveStateHelper(storage); err != nil {
			return err
		}
		log.Success("[aws] created the subnet ", "id: ", *response.Subnet.Tags[0].Value)

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

		naclresp, err := obj.client.BeginCreateNetworkAcl(ctx, client, naclinput)
		if err != nil {
			return err
		}

		awsCloudState.NetworkAclID = *naclresp.NetworkAcl.NetworkAclId

		if err := saveStateHelper(storage); err != nil {
			log.Error("Error saving state", "error", err)
		}
		log.Success("[aws] created the network acl ", "id", *naclresp.NetworkAcl.NetworkAclId)
	}

	return nil
}

// Implements internetgateway, route table
func (obj *AwsProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, resName string) error {

	if len(awsCloudState.RouteTableID) != 0 {
		log.Success("[skip] already created the route table ", "id: ", awsCloudState.RouteTableID)
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

		if err := saveStateHelper(storage); err != nil {
			return err
		}

		log.Success("[aws] created the internet gateway ", "id: ", *gatewayresp.InternetGateway.InternetGatewayId)
		log.Success("[aws] created the route table ", "id: ", *routeresponce.RouteTable.RouteTableId)
	}

	return nil
}
