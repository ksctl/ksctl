package resources

import (
	"os"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type KsctlClient struct {
	Cloud   CloudInfrastructure
	Distro  Distributions
	Storage StorageInfrastructure
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

	NoWP int // No of woerkplane VMs
	NoCP int // No of Controlplane VMs
	NoDS int // No of DataStore VMs

	Applications []string `json:"application"`
	CNIPlugin    *string  `json:"cni_plugin"`
}

type CloudInfrastructure interface {
	NewVM(StorageInfrastructure) error
	DelVM(StorageInfrastructure) error

	NewFirewall(StorageInfrastructure) error
	DelFirewall(StorageInfrastructure) error

	NewNetwork(StorageInfrastructure) error
	DelNetwork(StorageInfrastructure) error

	InitState(StorageInfrastructure, string) error

	CreateUploadSSHKeyPair(StorageInfrastructure) error
	DelSSHKeyPair(StorageInfrastructure) error

	// get the state required for the kubernetes dributions to configure
	GetStateForHACluster(StorageInfrastructure) (cloud.CloudResourceState, error)

	NewManagedCluster(StorageInfrastructure) error
	DelManagedCluster(StorageInfrastructure) error
	GetManagedKubernetes(StorageInfrastructure)

	// used by controller
	Name(string) CloudInfrastructure
	Role(string) CloudInfrastructure
	VMType(string) CloudInfrastructure
	Visibility(bool) CloudInfrastructure

	SupportForApplications() bool
	SupportForCNI() bool
}

type KubernetesInfrastructure interface {
	InitState(cloud.CloudResourceState)

	// it recieves no of controlplane to which we want to configure
	// NOTE: make the first controlplane return server token as possible
	ConfigureControlPlane(int, StorageInfrastructure)
	// DestroyControlPlane(StorageInfrastructure)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	JoinWorkerplane(StorageInfrastructure) error
	DestroyWorkerPlane(StorageInfrastructure)

	ConfigureLoadbalancer(StorageInfrastructure)
	// DestroyLoadbalancer(StorageInfrastructure)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	ConfigureDataStore(StorageInfrastructure)
	// DestroyDataStore(StorageInfrastructure)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	InstallApplication(StorageInfrastructure)

	GetKubeConfig(StorageInfrastructure) (string, error)
}

// FEATURE: non kubernetes distrobutions like nomad
// type NonKubernetesInfrastructure interface {
// 	InstallApplications()
// }

type Distributions interface {
	KubernetesInfrastructure
	// NonKubernetesInfrastructure
}

type StorageInfrastructure interface {
	Save([]byte) error
	Destroy() error
	Load() ([]byte, error) // try to make the return type defined

	// for modifier
	Path(string) StorageInfrastructure
	Permission(mode os.FileMode) StorageInfrastructure
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
