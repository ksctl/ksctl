package resources

import (
	"os"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type KsctlClient struct {
	// Cloud is the CloudProvider's factory interface
	Cloud CloudFactory

	// Distro is the Distrobution's factory interface
	Distro DistroFactory

	// Storage is the Storage's factory interface
	Storage StorageFactory

	// Metadata is used by the cloudController and manager to use data from cli
	Metadata
}

type Metadata struct {
	ClusterName   string
	Region        string
	Provider      string
	K8sDistro     string
	K8sVersion    string
	StateLocation string
	IsHA          bool

	// TODO: is it required?
	// try to see if string could be replaced by pointer to reduce memory
	ManagedNodeType      string
	WorkerPlaneNodeType  string
	ControlPlaneNodeType string
	DataStoreNodeType    string
	LoadBalancerNodeType string

	NoMP int // No of managed Nodes

	NoWP int // No of woerkplane VMs
	NoCP int // No of Controlplane VMs
	NoDS int // No of DataStore VMs

	Applications string `json:"application"`
	CNIPlugin    string `json:"cni_plugin"`
}

type CloudFactory interface {
	NewVM(StorageFactory, int) error
	DelVM(StorageFactory, int) error

	NewFirewall(StorageFactory) error
	DelFirewall(StorageFactory) error

	NewNetwork(StorageFactory) error
	DelNetwork(StorageFactory) error

	InitState(StorageFactory, string) error

	CreateUploadSSHKeyPair(StorageFactory) error
	DelSSHKeyPair(StorageFactory) error

	// get the state required for the kubernetes dributions to configure
	GetStateForHACluster(StorageFactory) (cloud.CloudResourceState, error)

	NewManagedCluster(StorageFactory, int) error
	DelManagedCluster(StorageFactory) error
	GetManagedKubernetes(StorageFactory)

	// used by controller
	Name(string) CloudFactory
	Role(string) CloudFactory
	VMType(string) CloudFactory
	Visibility(bool) CloudFactory

	SupportForApplications() bool
	SupportForCNI() bool

	// these are meant to be used for managed clusters
	Application(string) CloudFactory
	CNI(string) CloudFactory
	Version(string) CloudFactory

	// for the state of instances created and destroyed
	NoOfWorkerPlane(int, bool) (int, error)
	NoOfControlPlane(int, bool) (int, error)
	NoOfDataStore(int, bool) (int, error)
}

type KubernetesFactory interface {
	InitState(cloud.CloudResourceState, StorageFactory)

	// it recieves no of controlplane to which we want to configure
	// NOTE: make the first controlplane return server token as possible
	ConfigureControlPlane(int, StorageFactory) error
	// DestroyControlPlane(StorageFactory)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	JoinWorkerplane(int, StorageFactory) error
	DestroyWorkerPlane(StorageFactory) error

	ConfigureLoadbalancer(StorageFactory) error
	// DestroyLoadbalancer(StorageFactory)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	ConfigureDataStore(int, StorageFactory) error
	// DestroyDataStore(StorageFactory)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	// meant to be used for the HA clusters
	InstallApplication(StorageFactory) error

	//GetKubeConfig(StorageFactory) (string, error)
}

// FEATURE: non kubernetes distrobutions like nomad
// type NonKubernetesInfrastructure interface {
// 	InstallApplications()
// }

type DistroFactory interface {
	KubernetesFactory
	// NonKubernetesInfrastructure
}

type StorageFactory interface {
	Save([]byte) error
	Destroy() error
	Load() ([]byte, error) // try to make the return type defined

	// for modifier
	Path(string) StorageFactory
	Permission(mode os.FileMode) StorageFactory
	CreateDir() error
	DeleteDir() error
	GetFolders() ([][]string, error)

	// to access logger
	Logger() logger.LogFactory
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      KsctlClient
	Version     string
}
