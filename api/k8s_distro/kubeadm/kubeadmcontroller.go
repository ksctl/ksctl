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

// ConfigureControlPlane implements resources.Distributions.
func (*KubeadmDistro) ConfigureControlPlane(noOfCP int, state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// ConfigureDataStore implements resources.Distributions.
func (*KubeadmDistro) ConfigureDataStore(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// ConfigureLoadbalancer implements resources.Distributions.
func (*KubeadmDistro) ConfigureLoadbalancer(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// DestroyWorkerPlane implements resources.Distributions.
func (*KubeadmDistro) DestroyWorkerPlane(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// GetKubeConfig implements resources.Distributions.
func (*KubeadmDistro) GetKubeConfig(state resources.StateManagementInfrastructure) (string, error) {
	panic("unimplemented")
}

// InitState implements resources.Distributions.
func (k8s *KubeadmDistro) InitState(cloud.CloudResourceState) {
	k8sState = &StateConfiguration{}
}

// InstallApplication implements resources.Distributions.
func (*KubeadmDistro) InstallApplication(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// JoinWorkerplane implements resources.Distributions.
func (*KubeadmDistro) JoinWorkerplane(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

var (
	k8sState *StateConfiguration
)

func ReturnKubeadmStruct() *KubeadmDistro {
	return &KubeadmDistro{}
}
