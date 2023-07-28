package k3s

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

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
	Version string
}

// ConfigureControlPlane implements resources.Distributions.
func (*K3sDistro) ConfigureControlPlane(noOfCP int, state resources.StateManagementInfrastructure) {
	fmt.Printf("[K3s] Configuring Controlplane[%v]....\n", noOfCP)
}

// ConfigureDataStore implements resources.Distributions.
func (*K3sDistro) ConfigureDataStore(state resources.StateManagementInfrastructure) {
	fmt.Println("[K3s] Configuring DataStore....")
}

// ConfigureLoadbalancer implements resources.Distributions.
func (k8s *K3sDistro) ConfigureLoadbalancer(state resources.StateManagementInfrastructure) {
	fmt.Println("[K3s] Configuring Loadbalancer....")
}

// DestroyWorkerPlane implements resources.Distributions.
func (*K3sDistro) DestroyWorkerPlane(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// GetKubeConfig implements resources.Distributions.
func (*K3sDistro) GetKubeConfig(state resources.StateManagementInfrastructure) (string, error) {
	fmt.Println("[K3s] Kubeconfig fetch....")
	return "{}", nil
}

// InitState implements resources.Distributions.
// try to achieve deepCopy
func (*K3sDistro) InitState(cloudState cloud.CloudResourceState) {
	// add the nil check here as well
	k8sState = &StateConfiguration{}
	k8sState.PublicIPs.ControlPlanes = cloudState.IPv4ControlPlanes
	//.....
	fmt.Println("[K3s] Initialized K3s from cloudprovider", k8sState)
}

// InstallApplication implements resources.Distributions.
func (*K3sDistro) InstallApplication(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// JoinWorkerplane implements resources.Distributions.
func (*K3sDistro) JoinWorkerplane(state resources.StateManagementInfrastructure) error {
	fmt.Println("[K3s] Adding WorkerPlane....")
	return nil
}

// TODO: Add the SSH functionality here

var (
	k8sState *StateConfiguration
)

func ReturnK3sStruct() *K3sDistro {
	return &K3sDistro{}
}
