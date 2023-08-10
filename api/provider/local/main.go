package local

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type StateConfiguration struct {
	ClusterName string `json:"cluster_name"`
}

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool

	// purpose: application in managed cluster
	Apps string
	Cni  string
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	// Spec        Machine `json:"spec"`
	Metadata
}

func (*LocalProvider) GetHostNameAllWorkerNode() []string {
	//TODO implement me
	panic("[local] unsupported operation")
}

// Version implements resources.CloudFactory.
func (*LocalProvider) Version(string) resources.CloudFactory {
	panic("[local] unsupported operation")
}

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (*LocalProvider) CreateUploadSSHKeyPair(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// DelFirewall implements resources.CloudFactory.
func (*LocalProvider) DelFirewall(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// DelManagedCluster implements resources.CloudFactory.
func (*LocalProvider) DelManagedCluster(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// DelNetwork implements resources.CloudFactory.
func (*LocalProvider) DelNetwork(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// DelSSHKeyPair implements resources.CloudFactory.
func (*LocalProvider) DelSSHKeyPair(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// DelVM implements resources.CloudFactory.
func (*LocalProvider) DelVM(resources.StorageFactory, int) error {
	panic("[local] unsupported operation")
}

// GetManagedKubernetes implements resources.CloudFactory.
func (*LocalProvider) GetManagedKubernetes(state resources.StorageFactory) {
	panic("[local] unsupported operation")
}

// GetStateForHACluster implements resources.CloudFactory.
func (*LocalProvider) GetStateForHACluster(state resources.StorageFactory) (cloud.CloudResourceState, error) {
	panic("[local] unsupported operation")
}

// InitState implements resources.CloudFactory.
func (*LocalProvider) InitState(state resources.StorageFactory, operation string) error {
	panic("[local] unsupported operation")
}

// NewFirewall implements resources.CloudFactory.
func (*LocalProvider) NewFirewall(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// NewManagedCluster implements resources.CloudFactory.
func (*LocalProvider) NewManagedCluster(state resources.StorageFactory, noOfNodes int) error {
	panic("[local] unsupported operation")
}

// NewNetwork implements resources.CloudFactory.
func (*LocalProvider) NewNetwork(state resources.StorageFactory) error {
	panic("[local] unsupported operation")
}

// NewVM implements resources.CloudFactory.
func (*LocalProvider) NewVM(resources.StorageFactory, int) error {
	panic("[local] unsupported operation")
}

func ReturnLocalStruct(metadata resources.Metadata) *LocalProvider {
	return &LocalProvider{
		ClusterName: metadata.ClusterName,
	}
}

// it will contain the name of the resource to be created
func (cloud *LocalProvider) Name(resName string) resources.CloudFactory {
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *LocalProvider) Role(resRole string) resources.CloudFactory {
	cloud.Metadata.Role = resRole
	return cloud
}

// it will contain which vmType to create
func (cloud *LocalProvider) VMType(size string) resources.CloudFactory {
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *LocalProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.Metadata.Public = toBePublic
	return cloud
}

// if its ha its always false instead it tells whether the provider has support in their managed offerering
func (cloud *LocalProvider) SupportForApplications() bool {
	return false
}

func (cloud *LocalProvider) SupportForCNI() bool {
	return false
}

func (client *LocalProvider) Application(s string) resources.CloudFactory {
	client.Metadata.Apps = s
	return client
}

func (client *LocalProvider) CNI(s string) resources.CloudFactory {
	client.Metadata.Cni = s
	return client
}

// NoOfControlPlane implements resources.CloudFactory.
func (obj *LocalProvider) NoOfControlPlane(int, bool) (int, error) {
	return -1, fmt.Errorf("[local] unsupported operation")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *LocalProvider) NoOfDataStore(int, bool) (int, error) {
	return -1, fmt.Errorf("[local] unsupported operation")
}

// NoOfWorkerPlane implements resources.CloudFactory.
func (obj *LocalProvider) NoOfWorkerPlane(resources.StorageFactory, int, bool) (int, error) {
	return -1, fmt.Errorf("[local] unsupported operation")
}
