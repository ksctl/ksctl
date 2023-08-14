package local

import (
	"encoding/json"
	"fmt"

	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	cloud_control_res "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type StateConfiguration struct {
	ClusterName string `json:"cluster_name"`
	Distro      string `json:"distro"`
	Version     string `json:"version"`
	Nodes       int    `json:"nodes"`
}

type Metadata struct {
	ResName string
	Version string

	// purpose: application in managed cluster
	Apps string
	Cni  string
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	NoNodes     int    `json:"no_nodes"`
	Metadata
}

var (
	localState *StateConfiguration
)

const (
	STATE_FILE = "kind-state.json"
	KUBECONFIG = "kubeconfig"
)

func ReturnLocalStruct(metadata resources.Metadata) (*LocalProvider, error) {
	return &LocalProvider{
		ClusterName: metadata.ClusterName,
	}, nil
}

// InitState implements resources.CloudFactory.
func (cloud *LocalProvider) InitState(storage resources.StorageFactory, operation string) error {
	switch operation {
	case utils.OPERATION_STATE_CREATE:
		if isPresent(storage, cloud.ClusterName) {
			return fmt.Errorf("[local] already present")
		}
		localState = &StateConfiguration{
			ClusterName: cloud.ClusterName,
			Distro:      "kind",
		}
		var err error
		err = storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, cloud.ClusterName)).
			Permission(0750).CreateDir()
		if err != nil {
			return err
		}

		err = saveStateHelper(storage, utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, cloud.ClusterName, STATE_FILE))
		if err != nil {
			return err
		}
	case utils.OPERATION_STATE_DELETE, utils.OPERATION_STATE_GET:
		err := loadStateHelper(storage, utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, cloud.ClusterName, STATE_FILE))
		if err != nil {
			return err
		}
	}
	storage.Logger().Success("[local] initialized the state")
	return nil
}

// it will contain the name of the resource to be created
func (cloud *LocalProvider) Name(resName string) resources.CloudFactory {
	cloud.Metadata.ResName = resName
	return cloud
}

// if its ha its always false instead it tells whether the provider has support in their managed offerering
func (cloud *LocalProvider) SupportForApplications() bool {
	return false
}

func (cloud *LocalProvider) SupportForCNI() bool {
	return false
}

func (cloud *LocalProvider) Application(s string) resources.CloudFactory {
	cloud.Metadata.Apps = s
	return cloud
}

func (cloud *LocalProvider) CNI(s string) resources.CloudFactory {
	cloud.Metadata.Cni = s
	return cloud
}

// Version implements resources.CloudFactory.
func (cloud *LocalProvider) Version(ver string) resources.CloudFactory {
	// TODO: validation of version
	cloud.Metadata.Version = ver
	return cloud
}

func GetRAWClusterInfos(storage resources.StorageFactory) ([]cloud_control_res.AllClusterData, error) {
	var data []cloud_control_res.AllClusterData

	managedFolders, err := storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL)).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, folder := range managedFolders {

		path := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, folder[0], STATE_FILE)
		raw, err := storage.Path(path).Load()
		if err != nil {
			return nil, err
		}
		var clusterState *StateConfiguration
		if err := json.Unmarshal(raw, &clusterState); err != nil {
			return nil, err
		}

		data = append(data,
			cloud_control_res.AllClusterData{
				Provider:   utils.CLOUD_LOCAL,
				Name:       folder[0],
				Region:     "N/A",
				Type:       utils.CLUSTER_TYPE_MANG,
				K8sDistro:  clusterState.Distro,
				K8sVersion: clusterState.Version,
				NoMgt:      clusterState.Nodes,
			})
	}
	return data, nil
}

// //// NOT IMPLEMENTED //////

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *LocalProvider) Role(string) resources.CloudFactory {
	return nil
}

// it will contain which vmType to create
func (cloud *LocalProvider) VMType(string) resources.CloudFactory {
	return nil
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *LocalProvider) Visibility(bool) resources.CloudFactory {
	return nil
}

func (*LocalProvider) GetHostNameAllWorkerNode() []string {
	//TODO implement me
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

// NewFirewall implements resources.CloudFactory.
func (*LocalProvider) NewFirewall(state resources.StorageFactory) error {
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

// NoOfControlPlane implements resources.CloudFactory.
func (cloud *LocalProvider) NoOfControlPlane(int, bool) (int, error) {
	return -1, fmt.Errorf("[local] unsupported operation")
}

// NoOfDataStore implements resources.CloudFactory.
func (cloud *LocalProvider) NoOfDataStore(int, bool) (int, error) {
	return -1, fmt.Errorf("[local] unsupported operation")
}

// NoOfWorkerPlane implements resources.CloudFactory.
func (cloud *LocalProvider) NoOfWorkerPlane(resources.StorageFactory, int, bool) (int, error) {
	return -1, fmt.Errorf("[local] unsupported operation")
}

func (obj *LocalProvider) SwitchCluster(storage resources.StorageFactory) error {

	if isPresent(storage, obj.ClusterName) {
		printKubeconfig(storage, utils.OPERATION_STATE_CREATE)
		return nil
	}
	return fmt.Errorf("[local] Cluster not found")
}
