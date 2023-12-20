package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloud_control_res "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
)

var (
	awsCloudState *StateConfiguration
	GatewayID     string
	RouteTableID  string
	VPCID         string
	SUBNETID      []string

	log = logger.NewDefaultLogger(-1, os.Stdout)
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

func (obj *AwsProvider) GetKubeconfigPath() string {

	return helpers.GetPath(consts.UtilClusterPath, consts.CloudAws, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
}

type AWSStateVms struct {
	Names []string `json:"names"`
	//SecurityGroupName     string   `json:"network_security_group_name"`
	//SecurityGroupID       string   `json:"network_security_group_id"`
	DiskNames  []string `json:"disk_name"`
	PrivateIPs []string `json:"private_ip"`
	PublicIPs  []string `json:"public_ip"`
	//NetworkInterfaceNames []string `json:"network_interface_name"`
	NetworkInterfaceIDs  []string `json:"network_interface_id"`
	SubnetNames          []string `json:"subnet_name"`
	SubnetIDs            []string `json:"subnet_id"`
	NetworkSecurityGroup string   `json:"network_security_group"`
}

var (
	clusterDirName string
	clusterType    consts.KsctlClusterType
)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

type AWSStateVm struct {
	Vpc                  string `json:"vpc"`
	Name                 string `json:"name"`
	DiskSize             string `json:"disk_size"`
	InstanceType         string `json:"instance_type"`
	Subnet               string `json:"subnet"`
	NetworkSecurityGroup string `json:"network_security_group"`
	PublicIPName         string `json:"public_ip_name"`
	PublicIP             string `json:"public_ip"`
	PrivateIP            string `json:"private_ip"`
	NetworkInterfaceName string `json:"network_interface_name"`
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
	NetworkAclID       string `json:"network_acl_id"`

	//SecurityGroupRole [4]string `json:"security_group_name"`
	//SecurityGroupID   [4]string `json:"security_group_id"`

	//GatewayRole    string `json:"gateway_role"`
	GatewayID string `json:"gateway_id"`
	//RouteTableName string `json:"route_table_name"`
	RouteTableID string `json:"route_table_id"`

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
	role    consts.KsctlRole
	vmType  string
	public  bool

	apps    string
	cni     string
	version string

	noCP int
	noWP int
	noDS int

	k8sName    consts.KsctlKubernetes
	k8sVersion string
}

func ReturnAwsStruct(meta resources.Metadata, ClientOption func() AwsGo) (*AwsProvider, error) {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAws))

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
		log.Error(err.Error())
	}

	// obj.Session = &NewSession
	fmt.Println("AWS Session Created Successfully")

	return NewSession
}

func (obj *AwsProvider) Name(resName string) resources.CloudFactory {

	obj.mxName.Lock()

	if err := helpers.IsValidName(resName); err != nil {
		log.Error(err.Error())
		return nil
	}

	obj.metadata.resName = resName
	return obj
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
	path := helpers.GetPath(consts.UtilClusterPath, consts.CloudAws, clusterType, clusterDirName, STATE_FILE_NAME)
	fmt.Println(path)
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	log.Print("raw", "", string(raw))
	return convertStateFromBytes(raw)
}

