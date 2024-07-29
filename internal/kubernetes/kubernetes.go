package kubernetes

import (
	"context"
	"strings"

	"github.com/ksctl/ksctl/internal/kubernetes/helmclient"
	"github.com/ksctl/ksctl/internal/kubernetes/k8sclient"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClusterClient struct {
	storageDriver types.StorageFactory
	helmClient    *helmclient.HelmClient
	k8sClient     *k8sclient.K8sClient
	inCluster     bool
}

var (
	log           types.LoggerFactory
	kubernetesCtx context.Context
)

func (k *K8sClusterClient) DeleteWorkerNodes(nodeName string) error {
	nodes, err := k.k8sClient.NodesList(kubernetesCtx, log)
	if err != nil {
		return err
	}

	kNodeName := ""
	for _, node := range nodes.Items {
		log.Debug(kubernetesCtx, "string compariazion", "nodeToDelete", nodeName, "kubernetesNodeName", node.Name)
		if strings.HasPrefix(node.Name, nodeName) {
			kNodeName = node.Name
			break
		}
	}

	if len(kNodeName) == 0 {
		return ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
			log.NewError(kubernetesCtx, "node not found!"),
		)
	}
	err = k.k8sClient.NodeDelete(kubernetesCtx, log, kNodeName)
	if err != nil {
		return err
	}
	log.Success(kubernetesCtx, "Deleted Node", "name", kNodeName)
	return nil
}

// ////////////////////
type K8sClient interface{}

type HelmClient interface{}

func NewInClusterClient(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	storage types.StorageFactory,
	k8s K8sClient,
	helm HelmClient,
) (k *K8sClusterClient, err error) {
	if k8s == nil && helm == nil {
		return newInClusterClientWithConfig(parentCtx, parentLog, storage, &k8sclient.K8sClient{}, &helmclient.HelmClient{})
	}
	return newInClusterClientWithConfig(parentCtx, parentLog, storage, k8s, helm)
}

func newInClusterClientWithConfig(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	storage types.StorageFactory,
	k8s K8sClient,
	helm HelmClient,
) (k *K8sClusterClient, err error) {
	return
}

//////////////////////

func NewInClusterClientWithConfig(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	storage types.StorageFactory) (k *K8sClusterClient, err error) {

	kubernetesCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	log = parentLog

	k = &K8sClusterClient{
		storageDriver: storage,
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}

	k.k8sClient, err = k8sclient.NewK8sClient(config)
	if err != nil {
		return
	}

	k.helmClient, err = helmclient.NewInClusterHelmClient(
		kubernetesCtx,
		log,
	)
	if err != nil {
		return
	}
	k.inCluster = true // it helps us to identify if we are inside the cluster or not

	return k, nil
}

func NewKubeconfigClient(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	storage types.StorageFactory,
	kubeconfig string) (k *K8sClusterClient, err error) {

	kubernetesCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	log = parentLog

	k = &K8sClusterClient{
		storageDriver: storage,
	}

	rawKubeconfig := []byte(kubeconfig)

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		return clientcmd.Load(rawKubeconfig)
	})
	if err != nil {
		return
	}

	k.k8sClient, err = k8sclient.NewK8sClient(config)
	if err != nil {
		return
	}

	k.helmClient, err = helmclient.NewKubeconfigHelmClient(
		kubernetesCtx,
		log,
		kubeconfig,
	)
	if err != nil {
		return
	}

	return k, nil
}
