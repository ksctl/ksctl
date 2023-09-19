package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/kubesimplify/ksctl/api/resources"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// var (
// 	awsCloudState  *StateConfiguration
// 	clusterDirName string
// 	clusterType    string
// 	ctx            context.Context
// )

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

	client  AwsGo
	SSHPath string `json:"ssh_key"`
}

type AWSStateVms struct {
	Names                 []string `json:"names"`
	SecurityGroupName     string   `json:"network_security_group_name"`
	SecurityGroupID       string   `json:"network_security_group_id"`
	DiskNames             []string `json:"disk_name"`
	PublicIPNames         []string `json:"public_ip_"`
	PrivateIPs            []string `json:"private_ip"`
	PublicIPs             []string `json:"public_ip"`
	NetworkInterfaceNames []string `json:"network_interface_name"`
	NetworkInterfaceIDs   []string `json:"network_interface_id"`
	SubnetNames           []string `json:"subnet_name"`
	SubnetIDs             []string `json:"subnet_id"`
	Hostnames             []string `json:"hostname"`
	PublicIPIDs           []string `json:"publicipids"`
}

var (
	azureCloudState *StateConfiguration

	clusterDirName string
	clusterType    string // it stores the ha or managed

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
	IsCompleted        bool
	ClusterName        string `json:"cluster_name"`
	Region             string `json:"region"`
	VPC                string `json:"vpc"`
	SSHUSER            string `json:"ssh_user"`
	SSHPrivateKeyLoc   string `json:"ssh_private_key_loc"`
	SSHKeyName         string `json:"ssh_key_name"`
	ManagedClusterName string `json:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_nodes"`
	SubnetName         string `json:"subnet_name"`
	SubnetID           string `json:"subnet_id"`
	SecurityGroupName  string `json:"security_group_name"`
	SecurityGroupID    string `json:"security_group_id"`
	GatewayName        string `json:"gateway_name"`
	GatewayID          string `json:"gateway_id"`
	RouteTableName     string `json:"route_table_name"`
	RouteTableID       string `json:"route_table_id"`

	InfoControlPlanes AWSStateVms `json:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database"`
	InfoLoadBalancer  AWSStateVm  `json:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
}

type metadata struct {
	resName string
	role    string
	vmType  string
	public  bool

	apps    string
	cni     string
	version string

	noCP int
	noWP int
	noDS int

	k8sName    string
	k8sVersion string
}

func ReturnAwsStruct(Metadata resources.Metadata, ClientOption func() *interface{}) (resources.CloudFactory, error) {
	return &AwsProvider{
		clusterName: Metadata.ClusterName,
		region:      "ap-south-1",
		haCluster:   Metadata.IsHA,
		metadata: metadata{
			k8sVersion: Metadata.K8sVersion,
			k8sName:    Metadata.K8sDistro,
		},

		session: NEWCLIENT(),
		client:  ClientOption(),
	}, nil
}

func NEWCLIENT() aws.Config {
	NewSession, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"),
		config.WithSharedConfigProfile("default"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dsvvvsvrvsvrv", "vsavveawg4gave4ds4ees4ge", "")),
	)
	if err != nil {
		panic(err)
	}

	// obj.Session = &NewSession
	fmt.Println("AWS Session Created Successfully")

	return NewSession

}

func (obj *AwsProvider) Name(resName string) resources.CloudFactory {

	if err := utils.IsValidName(resName); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}

	obj.metadata.resName = resName
	fmt.Println(obj.metadata.resName)
	fmt.Println("named the resource successfully")
	return nil
}

func (obj *AwsProvider) DelVM(factory resources.StorageFactory, i int) error {
	//TODO implement me
	fmt.Println("AWS Del VM")
	return nil
}

func (obj *AwsProvider) NewFirewall(factory resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS New Firewall")
	return nil
}

func (obj *AwsProvider) DelFirewall(factory resources.StorageFactory) error {
	//TODO implement me
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
	path := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AWS, clusterType, clusterDirName, STATE_FILE_NAME)
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	return convertStateFromBytes(raw)
}

func (obj *AwsProvider) InitState(storage resources.StorageFactory, opration string) error {

	switch obj.haCluster {
	case false:
		clusterType = utils.CLUSTER_TYPE_MANG
	case true:
		clusterType = utils.CLUSTER_TYPE_HA

	}
	obj.vpc = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.clusterName, clusterType)
	clusterDirName = obj.clusterName + " " + obj.vpc + " " + obj.region

	errLoadState := loadStateHelper(storage)
	switch opration {
	case utils.OPERATION_STATE_CREATE:
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
				KubernetesDistro: obj.metadata.k8sName,
				KubernetesVer:    obj.metadata.k8sVersion,
			}
		}

	case utils.OPERATION_STATE_DELETE:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[aws] Delete resource(s)")

	case utils.OPERATION_STATE_GET:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[aws] Get resources")
		clusterDirName = awsCloudState.ClusterName + " " + awsCloudState.VPC + " " + awsCloudState.Region
	default:
		return errors.New("[aws] Invalid operation for init state")
	}

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

func (obj *AwsProvider) InitClient(storage resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Init Client")
	return nil

}

func (obj *AwsProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Create Upload SSH Key Pair")
	return nil

}

func (obj *AwsProvider) DelSSHKeyPair(storage resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Del SSH Key Pair")
	return nil

}

func (obj *AwsProvider) GetStateForHACluster(factory resources.StorageFactory) (cloud.CloudResourceState, error) {
	//TODO implement me
	fmt.Println("AWS Get State for HA Cluster")

	str := cloud.CloudResourceState{}
	return str, nil

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

func (obj *AwsProvider) Role(s string) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS Role")
	return nil

}

func (obj *AwsProvider) VMType(s string) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS VM Type")
	return nil

}

func (obj *AwsProvider) Visibility(b bool) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS Visibility")
	return nil

}

func (obj *AwsProvider) SupportForApplications() bool {
	//TODO implement me
	fmt.Println("AWS Support for Applications")
	return true

}

func (obj *AwsProvider) SupportForCNI() bool {
	//TODO implement me
	fmt.Println("AWS Support for CNI")
	return true

}

func (obj *AwsProvider) Application(s string) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS Application")
	return nil

}

func (obj *AwsProvider) CNI(s string) resources.CloudFactory {
	//TODO implement me
	fmt.Println("AWS CNI")
	return nil

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
