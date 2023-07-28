package resources

type KsctlClient struct {
	Cloud    CloudInfrastructure
	Distro   Distributions
	State    StateManagementInfrastructure
	Metadata Metadata
}

type Metadata struct {
	ClusterName   string
	Region        string
	Provider      string
	K8sDistro     string
	K8sVersion    string
	StateLocation string
	IsHA          bool
}

// NOTE: local cluster are also supported but with feature flags only managedcluster available
type CloudInfrastructure interface {
	NewVM() error
	DelVM() error

	NewFirewall() error
	DelFirewall() error

	NewNetwork() error
	DelNetwork() error

	InitState() error

	CreateUploadSSHKeyPair() error
	DelSSHKeyPair() error

	// get the state required for the kubernetes dributions to configure
	GetStateForHACluster() (any, error)

	NewManagedCluster() error
	DelManagedCluster() error
	GetManagedKubernetes()
}

type KubernetesInfrastructure interface {
	InitState(any)

	// it recieves no of controlplane to which we want to configure
	// NOTE: make the first controlplane return server token as possible
	ConfigureControlPlane(int)
	// DestroyControlPlane()  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	JoinWorkerplane() error
	DestroyWorkerPlane()

	ConfigureLoadbalancer()
	// DestroyLoadbalancer()  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	ConfigureDataStore()
	// DestroyDataStore()  // NOTE: [FEATURE] destroy not available
	// only able to remove the VirtualMachine

	InstallApplication()

	GetKubeConfig() (string, error)
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
	Save(string, any) error
	Load(string) (any, error) // try to make the return type defined
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      KsctlClient
	Version     string
}
