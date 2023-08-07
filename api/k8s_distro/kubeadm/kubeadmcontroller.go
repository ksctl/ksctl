package kubeadm

import (
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

// OTHER CONFIGURATIONS
type Instances struct {
	ControlPlanes []string
	WorkerPlanes  []string
	DataStores    []string
	Loadbalancer  string
}

type StateConfiguration struct {
	JoinControlToken string
	JoinWorkerToken  string
	SSHUser          string
	PublicIPs        Instances
	PrivateIPs       Instances
}

type KubeadmDistro struct {
	Version string
}

// ConfigureControlPlane implements resources.DistroFactory.
func (*KubeadmDistro) ConfigureControlPlane(noOfCP int, state resources.StorageFactory) error {
	panic("unimplemented")
}

// ConfigureDataStore implements resources.DistroFactory.
func (*KubeadmDistro) ConfigureDataStore(int, resources.StorageFactory) error {
	panic("unimplemented")
}

// ConfigureLoadbalancer implements resources.DistroFactory.
func (*KubeadmDistro) ConfigureLoadbalancer(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DestroyWorkerPlane implements resources.DistroFactory.
func (*KubeadmDistro) DestroyWorkerPlane(state resources.StorageFactory) error {
	panic("unimplemented")
}

// GetKubeConfig implements resources.DistroFactory.
func (*KubeadmDistro) GetKubeConfig(state resources.StorageFactory) (string, error) {
	panic("unimplemented")
}

// InitState implements resources.DistroFactory.
func (k8s *KubeadmDistro) InitState(cloud.CloudResourceState, resources.StorageFactory) {
	k8sState = &StateConfiguration{}
}

// InstallApplication implements resources.DistroFactory.
func (*KubeadmDistro) InstallApplication(state resources.StorageFactory) error {
	panic("unimplemented")
}

// JoinWorkerplane implements resources.DistroFactory.
func (*KubeadmDistro) JoinWorkerplane(int, resources.StorageFactory) error {
	panic("unimplemented")
}

var (
	k8sState *StateConfiguration
)

func ReturnKubeadmStruct() *KubeadmDistro {
	return &KubeadmDistro{}
}
