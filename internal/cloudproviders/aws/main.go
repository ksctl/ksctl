package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloud_control_res "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

var (
	awsCloudState *StateConfiguration
	GatewayID     string
	RouteTableID  string
	VPCID         string
	SUBNETID      []string
)

type Credential struct {
	AccessKeyID string `json:"access_key_id"`
	Secret      string `json:"secret_access_key"`
}

type AwsProvider struct {
	clusterName string `json:"cluster_name"`
	haCluster   bool   `json:"ha_cluster"`
	region      string `json:"region"`
	vpc         string `json:"vpc"`
	accessKeyID string `json:"access_key_id"`
	secret      string `json:"secret_access_key"`
	session     aws.Config
	metadata

	mxName   sync.Mutex
	mxRole   sync.Mutex
	mxVMType sync.Mutex
	mxState  sync.Mutex

	client  AwsGo
	SSHPath string `json:"ssh_key"`
}

type AWSStateVms struct {
	Names                 []string `json:"names"`
	SecurityGroupName     string   `json:"network_security_group_name"`
	SecurityGroupID       string   `json:"network_security_group_id"`
	DiskNames             []string `json:"disk_name"`
	PrivateIPs            []string `json:"private_ip"`
	PublicIPs             []string `json:"public_ip"`
	NetworkInterfaceNames []string `json:"network_interface_name"`
	NetworkInterfaceIDs   []string `json:"network_interface_id"`
	SubnetNames           []string `json:"subnet_name"`
	SubnetIDs             []string `json:"subnet_id"`
}

var (
	clusterDirName string
	clusterType    KsctlClusterType
)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

type AWSStateVm struct {
	Vpc           string `json:"vpc"`
	Name          string `json:"name"`
	DiskSize      string `json:"disk_size"`
	InstanceType  string `json:"instance_type"`
	Subnet        string `json:"subnet"`
	SecurityGroup string `json:"security_group"`
	PublicIPName  string `json:"public_ip_name"`
	PublicIP      string `json:"public_ip"`
	PrivateIP     string `json:"private_ip"`
}