func (obj *AwsProvider) InitState(storage resources.StorageFactory, opration consts.KsctlOperation) error {

	switch obj.haCluster {
	case false:
		clusterType = consts.ClusterTypeMang
	case true:
		clusterType = consts.ClusterTypeHa
	}
	obj.vpc = fmt.Sprintf("%s-ksctl-%s-vpc", obj.clusterName, clusterType)
	clusterDirName = obj.clusterName + " " + obj.vpc + " " + obj.region
	log.Print("clusterDirName", "", clusterDirName)

	errLoadState := loadStateHelper(storage)

	switch opration {
	case consts.OperationStateCreate:
		if errLoadState == nil && awsCloudState.IsCompleted {
			return fmt.Errorf("cluster %s already exists", obj.clusterName)
		}
		if errLoadState == nil && !awsCloudState.IsCompleted {
			log.Warn("Cluster state found but not completed, resuming operation")
		} else {
			log.Print("Fresh state!!")
			awsCloudState = &StateConfiguration{
				IsCompleted:      false,
				ClusterName:      obj.clusterName,
				Region:           obj.region,
				KubernetesDistro: string(obj.metadata.k8sName),
				KubernetesVer:    obj.metadata.k8sVersion,
			}
		}

	case consts.OperationStateDelete:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Print("[aws] Delete resource(s)")

	case consts.OperationStateGet:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Print("[aws] Get resources")
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

	log.Success("[aws] init cloud state")

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

	log.Success("[aws] Transferred Data, it's ready to be shipped!")
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

func (obj *AwsProvider) Role(resRole consts.KsctlRole) resources.CloudFactory {

	obj.mxRole.Lock()

	switch resRole {
	case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
		obj.metadata.role = resRole
		return obj
	default:
		log.Error("invalid role assumed")

		return nil
	}
}

func (obj *AwsProvider) VMType(size string) resources.CloudFactory {
	obj.mxVMType.Lock()
	obj.metadata.vmType = size
	if err := isValidVMSize(obj, size); err != nil {
		log.Error("invalid vm size assumed")
	}
	return obj
}

func (obj *AwsProvider) Visibility(toBePublic bool) resources.CloudFactory {
	obj.metadata.public = toBePublic
	return obj
}

func (obj *AwsProvider) SupportForApplications() bool {
	//TODO implement me
	fmt.Println("AWS Support for Applications")
	return false

}

//func (obj *AwsProvider) SupportForCNI() bool {
//	//TODO implement me
//	fmt.Println("AWS Support for CNI")
//	return false
//
//}

func (obj *AwsProvider) Application(s string) bool {
	return true
}

func (obj *AwsProvider) CNI(s string) bool {
	return true
}

func (obj *AwsProvider) Version(s string) resources.CloudFactory {
	fmt.Println("AWS Version")
	return nil
}

func (obj *AwsProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	log.Print("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if awsCloudState == nil {
			return -1, log.NewError("state init not called")
		}
		if awsCloudState.InfoWorkerPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}
		//log.Print("Prnting", "awsCloudState.InfoWorkerPlanes.Names", awsCloudState.InfoWorkerPlanes.Names)
		return len(awsCloudState.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if awsCloudState == nil {
			return -1, log.NewError("state init not called")
		}
		currLen := len(awsCloudState.InfoWorkerPlanes.Names)

		newLen := no

		if currLen == 0 {
			awsCloudState.InfoWorkerPlanes.Names = make([]string, no)
			//awsCloudState.InfoWorkerPlanes.Hostnames = make([]string, no)
			awsCloudState.InfoWorkerPlanes.PublicIPs = make([]string, no)
			awsCloudState.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			awsCloudState.InfoWorkerPlanes.DiskNames = make([]string, no)
			//awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames = make([]string, no)
			awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
			//awsCloudState.InfoWorkerPlanes.PublicIPNames = make([]string, no)
			//awsCloudState.InfoWorkerPlanes.PublicIPIDs = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					awsCloudState.InfoWorkerPlanes.Names = append(awsCloudState.InfoWorkerPlanes.Names, "")
					//awsCloudState.InfoWorkerPlanes.Hostnames = append(awsCloudState.InfoWorkerPlanes.Hostnames, "")
					awsCloudState.InfoWorkerPlanes.PublicIPs = append(awsCloudState.InfoWorkerPlanes.PublicIPs, "")
					awsCloudState.InfoWorkerPlanes.PrivateIPs = append(awsCloudState.InfoWorkerPlanes.PrivateIPs, "")
					awsCloudState.InfoWorkerPlanes.DiskNames = append(awsCloudState.InfoWorkerPlanes.DiskNames, "")
					//awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames = append(awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames, "")
					awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs = append(awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs, "")
					//awsCloudState.InfoWorkerPlanes.PublicIPNames = append(awsCloudState.InfoWorkerPlanes.PublicIPNames, "")
					//awsCloudState.InfoWorkerPlanes.PublicIPIDs = append(awsCloudState.InfoWorkerPlanes.PublicIPIDs, "")
				}
			} else {
				// for downscaling
				awsCloudState.InfoWorkerPlanes.Names = awsCloudState.InfoWorkerPlanes.Names[:newLen]
				//awsCloudState.InfoWorkerPlanes.Hostnames = awsCloudState.InfoWorkerPlanes.Hostnames[:newLen]
				awsCloudState.InfoWorkerPlanes.PublicIPs = awsCloudState.InfoWorkerPlanes.PublicIPs[:newLen]
				awsCloudState.InfoWorkerPlanes.PrivateIPs = awsCloudState.InfoWorkerPlanes.PrivateIPs[:newLen]
				awsCloudState.InfoWorkerPlanes.DiskNames = awsCloudState.InfoWorkerPlanes.DiskNames[:newLen]
				//awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames = awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames[:newLen]
				awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs = awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
				//awsCloudState.InfoWorkerPlanes.PublicIPNames = awsCloudState.InfoWorkerPlanes.PublicIPNames[:newLen]
				//awsCloudState.InfoWorkerPlanes.PublicIPIDs = awsCloudState.InfoWorkerPlanes.PublicIPIDs[:newLen]
			}
		}

		if err := saveStateHelper(storage); err != nil {
			return -1, err
		}

		//log.Print("Printing", "awsCloudState.InfoWorkerPlanes", awsCloudState.InfoWorkerPlanes)

		return -1, nil
	}
	return -1, log.NewError("constrains for no of workplane >= 0")

}

