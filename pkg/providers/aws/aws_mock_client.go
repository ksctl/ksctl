//go:build testing_aws

package aws

import (
	"context"

	eksTypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func ProvideClient() AwsGo {
	return &AwsClient{}
}

type AwsClient struct {
	region string
}

func (*AwsClient) AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error {
	return nil
}

func (mock *AwsClient) AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error {
	return nil
}

func (mock *AwsClient) BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {

	nic := &ec2.CreateNetworkInterfaceOutput{
		NetworkInterface: &types.NetworkInterface{
			NetworkInterfaceId: aws.String("test-nic-1234567890"),
			TagSet: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(*parameter.TagSpecifications[0].Tags[0].Value),
				},
			},
		},
	}

	return nic, nil

}

func (mock *AwsClient) BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	subnet := &ec2.CreateSubnetOutput{
		Subnet: &types.Subnet{
			SubnetId: aws.String("3456d25f36g474g546"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(*parameter.TagSpecifications[0].Tags[0].Value),
				},
			},
		},
	}

	return subnet, nil
}

func (mock *AwsClient) BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {

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

func (mock *AwsClient) BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error) {

	routeTable := &ec2.CreateRouteTableOutput{
		RouteTable: &types.RouteTable{
			RouteTableId: aws.String("3456d25f36g474g546"),
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
			InternetGatewayId: aws.String("3456d25f36g474g546"),
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

func (mock *AwsClient) BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	vpc := &ec2.CreateVpcOutput{
		Vpc: &types.Vpc{
			VpcId: aws.String("3456d25f36g474g546"),
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(*parameter.TagSpecifications[0].Tags[0].Value),
				},
			},
		},
	}
	return vpc, nil
}

func (mock *AwsClient) BeginDeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	mainStateDocument.CloudInfra.Aws.VpcId = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(awsCtx, "Error Writing State File", "Reason", err)
	}

	log.Success(awsCtx, "deleted the vpc", "id", mainStateDocument.CloudInfra.Aws.VpcId)

	return nil

}

func (mock *AwsClient) BeginDeleteNIC(nicID string) error {

	return nil
}

func (mock *AwsClient) FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error) {
	return "ami-1234567890", nil
}

func (mock *AwsClient) BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error {

	return nil
}

func (mock *AwsClient) GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error) {
	return &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []types.AvailabilityZone{
			{
				ZoneName: aws.String("us-east-1a"),
			},
			{
				ZoneName: aws.String("us-east-1b"),
			},
			{
				ZoneName: aws.String("us-east-1c"),
			},
		},
	}, nil
}

func (mock *AwsClient) BeginDeleteSubNet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetID string) error {

	for i := 0; i < len(mainStateDocument.CloudInfra.Aws.SubnetIDs); i++ {
		mainStateDocument.CloudInfra.Aws.SubnetIDs[i] = ""

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(awsCtx, "Error Writing State File", "Reason", err)
		}

		log.Success(awsCtx, "deleted the subnet ", mainStateDocument.CloudInfra.Aws.SubnetNames)

	}

	return nil

}

func (mock *AwsClient) BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error) {

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

func (mock *AwsClient) BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {

	securitygroup := &ec2.CreateSecurityGroupOutput{
		GroupId: aws.String("test-security-group-1234567890"),
	}

	return securitygroup, nil
}

func (mock *AwsClient) BeginDeleteVM(vmname string) error {
	return nil
}

func (mock *AwsClient) BeginDeleteVirtNet(ctx context.Context, storage ksctlTypes.StorageFactory) error {

	mainStateDocument.CloudInfra.Aws.GatewayID = ""
	mainStateDocument.CloudInfra.Aws.RouteTableID = ""
	mainStateDocument.CloudInfra.Aws.NetworkAclID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(awsCtx, "Error Writing State File", "Reason", err)
	}

	return nil
}

func (mock *AwsClient) CreateSSHKey() error {
	return nil
}

func (mock *AwsClient) DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error) {

	instanceinforesponse := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
						PublicIpAddress:  aws.String("A.B.C.D"),
						PrivateIpAddress: aws.String("192.168.1.2"),
					},
				},
			},
		},
	}

	return instanceinforesponse, nil
}