type StateConfiguration struct {
	IsCompleted bool
	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`
	VPCNAME     string `json:"vpc"`
	VPCID       string `json:"vpc_id"`

	ManagedClusterName string `json:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_nodes"`
	SubnetName         string `json:"subnet_name"`
	SubnetID           string `json:"subnet_id"`

	SecurityGroupRole [4]string `json:"security_group_name"`
	SecurityGroupID   [4]string `json:"security_group_id"`

	GatewayRole    string `json:"gateway_role"`
	GatewayID      string `json:"gateway_id"`
	RouteTableName string `json:"route_table_name"`
	RouteTableID   string `json:"route_table_id"`

	SSHUser          string `json:"ssh_usr"`
	SSHPrivateKeyLoc string `json:"ssh_private_key_location"`
	SSHKeyName       string `json:"sshkey_name"`

	InfoControlPlanes AWSStateVms `json:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database"`
	InfoLoadBalancer  AWSStateVm  `json:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
}

type metadata struct {
	resName string
	role    KsctlRole
	vmType  string
	public  bool

	apps    string
	cni     string
	version string

	noCP int
	noWP int
	noDS int

	k8sName    KsctlKubernetes
	k8sVersion string
}

func ReturnAwsStruct(meta resources.Metadata, ClientOption func() AwsGo) (*AwsProvider, error) {
	return &AwsProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
			k8sName:    meta.K8sDistro,
		},

		session: NEWCLIENT(),
		client:  ClientOption(),
	}, nil
}

func NEWCLIENT() aws.Config {
	NewSession, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"),
		config.WithSharedConfigProfile("default"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")),
	)
	if err != nil {
		panic(err)
	}

	// obj.Session = &NewSession
	fmt.Println("AWS Session Created Successfully")

	return NewSession
}

func (obj *AwsProvider) Name(resName string) resources.CloudFactory {

	obj.mxName.Lock()

	if err := utils.IsValidName(resName); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}

	obj.metadata.resName = resName
	fmt.Println(obj.metadata.resName)
	return obj
}

func (obj *AwsProvider) DelFirewall(factory resources.StorageFactory) error {
	//TODO implement me

	obj.mxRole.Lock()
	fmt.Println("AWS Del Firewall")

	return nil
}

func (obj *AwsProvider) DelNetwork(factory resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Del Network")
	return nil

}

func convertStateFromBytes(raw []byte) error {
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	awsCloudState = data
	return nil
}

func loadStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(UtilClusterPath, CloudAws, clusterType, clusterDirName, STATE_FILE_NAME)
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	return convertStateFromBytes(raw)
}

func (obj *AwsProvider) InitState(storage resources.StorageFactory, opration KsctlOperation) error {

	switch obj.haCluster {
	case false:
		clusterType = ClusterTypeMang
	case true:
		clusterType = ClusterTypeHa
	}
	obj.vpc = fmt.Sprintf("%s-ksctl-%s-vpc", obj.clusterName, clusterType)
	clusterDirName = obj.clusterName + "/" + obj.vpc + "/" + obj.region

	errLoadState := loadStateHelper(storage)
	switch opration {
	case OperationStateCreate:
		if errLoadState == nil && awsCloudState.IsCompleted {
			return fmt.Errorf("cluster %s already exists", obj.clusterName)
		}
		if errLoadState == nil && !awsCloudState.IsCompleted {
			storage.Logger().Note("[aws] RESUME triggered!!")
		} else {
			storage.Logger().Note("[aws] NEW cluster triggered!!")
			awsCloudState = &StateConfiguration{
				IsCompleted:      false,
				ClusterName:      obj.clusterName,
				Region:           obj.region,
				KubernetesDistro: string(obj.metadata.k8sName),
				KubernetesVer:    obj.metadata.k8sVersion,
			}
		}

	case OperationStateDelete:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[aws] Delete resource(s)")

	case OperationStateGet:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[aws] Get resources")
		clusterDirName = awsCloudState.ClusterName + " " + awsCloudState.VPCNAME + " " + awsCloudState.Region
	default:
		return fmt.Errorf("invalid operation")
	}

	// TODO return error  ---------------------------------
	if err := obj.client.InitClient(storage); err != nil {
		fmt.Errorf("failed to initialize aws client reason:%s\n", err.Error())
	}

	obj.client.SetRegion(obj.region)
	obj.client.SetVpc(obj.vpc)

	if err := validationOfArguments(obj); err != nil {
		return err
	}

	storage.Logger().Success("[aws] init cloud state")

	return nil
}

func (obj *AwsProvider) DelSSHKeyPair(storage resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Del SSH Key Pair")
	return nil

}

func (obj *AwsProvider) GetStateForHACluster(storage resources.StorageFactory) (cloud_control_res.CloudResourceState, error) {

	payload := cloud_control_res.CloudResourceState{
		SSHState: cloud_control_res.SSHInfo{
			PathPrivateKey: awsCloudState.SSHPrivateKeyLoc,
			UserName:       awsCloudState.SSHUser,
		},
		Metadata: cloud_control_res.Metadata{
			ClusterName: awsCloudState.ClusterName,
			Provider:    "aws",
			Region:      awsCloudState.Region,
			ClusterType: clusterType,
			ClusterDir:  clusterDirName,
		},
		// public IPs
		IPv4ControlPlanes: awsCloudState.InfoControlPlanes.PublicIPs,
		IPv4DataStores:    awsCloudState.InfoDatabase.PublicIPs,
		IPv4WorkerPlanes:  awsCloudState.InfoWorkerPlanes.PublicIPs,
		IPv4LoadBalancer:  awsCloudState.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: awsCloudState.InfoControlPlanes.PrivateIPs,
		PrivateIPv4DataStores:    awsCloudState.InfoDatabase.PrivateIPs,
		PrivateIPv4LoadBalancer:  awsCloudState.InfoLoadBalancer.PrivateIP,
	}

	storage.Logger().Success("[azure] Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *AwsProvider) NewManagedCluster(factory resources.StorageFactory, i int) error {
	//TODO implement me
	fmt.Println("AWS New Managed Cluster")
	return nil

}

func (obj *AwsProvider) DelManagedCluster(factory resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Del Managed Cluster")
	return nil

}

func (obj *AwsProvider) Role(resRole KsctlRole) resources.CloudFactory {

	obj.mxRole.Lock()

	switch resRole {
	case RoleCp, RoleDs, RoleLb, RoleWp:
		obj.metadata.role = resRole
		return obj
	default:
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err("invalid role assumed")
		return nil
	}

}

func (obj *AwsProvider) VMType(size string) resources.CloudFactory {
	//TODO implement me as soon as possible :)

	//obj.mxVMType.Lock()
	//obj.metadata.vmType = size
	//if err := isValidVMSize(obj, size); err != nil {
	//	var logFactory logger.LogFactory = &logger.Logger{}
	//	logFactory.Err(err.Error())
	//	return nil
	//}
	return obj
}

func (obj *AwsProvider) Visibility(b bool) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS Visibility")
	return obj

}

func (obj *AwsProvider) SupportForApplications() bool {
	//TODO implement me
	fmt.Println("AWS Support for Applications")
	return false

}

func (obj *AwsProvider) SupportForCNI() bool {
	//TODO implement me
	fmt.Println("AWS Support for CNI")
	return false

}

func (obj *AwsProvider) Application(s string) bool {
	//TODO implement me
	fmt.Println("AWS Application")
	return false

}

func (obj *AwsProvider) CNI(s string) bool {
	//TODO implement me
	fmt.Println("AWS CNI")
	return false

}

func (obj *AwsProvider) Version(s string) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS Version")
	return nil

}

func (obj *AwsProvider) NoOfWorkerPlane(factory resources.StorageFactory, i int, b bool) (int, error) {
	//TODO implement me
	fmt.Println("AWS No of Worker Plane")
	i = 0
	return i, nil

}

func (obj *AwsProvider) NoOfControlPlane(i int, b bool) (int, error) {
	//TODO implement me
	fmt.Println("AWS No of Control Plane")
	i = 0
	return i, nil

}

func (obj *AwsProvider) NoOfDataStore(i int, b bool) (int, error) {
	//TODO implement me
	fmt.Println("AWS No of Data Store")
	i = 0
	return i, nil

}

func (obj *AwsProvider) GetHostNameAllWorkerNode() []string {
	//TODO implement me
	fmt.Println("AWS Get Host Name All Worker Node")
	return nil

}

func (obj *AwsProvider) SwitchCluster(factory resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Switch Cluster")
	return nil

}

func (obj *AwsProvider) GetStateFile(factory resources.StorageFactory) (string, error) {

	cloudstate, err := json.Marshal(awsCloudState)
	if err != nil {
		return "", err
	}
	return string(cloudstate), nil
}

func (obj *AwsProvider) GetSecretTokens(factory resources.StorageFactory) (map[string][]byte, error) {
	//TODO implement me
	acesskeyid := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	return map[string][]byte{
		"aws_access_key_id":     []byte(acesskeyid),
		"aws_secret_access_key": []byte(secret),
	}, nil
}

func isValidVMSize(obj *AwsProvider, size string) error {
	// TODO implement me
	validSize, err := obj.client.ListVMTypes()
	if err != nil {
		return err
	}

	for _, valid := range validSize {
		if valid == size {
			return nil
		}
	}

	return fmt.Errorf("INVALID VM SIZE\nValid options %v\n", validSize)
}