func (obj *AwsProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	//TODO implement me
	if !setter {
		// delete operation
		if awsCloudState == nil {
			return -1, log.NewError("state init not called")
		}
		if awsCloudState.InfoControlPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}

		log.Print("Printing", "awsCloudState.InfoControlPlanes.Names", awsCloudState.InfoControlPlanes.Names)
		return len(awsCloudState.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if awsCloudState == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(awsCloudState.InfoControlPlanes.Names)
		if currLen == 0 {
			// What id hostname ?
			awsCloudState.InfoControlPlanes.Names = make([]string, no)
			//awsCloudState.InfoControlPlanes.Hostnames = make([]string, no)
			awsCloudState.InfoControlPlanes.PublicIPs = make([]string, no)
			awsCloudState.InfoControlPlanes.PrivateIPs = make([]string, no)
			awsCloudState.InfoControlPlanes.DiskNames = make([]string, no)
			//awsCloudState.InfoControlPlanes.NetworkInterfaceNames = make([]string, no)
			awsCloudState.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
			//awsCloudState.InfoControlPlanes.PublicIPs = make([]string, no)
			//awsCloudState.InfoControlPlanes.PublicIPIDs = make([]string, no)
		}

		log.Debug("Printing", "awsCloudState.InfoControlplanes", awsCloudState.InfoControlPlanes)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of controlplane >= 3 and odd number")

}

func (obj *AwsProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Print("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if awsCloudState == nil {
			return -1, log.NewError("state init not called")
		}
		if awsCloudState.InfoDatabase.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}

		log.Print("Printing", "awsCloudState.InfoDatabase.Names", awsCloudState.InfoDatabase.Names)
		return len(awsCloudState.InfoDatabase.Names), nil
	}
	if no >= 1 && (no&1) == 1 {
		obj.metadata.noDS = no

		if awsCloudState == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(awsCloudState.InfoDatabase.Names)
		if currLen == 0 {
			awsCloudState.InfoDatabase.Names = make([]string, no)
			//awsCloudState.InfoDatabase.Hostnames = make([]string, no)
			awsCloudState.InfoDatabase.PublicIPs = make([]string, no)
			awsCloudState.InfoDatabase.PrivateIPs = make([]string, no)
			awsCloudState.InfoDatabase.DiskNames = make([]string, no)
			//awsCloudState.InfoDatabase.NetworkInterfaceNames = make([]string, no)
			awsCloudState.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
			//awsCloudState.InfoDatabase.PublicIPNames = make([]string, no)
			//awsCloudState.InfoDatabase.PublicIPIDs = make([]string, no)
		}

		log.Debug("Printing", "awsCloudState.InfoDatabase", awsCloudState.InfoDatabase)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of Datastore>= 1 and odd number")
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
	log.Print("Printing", "cloudstate", cloudstate)

	// now create a file cloud-state.json in .ksctl/cluster/aws/clusterName
	_, err = os.Create(helpers.GetPath(consts.UtilClusterPath, consts.CloudAws, clusterType, clusterDirName, STATE_FILE_NAME))
	if err != nil {
		return "", err
	}
	//now write the cloudstate in the file

	os.WriteFile(helpers.GetPath(consts.UtilClusterPath, consts.CloudAws, clusterType, clusterDirName, STATE_FILE_NAME), cloudstate, FILE_PERM_CLUSTER_STATE)
	return string(cloudstate), nil
}

func (obj *AwsProvider) GetSecretTokens(factory resources.StorageFactory) (map[string][]byte, error) {

	acesskeyid := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	return map[string][]byte{
		"aws_access_key_id":     []byte(acesskeyid),
		"aws_secret_access_key": []byte(secret),
	}, nil
}
