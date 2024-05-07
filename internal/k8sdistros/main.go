package k8sdistros

import (
	"sync"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	mainStateDocument *types.StorageDocument
	log               resources.LoggerFactory
)

func NewPreBootStrap(m resources.Metadata,
	state *types.StorageDocument) resources.PreKubernetesBootstrap {

	log = logger.NewDefaultLogger(m.LogVerbosity, m.LogWritter)
	log.SetPackageName("bootstrap")

	mainStateDocument = state
	return &PreBootstrap{mu: &sync.Mutex{}}
}

func (p *PreBootstrap) Setup(cloudState cloud.CloudResourceState,
	storage resources.StorageFactory, operation consts.KsctlOperation) error {

	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap = &types.KubernetesBootstrapState{}
		var err error
		mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey,
			err = helpers.GenerateCerts(log, cloudState.PrivateIPv4DataStores)
		if err != nil {
			return err
		}
	}

	mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes =
		utilities.DeepCopySlice[string](cloudState.IPv4ControlPlanes)

	mainStateDocument.K8sBootstrap.B.PrivateIPs.ControlPlanes =
		utilities.DeepCopySlice[string](cloudState.PrivateIPv4ControlPlanes)

	mainStateDocument.K8sBootstrap.B.PublicIPs.DataStores =
		utilities.DeepCopySlice[string](cloudState.IPv4DataStores)
	mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores =
		utilities.DeepCopySlice[string](cloudState.PrivateIPv4DataStores)

	mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes =
		utilities.DeepCopySlice[string](cloudState.IPv4WorkerPlanes)

	mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer =
		cloudState.IPv4LoadBalancer

	mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer =
		cloudState.PrivateIPv4LoadBalancer

	mainStateDocument.K8sBootstrap.B.SSHInfo = cloudState.SSHState

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError("failed to Initialized state from Cloud reason: %v", err)
	}

	log.Debug("Printing", "k3sState", mainStateDocument)

	log.Print("Initialized state from Cloud")
	return nil
}
