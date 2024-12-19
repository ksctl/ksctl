package providers

import "github.com/ksctl/ksctl/pkg/consts"

type Cloud interface {
	NewVM(int) error

	DelVM(int) error

	NewFirewall() error

	DelFirewall() error

	NewNetwork() error

	DelNetwork() error

	Credential() error

	InitState(consts.KsctlOperation) error

	CreateUploadSSHKeyPair() error

	DelSSHKeyPair() error

	GetStateForHACluster() (cloud.CloudResourceState, error)

	NewManagedCluster(int) error

	DelManagedCluster() error

	GetRAWClusterInfos() ([]cloud.AllClusterData, error)

	Name(string) Cloud

	Role(consts.KsctlRole) Cloud

	VMType(string) Cloud

	Visibility(bool) Cloud

	Application([]string) bool

	CNI(string) (willBeInstalled bool)

	ManagedK8sVersion(string) Cloud

	NoOfWorkerPlane(int, bool) (int, error)

	NoOfControlPlane(int, bool) (int, error)

	NoOfDataStore(int, bool) (int, error)

	GetHostNameAllWorkerNode() []string

	IsPresent() error

	GetKubeconfig() (*string, error)
}
