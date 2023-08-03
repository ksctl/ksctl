package resources

import (
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"os"
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
	NoWP          int8 // NO of woerkplane VMs
	NoCP          int8 // No of Controlplane VMs
	NoDS          int8 // No of DataStore VMs
}

// NOTE: local cluster are also supported but with feature flags only managedcluster available
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
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      KsctlClient
	Version     string
}
