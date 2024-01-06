package kubeadm

import (
	"github.com/kubesimplify/ksctl/internal/storage/types"
	. "github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
)

type KubeadmDistro struct {
	KubeadmVer string
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

// InitState implements resources.DistroFactory.
func (k8s *KubeadmDistro) InitState(cloud.CloudResourceState, resources.StorageFactory, KsctlOperation) error {
	return nil
}

// JoinWorkerplane implements resources.DistroFactory.
func (*KubeadmDistro) JoinWorkerplane(int, resources.StorageFactory) error {
	panic("unimplemented")
}

var (
	k8sState *types.StorageDocument
)

func ReturnKubeadmStruct(resources.Metadata) resources.DistroFactory {
	return &KubeadmDistro{}
}

func (kubeadm *KubeadmDistro) Version(string) resources.DistroFactory {
	// TODO: Implement
	return kubeadm
}

func (kubeadm *KubeadmDistro) CNI(cni string) (externalCNI bool) {
	return true
}
