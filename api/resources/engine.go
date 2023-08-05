package resources

import (
	"os"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type KsctlClient struct {
	Cloud  CloudInfrastructure
	Distro Distributions
	State  StateManagementInfrastructure
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
	NewVM(StateManagementInfrastructure) error
	DelVM(StateManagementInfrastructure) error

	NewFirewall(StateManagementInfrastructure) error
	DelFirewall(StateManagementInfrastructure) error

	NewNetwork(StateManagementInfrastructure) error
	DelNetwork(StateManagementInfrastructure) error

	InitState(StateManagementInfrastructure, string) error

	CreateUploadSSHKeyPair(StateManagementInfrastructure) error
	DelSSHKeyPair(StateManagementInfrastructure) error

	// get the state required for the kubernetes dributions to configure
	GetStateForHACluster(StateManagementInfrastructure) (cloud.CloudResourceState, error)

	NewManagedCluster(StateManagementInfrastructure) error
	DelManagedCluster(StateManagementInfrastructure) error
	GetManagedKubernetes(StateManagementInfrastructure)

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
	ConfigureControlPlane(int, StateManagementInfrastructure)
	// DestroyControlPlane(StateManagementInfrastructure)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	JoinWorkerplane(StateManagementInfrastructure) error
	DestroyWorkerPlane(StateManagementInfrastructure)

	ConfigureLoadbalancer(StateManagementInfrastructure)
	// DestroyLoadbalancer(StateManagementInfrastructure)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	ConfigureDataStore(StateManagementInfrastructure)
	// DestroyDataStore(StateManagementInfrastructure)  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	InstallApplication(StateManagementInfrastructure)

	GetKubeConfig(StateManagementInfrastructure) (string, error)
}

// FEATURE: non kubernetes distrobutions like nomad
// type NonKubernetesInfrastructure interface {
// 	InstallApplications()
// }

type Distributions interface {
	KubernetesInfrastructure
	// NonKubernetesInfrastructure
}

type StateManagementInfrastructure interface {
	Save([]byte) error
	Destroy() error
	Load() ([]byte, error) // try to make the return type defined

	// for modifier
	Path(string) StateManagementInfrastructure
	Permission(mode os.FileMode) StateManagementInfrastructure
	CreateDir() error
	DeleteDir() error

	// to access logger
	Logger() logger.LogFactory
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      KsctlClient
	Version     string
}