func (mock *AwsClient) DeleteSSHKey(ctx context.Context, name string) error {

	return nil
}

func (mock *AwsClient) InstanceInitialWaiter(ctx context.Context, instanceID string) error {
	return nil
}

func (mock *AwsClient) InitClient(storage ksctlTypes.StorageFactory) error {
	return nil
}

func (mock *AwsClient) ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error {

	return nil
}

func (awsclient *AwsClient) ListLocations() ([]string, error) {

	return []string{"fake-region"}, nil
}

func (mock *AwsClient) ListVMTypes() (ec2.DescribeInstanceTypesOutput, error) {
	return ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []types.InstanceTypeInfo{
			{
				InstanceType: "fake",
			},
		},
	}, nil
}

func (mock *AwsClient) ModifyVpcAttribute(ctx context.Context) error {
	return nil
}

func (mock *AwsClient) ModifySubnetAttribute(ctx context.Context, i int) error {
	return nil
}
func (mock *AwsClient) SetRegion(string) {
	mock.region = "fake-region"
}

func (mock *AwsClient) SetVpc(string) string {
	return "fake-vpc"
}

func (mock *AwsClient) InitState(storage ksctlTypes.StorageFactory) error {
	return nil
}

func (mock *AwsClient) BeginCreateEKS(ctx context.Context, paramter *eks.CreateClusterInput) (*eks.CreateClusterOutput, error) {
	return &eks.CreateClusterOutput{
		Cluster: &eksTypes.Cluster{
			Name: aws.String("test-cluster"),
		},
	}, nil
}
func (mock *AwsClient) BeginCreateNodeGroup(ctx context.Context, paramter *eks.CreateNodegroupInput) (*eks.CreateNodegroupOutput, error) {
	return &eks.CreateNodegroupOutput{
		Nodegroup: &eksTypes.Nodegroup{
			NodegroupName: aws.String("test-nodegroup"),
			NodegroupArn:  aws.String("arn:aws:eks:us-west-2:1234567890:nodegroup/test-cluster/test-nodegroup"),
		},
	}, nil
}

func (mock *AwsClient) BeginDeleteNodeGroup(ctx context.Context, parameter *eks.DeleteNodegroupInput) (*eks.DeleteNodegroupOutput, error) {
	return &eks.DeleteNodegroupOutput{
		Nodegroup: &eksTypes.Nodegroup{
			NodegroupName: aws.String("test-nodegroup"),
		},
	}, nil
}

func (mock *AwsClient) BeginDeleteManagedCluster(ctx context.Context, parameter *eks.DeleteClusterInput) (*eks.DeleteClusterOutput, error) {
	return &eks.DeleteClusterOutput{
		Cluster: &eksTypes.Cluster{
			Name: aws.String("test-cluster"),
		},
	}, nil
}

func (mock *AwsClient) BeginCreateIAM(ctx context.Context, node string, parameter *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	return &iam.CreateRoleOutput{
		Role: &iamTypes.Role{
			RoleName: aws.String("test-role"),
			Arn:      aws.String("arn:aws:iam::1234567890:role/test-role"),
		},
	}, nil
}

func (mock *AwsClient) BeginDeleteIAM(ctx context.Context, parameter *iam.DeleteRoleInput, role string) (*iam.DeleteRoleOutput, error) {
	return &iam.DeleteRoleOutput{
		ResultMetadata: middleware.Metadata{},
	}, nil

}

func (mock *AwsClient) DescribeCluster(ctx context.Context, parameter *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error) {
	return &eks.DescribeClusterOutput{
		Cluster: &eksTypes.Cluster{
			Name: aws.String("test-cluster"),
		},
	}, nil
}

func (mock *AwsClient) GetKubeConfig(ctx context.Context, cluster string) (string, error) {
	return "fake-kubeconfig", nil
}

func (mock *AwsClient) ListK8sVersions(ctx context.Context) ([]string, error) {
	return []string{"1.30", "1.29"}, nil
}
