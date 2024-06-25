package aws

import (
	"context"
	"sync"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"
)

type metadata struct {
	public bool
	cni    string

	noCP int
	noWP int
	noDS int

	k8sVersion string
}

type StsPresignClientInteface interface {
	PresignGetCallerIdentity(ctx context.Context, input *sts.GetCallerIdentityInput, options ...func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type STSTokenRetriever struct {
	PresignClient StsPresignClientInteface
}

type customHTTPPresignerV4 struct {
	client  sts.HTTPPresignerV4
	headers map[string]string
}

type KubeConfigData struct {
	ClusterEndpoint          string
	CertificateAuthorityData string
	ClusterName              string
	Token                    string
}

type AwsProvider struct {
	clusterName string
	haCluster   bool
	region      string
	vpc         string
	metadata    metadata

	mu sync.Mutex

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client AwsGo
}

type AwsGo interface {
	AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error

	InitClient(storage ksctlTypes.StorageFactory) error

	ListLocations() ([]string, error)

	ListVMTypes() (ec2.DescribeInstanceTypesOutput, error)

	BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error)

	BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteVirtNet(ctx context.Context, storage ksctlTypes.StorageFactory) error

	BeginDeleteSubNet(ctx context.Context, storage ksctlTypes.StorageFactory, subnetID string) error

	DeleteSSHKey(ctx context.Context, name string) error

	BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)

	BeginDeleteVM(vmname string) error

	BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error)

	BeginDeleteNIC(nicID string) error

	BeginDeleteVpc(ctx context.Context, storage ksctlTypes.StorageFactory) error

	BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error)

	BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)

	BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error

	DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error)

	FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error)
	GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error)
	AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error

	ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error
	InstanceInitialWaiter(ctx context.Context, instanceInput string) error

	ModifyVpcAttribute(ctx context.Context) error
	ModifySubnetAttribute(ctx context.Context, i int) error

	BeginCreateEKS(ctx context.Context, parameter *eks.CreateClusterInput) (*eks.CreateClusterOutput, error)
	BeignCreateNodeGroup(ctx context.Context, paramter *eks.CreateNodegroupInput) (*eks.CreateNodegroupOutput, error)

	BeginDeleteNodeGroup(ctx context.Context, parameter *eks.DeleteNodegroupInput) (*eks.DeleteNodegroupOutput, error)
	BeginDeleteManagedCluster(ctx context.Context, parameter *eks.DeleteClusterInput) (*eks.DeleteClusterOutput, error)
	DescribeCluster(ctx context.Context, parameter *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error)

	BeginCreateIAM(ctx context.Context, node string, parameter *iam.CreateRoleInput) (*iam.CreateRoleOutput, error)
	BeginDeleteIAM(ctx context.Context, parameter *iam.DeleteRoleInput, node string) (*iam.DeleteRoleOutput, error)

	GetKubeConfig(ctx context.Context, cluster string) (string, error)

	SetRegion(string)
	SetVpc(string) string
}
