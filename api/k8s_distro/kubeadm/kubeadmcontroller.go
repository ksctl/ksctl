package kubeadm

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/resources/controllers/kubernetes"
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

var (
	k8sState *StateConfiguration
)

// configuration management
type K8sController kubernetes.ClientBuilder

// TODO: HydrateCloudState implements kubernetes.ControllerInterface.
func (*K8sController) HydrateCloudState(state cloud.CloudResourceState) {
	k8sState = &StateConfiguration{}
	k8sState.PublicIPs.ControlPlanes = state.IPv4ControlPlanes
	k8sState.PrivateIPs.ControlPlanes = state.PrivateIPv4ControlPlanes
}

// GetKubeconfig implements kubernetes.ControllerInterface.
func (b *K8sController) GetKubeconfig() (string, error) {
	fmt.Println("get kubeconfig kubeadm")
	b.Distro.ConfigureControlPlane()
	return "", nil
}

// GetServerToken implements kubernetes.ControllerInterface.
func (b *K8sController) GetServerToken() (string, error) {
	panic("unimplemented")
}

// InitializeMasterControlPlane implements kubernetes.ControllerInterface.
func (b *K8sController) InitializeMasterControlPlane() error {
	panic("unimplemented")
}

// JoinControlplane implements kubernetes.ControllerInterface.
func (b *K8sController) JoinControlplane() (string, error) {
	panic("unimplemented")
}

// JoinDatastore implements kubernetes.ControllerInterface.
func (b *K8sController) JoinDatastore() (string, error) {
	panic("unimplemented")
}

// JoinWorkerplane implements kubernetes.ControllerInterface.
func (b *K8sController) JoinWorkerplane() (string, error) {
	panic("unimplemented")
}

// SetupDatastore implements kubernetes.ControllerInterface.
func (b *K8sController) SetupDatastore() (string, error) {
	panic("unimplemented")
}

// SetupLoadBalancer implements kubernetes.ControllerInterface.
func (b *K8sController) SetupLoadBalancer() {
	panic("unimplemented")
}

// SetupWorkerplane implements kubernetes.ControllerInterface.
func (b *K8sController) SetupWorkerplane() (string, error) {
	panic("unimplemented")
}

func WrapK8sControllerBuilder(b *kubernetes.ClientBuilder) *K8sController {
	k8s := (*K8sController)(b)
	return k8s
}
