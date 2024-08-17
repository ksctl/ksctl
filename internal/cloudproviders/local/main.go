package local

import (
	"context"
	"encoding/json"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	cloudControlRes "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

func (*LocalProvider) GetStateFile(types.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(
			log.NewError(localCtx, "failed to serialize the state", "Reason", err),
		)
	}
	return string(cloudstate), nil
}

func NewClient(parentCtx context.Context, meta types.Metadata, parentLogger types.LoggerFactory, state *storageTypes.StorageDocument, ClientOption func() LocalGo) (*LocalProvider, error) {
	log = parentLogger // intentional shallow copy so that we can use the same
	// logger to be used multiple places
	localCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.CloudLocal))

	mainStateDocument = state

	obj := &LocalProvider{
		clusterName: meta.ClusterName,
		client:      ClientOption(),
		region:      meta.Region,
	}
	obj.metadata.version = meta.K8sVersion

	log.Debug(localCtx, "Printing", "localProvider", obj)

	return obj, nil
}

func (cloud *LocalProvider) InitState(storage types.StorageFactory, operation consts.KsctlOperation) error {
	switch operation {
	case consts.OperationCreate:
		if err := isPresent(storage, cloud.clusterName); err == nil {
			return ksctlErrors.ErrDuplicateRecords.Wrap(
				log.NewError(localCtx, "already present", "name", cloud.clusterName),
			)
		}
		log.Debug(localCtx, "Fresh state!!")

		mainStateDocument.ClusterName = cloud.clusterName
		mainStateDocument.Region = cloud.region
		mainStateDocument.CloudInfra = &storageTypes.InfrastructureState{Local: &storageTypes.StateConfigurationLocal{}}
		mainStateDocument.InfraProvider = consts.CloudLocal
		mainStateDocument.ClusterType = string(consts.ClusterTypeMang)

		mainStateDocument.CloudInfra.Local.B.KubernetesVer = cloud.metadata.version
	case consts.OperationDelete, consts.OperationGet:
		err := loadStateHelper(storage)
		if err != nil {
			return err
		}
	}
	log.Debug(localCtx, "initialized the state")
	return nil
}

func (cloud *LocalProvider) Name(resName string) types.CloudFactory {
	cloud.metadata.resName = resName
	return cloud
}

func (cloud *LocalProvider) Application(s []string) (externalApps bool) {
	return true
}

func (client *LocalProvider) CNI(s string) (externalCNI bool) {
	log.Debug(localCtx, "Printing", "cni", s)

	switch consts.KsctlValidCNIPlugin(s) {
	case consts.CNIKind, "":
		client.metadata.cni = string(consts.CNIKind)
	default:
		client.metadata.cni = string(consts.CNINone)
		return true
	}

	return false
}

func (cloud *LocalProvider) ManagedK8sVersion(ver string) types.CloudFactory {
	log.Debug(localCtx, "Printing", "k8sVersion", ver)
	cloud.metadata.version = ver
	return cloud
}

func (cloud *LocalProvider) GetRAWClusterInfos(storage types.StorageFactory) ([]cloudControlRes.AllClusterData, error) {

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
				CloudProvider: consts.CloudLocal,
				Name:          v.ClusterName,
				Region:        v.Region,
				ClusterType:   K,

				NoMgt: v.CloudInfra.Local.Nodes,
				Mgt: cloudControlRes.VMData{
					VMSize: v.CloudInfra.Local.ManagedNodeSize,
				},

				K8sDistro:  v.BootstrapProvider,
				K8sVersion: v.CloudInfra.Local.B.KubernetesVer,
			})
			log.Debug(localCtx, "Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func (obj *LocalProvider) IsPresent(storage types.StorageFactory) error {

	return isPresent(storage, obj.clusterName)
}

func (cloud *LocalProvider) VMType(_ string) types.CloudFactory {
	cloud.vmType = "local_machine"
	return cloud
}

func (obj *LocalProvider) GetKubeconfig(storage types.StorageFactory) (*string, error) {
	_read, err := storage.Read()
	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig := _read.ClusterKubeConfig
	return &kubeconfig, nil
}

// //// NOT IMPLEMENTED //////

func (cloud *LocalProvider) Credential(_ types.StorageFactory) error {
	return log.NewError(localCtx, "no support")
}

func (cloud *LocalProvider) Role(consts.KsctlRole) types.CloudFactory {
	return nil
}

func (cloud *LocalProvider) Visibility(bool) types.CloudFactory {
	return nil
}

func (*LocalProvider) GetHostNameAllWorkerNode() []string {
	return nil
}

func (*LocalProvider) CreateUploadSSHKeyPair(state types.StorageFactory) error {
	return nil

}

func (*LocalProvider) DelFirewall(state types.StorageFactory) error {
	return nil
}

func (*LocalProvider) DelNetwork(state types.StorageFactory) error {
	return nil
}

func (*LocalProvider) DelSSHKeyPair(state types.StorageFactory) error {
	return nil
}

func (*LocalProvider) DelVM(types.StorageFactory, int) error {
	return nil
}

func (*LocalProvider) GetStateForHACluster(state types.StorageFactory) (cloud.CloudResourceState, error) {
	return cloud.CloudResourceState{}, log.NewError(localCtx, "should not be implemented")
}

func (*LocalProvider) NewFirewall(state types.StorageFactory) error {
	return nil
}

func (*LocalProvider) NewNetwork(state types.StorageFactory) error {
	return nil
}

func (*LocalProvider) NewVM(types.StorageFactory, int) error {
	return nil
}

func (cloud *LocalProvider) NoOfControlPlane(int, bool) (int, error) {
	return -1, log.NewError(localCtx, "unsupported operation")
}

func (cloud *LocalProvider) NoOfDataStore(int, bool) (int, error) {
	return -1, log.NewError(localCtx, "unsupported operation")
}

func (cloud *LocalProvider) NoOfWorkerPlane(types.StorageFactory, int, bool) (int, error) {
	return -1, log.NewError(localCtx, "unsupported operation")
}
