package k3s

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources/controllers/kubernetes"
)

type Instances struct {
	ControlPlanes []string
	WorkerPlanes  []string
	DataStores    []string
	Loadbalancer  string
}

type StateConfiguration struct {
	K3sToken  string
	SSHUser   string
	PublicIPs Instances
}

// type k3S struct {
// 	Metadata k3sInterface.K3sDistro
// 	State    StateConfiguration
// }

// configuration management
type K8sController kubernetes.ClientBuilder

// GetKubeconfig implements kubernetes.ControllerInterface.
func (b *K8sController) GetKubeconfig() (string, error) {
	fmt.Println("get kubeconfig k3s")
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
	k3s := (*K8sController)(b)
	return k3s
}
