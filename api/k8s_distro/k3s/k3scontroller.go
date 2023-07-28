package k3s

import "github.com/kubesimplify/ksctl/api/resources"

type Instances struct {
	ControlPlanes []string
	WorkerPlanes  []string
	DataStores    []string
	Loadbalancer  string
}

type StateConfiguration struct {
	K3sToken   string
	SSHUser    string
	PublicIPs  Instances
	PrivateIPs Instances
}

type K3sDistro struct {
	IsHA    bool
	Version string
}

// ConfigureControlPlane implements resources.Distributions.
func (*K3sDistro) ConfigureControlPlane(noOfCP int, state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// ConfigureDataStore implements resources.Distributions.
func (*K3sDistro) ConfigureDataStore(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// ConfigureLoadbalancer implements resources.Distributions.
func (*K3sDistro) ConfigureLoadbalancer(state resources.StateManagementInfrastructure) {
	k8sState = &StateConfiguration{}
	panic("unimplemented")
}

// DestroyWorkerPlane implements resources.Distributions.
func (*K3sDistro) DestroyWorkerPlane(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// GetKubeConfig implements resources.Distributions.
func (*K3sDistro) GetKubeConfig(state resources.StateManagementInfrastructure) (string, error) {
	panic("unimplemented")
}

// InitState implements resources.Distributions.
func (*K3sDistro) InitState(any) {
	// add the nil check here as well
}

// InstallApplication implements resources.Distributions.
func (*K3sDistro) InstallApplication(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// JoinWorkerplane implements resources.Distributions.
func (*K3sDistro) JoinWorkerplane(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// TODO: Add the SSH functionality here

var (
	k8sState *StateConfiguration
)

func ReturnK3sStruct() *K3sDistro {
	return &K3sDistro{}
}
