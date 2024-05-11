package types

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

// add context support for each
// fo rthat we can add timeout and other context related things
// so plan is it will use the already existing ${cloud}Ctx variable
// and it will create a temporary context when it will be performing the task

type CloudFactory interface {
	// NewVM create VirtualMachine with index for storing its state
	NewVM(StorageFactory, int) error

	// DelVM delete VirtualMachine with index for storing its state
	DelVM(StorageFactory, int) error

	// NewFirewall create Firewall
	NewFirewall(StorageFactory) error

	// DelFirewall delete Firewall
	DelFirewall(StorageFactory) error

	// NewNetwork create Network
	NewNetwork(StorageFactory) error

	// DelNetwork delete Network
	DelNetwork(StorageFactory) error

	// Credential
	Credential(StorageFactory) error

	// InitState is used to initalize the state of that partular cloud provider
	// its internal state and cloud provider's client
	// NOTE: multiple mode of OPERATIONS
	InitState(StorageFactory, consts.KsctlOperation) error

	// CreateUploadSSHKeyPair create SSH keypair in the host machine and then upload pub key
	// and store the path of private key, username, etc.. wrt to specific cloud provider
	CreateUploadSSHKeyPair(StorageFactory) error

	// DelSSHKeyPair delete SSH keypair from the Cloud provider
	DelSSHKeyPair(StorageFactory) error

	// GetStateForHACluster used to get the state info for transfer it to kubernetes distro
	// for further configurations
	GetStateForHACluster(StorageFactory) (cloud.CloudResourceState, error)

	// NewManagedCluster creates managed kubernetes from cloud offering
	// it requires the no of nodes to be created
	NewManagedCluster(StorageFactory, int) error

	// DelManagedCluster deletes managed kubernetes from cloud offering
	DelManagedCluster(StorageFactory) error

	// Name sets the name for the resource you want to operate
	Name(string) CloudFactory // TODO: add the error passing!

	// Role specify what is its role. Ex. Controlplane or WorkerPlane or DataStore...
	Role(consts.KsctlRole) CloudFactory // TODO: add the error passing!

	// VMType specifiy what is the VirtualMachine size to be used
	VMType(string) CloudFactory // TODO: add the error passing!

	// Visibility whether to make the VM public or private
	Visibility(bool) CloudFactory

	// Application for the comma seperated apps names (Managed cluster)
	Application([]string) bool

	// CNI for the CNI name (Managed cluster)
	CNI(string) (willBeInstalled bool)

	// Version for the Kubernetes Version (Managed cluster)
	Version(string) CloudFactory // TODO: rename to ManagedK8sVersion and add error passing

	// NoOfWorkerPlane if setter is enabled it writes the new no of workerplane to be used
	// if getter is enabled it returns the current no of workerplane
	// its imp function for (shrinking, scaling)
	NoOfWorkerPlane(StorageFactory, int, bool) (int, error)

	// NoOfControlPlane Getter and setter
	// setter to store no of controlplane nodes
	// NOTE: it is meant to be used only for first time
	// it has no functionalit as (shrinking, scaling) if tried it will erase existing data
	NoOfControlPlane(int, bool) (int, error)

	// NoOfDataStore Getter and setter
	// setter to store no of datastore nodes
	// NOTE: it is meant to be used only for first time
	// it has no functionalit as (shrinking, scaling) if tried it will erase existing data
	NoOfDataStore(int, bool) (int, error)

	// GetHostNameAllWorkerNode it returns all the hostnames of workerplane nodes
	// it's used for the universal kubernetes for deletion of nodes which have to scale down
	GetHostNameAllWorkerNode() []string

	IsPresent(StorageFactory) error
}
