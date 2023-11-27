package local

import (
	"encoding/json"
	"fmt"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/kubesimplify/ksctl/pkg/helpers"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	cloudControlRes "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
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

	Cni string
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	NoNodes     int    `json:"no_nodes"`
	Metadata

	client LocalGo
}

var (
	localState *StateConfiguration
	log        resources.LoggerFactory
)

// GetSecretTokens implements resources.CloudFactory.
func (*LocalProvider) GetSecretTokens(resources.StorageFactory) (map[string][]byte, error) {
	return nil, nil
}

// GetStateFile implements resources.CloudFactory.
func (*LocalProvider) GetStateFile(resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(localState)
	if err != nil {
		return "", err
	}
	log.Debug("Printing", "cloudState", cloudstate)
	return string(cloudstate), nil
}

const (
	STATE_FILE = "kind-state.json"
	KUBECONFIG = "kubeconfig"
)

func ReturnLocalStruct(metadata resources.Metadata, ClientOption func() LocalGo) (*LocalProvider, error) {
	log = logger.NewDefaultLogger(metadata.LogVerbosity, metadata.LogWritter)
	log.SetPackageName(string(consts.CloudLocal))

	obj := &LocalProvider{
		ClusterName: metadata.ClusterName,
		client:      ClientOption(),
	}

	log.Debug("Printing", "localProvider", obj)

	return obj, nil
}

// InitState implements resources.CloudFactory.
func (cloud *LocalProvider) InitState(storage resources.StorageFactory, operation consts.KsctlOperation) error {
	switch operation {
	case consts.OperationStateCreate:
		if isPresent(storage, cloud.ClusterName) {
			return log.NewError("already present")
		}
		localState = &StateConfiguration{
			ClusterName: cloud.ClusterName,
			Distro:      "kind",
		}
		var err error
		err = storage.Path(helpers.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, cloud.ClusterName)).
			Permission(0750).CreateDir()
		if err != nil {
			return log.NewError(err.Error())
		}

		err = saveStateHelper(storage, helpers.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, cloud.ClusterName, STATE_FILE))
		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.OperationStateDelete, consts.OperationStateGet:
		err := loadStateHelper(storage, helpers.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, cloud.ClusterName, STATE_FILE))
		if err != nil {
			return log.NewError(err.Error())
		}
	}
	log.Debug("initialized the state")
	return nil
}

// it will contain the name of the resource to be created
func (cloud *LocalProvider) Name(resName string) resources.CloudFactory {
	cloud.Metadata.ResName = resName
	return cloud
}

func (cloud *LocalProvider) Application(s string) (externalApps bool) {
	return true
}

func (client *LocalProvider) CNI(s string) (externalCNI bool) {
	log.Debug("Printing", "cni", s)

	switch consts.KsctlValidCNIPlugin(s) {
	case consts.CNIKind, "":
		client.Metadata.Cni = string(consts.CNIKind)
	default:
		client.Metadata.Cni = string(consts.CNINone)
		return true
	}

	return false
}

// Version implements resources.CloudFactory.
func (cloud *LocalProvider) Version(ver string) resources.CloudFactory {
	// TODO: validation of version
	log.Debug("Printing", "k8sVersion", ver)
	cloud.Metadata.Version = ver
	return cloud
}

func GetRAWClusterInfos(storage resources.StorageFactory, meta resources.Metadata) ([]cloudControlRes.AllClusterData, error) {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudLocal))

	var data []cloudControlRes.AllClusterData

	managedFolders, err := storage.Path(helpers.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang)).GetFolders()
	if err != nil {
		return nil, err
	}

	log.Debug("Printing", "managedFolderContents", managedFolders)

	for _, folder := range managedFolders {

		path := helpers.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, folder[0], STATE_FILE)
		raw, err := storage.Path(path).Load()
		if err != nil {
			return nil, err
		}
		var clusterState *StateConfiguration
		if err := json.Unmarshal(raw, &clusterState); err != nil {
			return nil, err
		}

		data = append(data,
			cloudControlRes.AllClusterData{
				Provider:   consts.CloudLocal,
				Name:       folder[0],
				Region:     "N/A",
				Type:       consts.ClusterTypeMang,
				K8sDistro:  consts.KsctlKubernetes(clusterState.Distro),
				K8sVersion: clusterState.Version,
				NoMgt:      clusterState.Nodes,
			})
	}

	log.Debug("Printing", "clusterInfo", data)
	return data, nil
}

func (obj *LocalProvider) SwitchCluster(storage resources.StorageFactory) error {

	if isPresent(storage, obj.ClusterName) {

		printKubeconfig(storage, consts.OperationStateCreate, obj.ClusterName)
		return nil
	}
	return log.NewError("Cluster not found")
}

// //// NOT IMPLEMENTED //////

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *LocalProvider) Role(consts.KsctlRole) resources.CloudFactory {
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
	return nil
}

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (*LocalProvider) CreateUploadSSHKeyPair(state resources.StorageFactory) error {
	return nil

}

// DelFirewall implements resources.CloudFactory.
func (*LocalProvider) DelFirewall(state resources.StorageFactory) error {
	return nil
}

// DelNetwork implements resources.CloudFactory.
func (*LocalProvider) DelNetwork(state resources.StorageFactory) error {
	return nil
}

// DelSSHKeyPair implements resources.CloudFactory.
func (*LocalProvider) DelSSHKeyPair(state resources.StorageFactory) error {
	return nil
}

// DelVM implements resources.CloudFactory.
func (*LocalProvider) DelVM(resources.StorageFactory, int) error {
	return nil
}

// GetStateForHACluster implements resources.CloudFactory.
func (*LocalProvider) GetStateForHACluster(state resources.StorageFactory) (cloud.CloudResourceState, error) {
	return cloud.CloudResourceState{}, fmt.Errorf("[local] should not be implemented")
}

// NewFirewall implements resources.CloudFactory.
func (*LocalProvider) NewFirewall(state resources.StorageFactory) error {
	return nil
}

// NewNetwork implements resources.CloudFactory.
func (*LocalProvider) NewNetwork(state resources.StorageFactory) error {
	return nil
}

// NewVM implements resources.CloudFactory.
func (*LocalProvider) NewVM(resources.StorageFactory, int) error {
	return nil
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
