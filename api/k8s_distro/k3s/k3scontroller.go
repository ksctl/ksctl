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

// ConfigureControlPlane implements resources.DistroFactory.
func (*K3sDistro) ConfigureControlPlane(noOfCP int, state resources.StorageFactory) {
	fmt.Printf("[K3s] Configuring Controlplane[%v]....\n", noOfCP)
}

// ConfigureDataStore implements resources.DistroFactory.
func (*K3sDistro) ConfigureDataStore(state resources.StorageFactory) {
	fmt.Println("[K3s] Configuring DataStore....")
}

// ConfigureLoadbalancer implements resources.DistroFactory.
func (k8s *K3sDistro) ConfigureLoadbalancer(state resources.StorageFactory) {
	fmt.Println("[K3s] Configuring Loadbalancer....")
}

// DestroyWorkerPlane implements resources.DistroFactory.
func (*K3sDistro) DestroyWorkerPlane(state resources.StorageFactory) {
	panic("unimplemented")
}

// GetKubeConfig implements resources.DistroFactory.
func (*K3sDistro) GetKubeConfig(state resources.StorageFactory) (string, error) {
	fmt.Println("[K3s] Kubeconfig fetch....")
	return "{}", nil
}

// InitState implements resources.DistroFactory.
// try to achieve deepCopy
func (*K3sDistro) InitState(cloudState cloud.CloudResourceState) {
	// add the nil check here as well
	k8sState = &StateConfiguration{}
	k8sState.PublicIPs.ControlPlanes = cloudState.IPv4ControlPlanes
	//.....
	fmt.Println("[K3s] Initialized K3s from cloudprovider", k8sState)
}

// InstallApplication implements resources.DistroFactory.
func (*K3sDistro) InstallApplication(state resources.StorageFactory) {
	panic("unimplemented")
}

// JoinWorkerplane implements resources.DistroFactory.
func (*K3sDistro) JoinWorkerplane(state resources.StorageFactory) error {
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
