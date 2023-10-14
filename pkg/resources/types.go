package resources

import (
	"os"

	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

type KsctlClient struct {
	// Cloud is the CloudProvider's factory interface
	Cloud CloudFactory

	// Distro is the Distrobution's factory interface
	Distro DistroFactory

	// Storage is the Storage's factory interface
	Storage StorageFactory

	// Metadata is used by the cloudController and manager to use data from cli
	Metadata Metadata
}

type Metadata struct {
	ClusterName   string          `json:"cluster_name"`
	Region        string          `json:"region"`
	Provider      KsctlCloud      `json:"cloud_provider"`
	K8sDistro     KsctlKubernetes `json:"kubernetes_distro"`
	K8sVersion    string          `json:"kubernetes_version"`
	StateLocation KsctlStore      `json:"storage_type"`
	IsHA          bool            `json:"ha_cluster"`

	ManagedNodeType      string `json:"node_type_managed"`
	WorkerPlaneNodeType  string `json:"node_type_workerplane"`
	ControlPlaneNodeType string `json:"node_type_controlplane"`
	DataStoreNodeType    string `json:"node_type_datastore"`
	LoadBalancerNodeType string `json:"node_type_loadbalancer"`

	NoMP int `json:"desired_no_of_managed_nodes"`      // No of managed Nodes
	NoWP int `json:"desired_no_of_workerplane_nodes"`  // No of woerkplane VMs
	NoCP int `json:"desired_no_of_controlplane_nodes"` // No of Controlplane VMs
	NoDS int `json:"desired_no_of_datastore_nodes"`    // No of DataStore VMs

	Applications string `json:"preinstalled_apps"`
	CNIPlugin    string `json:"cni_plugin"`
}

type CloudFactory interface {
	// NewVM create VirtualMachine with index for storing its state
	NewVM(StorageFactory, int) error

	// DelVM delete VirtualMachine with index for storing its state
	DelVM(StorageFactory, int) error

	// NewFirewall create Firewall
	NewFirewall(StorageFactory) error

	// DelFirewall delete Firewall
	DelFirewall(StorageFactory) error

	// NewNetwork create Network
	NewNetwork(StorageFactory) error

	// DelNetwork delete Network
	DelNetwork(StorageFactory) error

	// InitState is used to initalize the state of that partular cloud provider
	// its internal state and cloud provider's client
	// NOTE: multiple mode of OPERATIONS
	InitState(StorageFactory, KsctlOperation) error

	// CreateUploadSSHKeyPair create SSH keypair in the host machine and then upload pub key
	// and store the path of private key, username, etc.. wrt to specific cloud provider
	CreateUploadSSHKeyPair(StorageFactory) error

	// DelSSHKeyPair delete SSH keypair from the Cloud provider
	DelSSHKeyPair(StorageFactory) error

	// GetStateForHACluster used to get the state info for transfer it to kubernetes distro
	// for further configurations
	GetStateForHACluster(StorageFactory) (cloud.CloudResourceState, error)

	// NewManagedCluster creates managed kubernetes from cloud offering
	// it requires the no of nodes to be created
	NewManagedCluster(StorageFactory, int) error

	// DelManagedCluster deletes managed kubernetes from cloud offering
	DelManagedCluster(StorageFactory) error

	// Name sets the name for the resource you want to operate
	Name(string) CloudFactory

	// Role specify what is its role. Ex. Controlplane or WorkerPlane or DataStore...
	Role(KsctlRole) CloudFactory

	// VMType specifiy what is the VirtualMachine size to be used
	VMType(string) CloudFactory

	// Visibility whether to make the VM public or private
	Visibility(bool) CloudFactory

	// Application for the comma seperated apps names (Managed cluster)
	Application(string) bool

	// CNI for the CNI name (Managed cluster)
	CNI(string) (willBeInstalled bool)

	// Version for the Kubernetes Version (Managed cluster)
	Version(string) CloudFactory

	// NoOfWorkerPlane if setter is enabled it writes the new no of workerplane to be used
	// if getter is enabled it returns the current no of workerplane
	// its imp function for (shrinking, scaling)
	NoOfWorkerPlane(StorageFactory, int, bool) (int, error)

	// NoOfControlPlane Getter and setter
	// setter to store no of controlplane nodes
	// NOTE: it is meant to be used only for first time
	// it has no functionalit as (shrinking, scaling) if tried it will erase existing data
	NoOfControlPlane(int, bool) (int, error)

	// NoOfDataStore Getter and setter
	// setter to store no of datastore nodes
	// NOTE: it is meant to be used only for first time
	// it has no functionalit as (shrinking, scaling) if tried it will erase existing data
	NoOfDataStore(int, bool) (int, error)

	// GetHostNameAllWorkerNode it returns all the hostnames of workerplane nodes
	// it's used for the universal kubernetes for deletion of nodes which have to scale down
	GetHostNameAllWorkerNode() []string

	SwitchCluster(StorageFactory) error

	// GetStateFiles it returns the civo-state.json
	// WARN: sensitive info can be present
	GetStateFile(StorageFactory) (string, error)

	GetKubeconfigPath() string

	GetSecretTokens(StorageFactory) (map[string][]byte, error)
}

type KubernetesFactory interface {
	// InitState uses the cloud provider's shared state to initlaize itself
	// NOTE: multiple mode of OPERATIONS
	InitState(cloud.CloudResourceState, StorageFactory, KsctlOperation) error

	// ConfigureControlPlane to join or create VM as controlplane
	// it requires controlplane number
	ConfigureControlPlane(int, StorageFactory) error

	// JoinWorkerplane it joins to the existing cluster
	// it requires workerplane number
	JoinWorkerplane(int, StorageFactory) error

	//DestroyWorkerPlane(StorageFactory) ([]string, error)

	// ConfigureLoadbalancer it configures the Loadbalancer
	ConfigureLoadbalancer(StorageFactory) error

	// ConfigureDataStore it configure the datastore
	// it requires number
	ConfigureDataStore(int, StorageFactory) error

	// GetKubeConfig returns the path of kubeconfig
	GetKubeConfig(StorageFactory) (path string, data string, err error)

	// Version setter for version to be used
	Version(string) DistroFactory

	CNI(string) (externalCNI bool) // it will return error

	// GetStateFiles it returns the k8s-state.json
	// WARN: sensitive info can be present
	GetStateFile(StorageFactory) (string, error)
}

type DistroFactory interface {
	KubernetesFactory
	// NonKubernetesInfrastructure
}

type StorageFactory interface {
	// Save the data in bytes to specific location
	Save([]byte) error

	// TODO: check if required
	Destroy() error

	// Load gets contenets of file in bytes
	Load() ([]byte, error)

	// Path setter for path
	Path(string) StorageFactory

	// Permission setter for permission
	Permission(mode os.FileMode) StorageFactory

	// CreateDir creates directory
	CreateDir() error

	// DeleteDir deletes directories
	DeleteDir() error

	// GetFolders returns the folder's contents
	GetFolders() ([][]string, error)

	// Logger to access logger
	Logger() logger.LogFactory
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      KsctlClient
	Version     string
}
