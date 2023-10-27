package resources

import (
	"github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

type KubernetesFactory interface {
	// InitState uses the cloud provider's shared state to initlaize itself
	// NOTE: multiple mode of OPERATIONS
	InitState(cloud.CloudResourceState, StorageFactory, consts.KsctlOperation) error

	// ConfigureControlPlane to join or create VM as controlplane
	// it requires controlplane number
	ConfigureControlPlane(int, StorageFactory) error

	// JoinWorkerplane it joins to the existing cluster
	// it requires workerplane number
	JoinWorkerplane(int, StorageFactory) error

	//DestroyWorkerPlane(StorageFactory) ([]string, error)

	// ConfigureLoadbalancer it configures the Loadbalancer
	ConfigureLoadbalancer(StorageFactory) error

	// ConfigureDataStore it configure the datastore
	// it requires number
	ConfigureDataStore(int, StorageFactory) error

	// GetKubeConfig returns the path of kubeconfig
	GetKubeConfig(StorageFactory) (path string, data string, err error)

	// Version setter for version to be used
	Version(string) DistroFactory

	CNI(string) (externalCNI bool) // it will return error

	// GetStateFiles it returns the k8s-state.json
	// WARN: sensitive info can be present
	GetStateFile(StorageFactory) (string, error)
}

type DistroFactory interface {
	KubernetesFactory
	// NonKubernetesInfrastructure
}
