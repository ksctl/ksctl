package local

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ksctl/ksctl/internal/storage/types"

	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
	cloudControlRes "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

type Metadata struct {
	ResName           string
	Version           string
	tempDirKubeconfig string
	Cni               string
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	NoNodes     int    `json:"no_nodes"`
	Region      string
	Metadata

	client LocalGo
}

var (
	mainStateDocument *types.StorageDocument
	log               resources.LoggerFactory
	localCtx          context.Context
)

func (*LocalProvider) GetStateFile(resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", err
	}
	log.Debug("Printing", "cloudState", cloudstate)
	return string(cloudstate), nil
}

func NewClient(parentCtx context.Context, meta resources.Metadata, parentLogger resources.LoggerFactory, state *types.StorageDocument, ClientOption func() LocalGo) (*LocalProvider, error) {
	log = parentLogger // intentional shallow copy so that we can use the same
	// logger to be used multiple places
	localCtx = context.WithValue(parentCtx, consts.ContextModuleNameKey, "local")

	mainStateDocument = state

	obj := &LocalProvider{
		ClusterName: meta.ClusterName,
		client:      ClientOption(),
		Region:      meta.Region,
	}
	obj.Metadata.Version = meta.K8sVersion

	log.Debug(localCtx, "Printing", "localProvider", obj)

	return obj, nil
}

// InitState implements resources.CloudFactory.
func (cloud *LocalProvider) InitState(storage resources.StorageFactory, operation consts.KsctlOperation) error {
	switch operation {
	case consts.OperationCreate:
		if isPresent(storage, cloud.ClusterName) {
			return log.NewError("already present")
		}
		log.Debug("Fresh state!!")

		mainStateDocument.ClusterName = cloud.ClusterName
		mainStateDocument.Region = cloud.Region
		mainStateDocument.CloudInfra = &types.InfrastructureState{Local: &types.StateConfigurationLocal{}}
		mainStateDocument.InfraProvider = consts.CloudLocal
		mainStateDocument.ClusterType = string(consts.ClusterTypeMang)

		mainStateDocument.CloudInfra.Local.B.KubernetesDistro = "kind"
		mainStateDocument.CloudInfra.Local.B.KubernetesVer = cloud.Metadata.Version
	case consts.OperationDelete, consts.OperationGet:
		err := loadStateHelper(storage)
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

func (cloud *LocalProvider) Application(s []string) (externalApps bool) {
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
	log = logger.NewStructuredLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudLocal))

	var data []cloudControlRes.AllClusterData
	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudLocal),
		consts.ClusterType: string(consts.ClusterTypeMang),
	})
	if err != nil {
		return nil, err
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloudControlRes.AllClusterData{
				Provider: consts.CloudLocal,
				Name:     v.ClusterName,
				Region:   v.Region,
				Type:     K,

				NoMgt: v.CloudInfra.Local.Nodes,

				K8sDistro:  consts.KsctlKubernetes(v.CloudInfra.Local.B.KubernetesDistro),
				K8sVersion: v.CloudInfra.Local.B.KubernetesVer,
			})
			log.Debug("Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func (obj *LocalProvider) IsPresent(storage resources.StorageFactory) error {

	if isPresent(storage, obj.ClusterName) {

		return nil
	}
	return log.NewError("Cluster not found")
}

// //// NOT IMPLEMENTED //////
func (cloud *LocalProvider) Credential(_ resources.StorageFactory) error {
	panic("no support")
}

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
