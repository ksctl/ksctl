package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func (obj *AwsProvider) NewNetwork(storage resources.StorageFactory) error {
	_ = obj.metadata.resName
	if len(awsCloudState.VPCID) != 0 {
		storage.Logger().Success("[skip] already created the vpc", awsCloudState.VPCNAME)
	}
	ec2client := ec2.NewFromConfig(obj.session)
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

	vpc, err := obj.client.BeginCreateVpc(ec2client, vpcclient)
	if err != nil {
		return err
	}
	awsCloudState.VPCID = *vpc.Vpc.VpcId
	awsCloudState.VPCNAME = *vpc.Vpc.Tags[0].Value

	if err := storage.Path(generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)).
		Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		return err
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[aws] created the vpc ", *vpc.Vpc.VpcId)

	ctx := context.TODO()

	if obj.haCluster {
		virtNet := obj.clusterName + "-vnet"
		subNet := obj.clusterName + "-subnet"
		// virtual net

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
		storage.Logger().Success("[skip] already created the subnet", awsCloudState.SubnetID)
	}

	client := obj.ec2Client()

	parameter := ec2.CreateSubnetInput{
		CidrBlock: aws.String("172.31.16.0/20"),
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

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[aws] created the subnet ", *response.Subnet.Tags[0].Value)

	return nil
}

func saveStateHelper(storage resources.StorageFactory) error {

	return nil
}

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
		VpcId: aws.String(obj.vpc),
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

	response, err := obj.client.BeginCreateVirtNet(internetGateway, routeTableClient, ec2Client)
	if err != nil {
		return err
	}

	// TODO this is wrong impletimentation to prevent the error
	fmt.Println(response)

	// 3. Create  Internet Gateway            DONE
	// 4. Create  Route Table				  DONE
	// 5. Create  Firewall aka Security Group in AWS       implemented later on based on nodes role like worker or master
	// 6. Create Load Balancer				 TODO
	// create lb
	// create target group
	// register target group
	// create listener

	return nil
}
