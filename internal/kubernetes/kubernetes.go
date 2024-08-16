package kubernetes

import (
	"context"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/ksctl/ksctl/internal/kubernetes/helmclient"
	"github.com/ksctl/ksctl/internal/kubernetes/k8sclient"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

type K8sClusterClient struct {
	storageDriver types.StorageFactory
	helmClient    HelmClient
	k8sClient     K8sClient
	inCluster     bool
	c             *rest.Config
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

func NewInClusterClient(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	storage types.StorageFactory,
	useMock bool,
	k8sClient K8sClient,
	helmClient HelmClient,
) (k *K8sClusterClient, err error) {

	kubernetesCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	log = parentLog

	k = &K8sClusterClient{
		storageDriver: storage,
	}

	if !useMock {
		config := &rest.Config{}
		config, err = rest.InClusterConfig()
		if err != nil {
			return
		}

		k.k8sClient, err = k8sclient.NewK8sClient(config)
		if err != nil {
			return
		}

		k.c = config

		k.helmClient, err = helmclient.NewInClusterHelmClient(
			kubernetesCtx,
			log,
		)
		if err != nil {
			return
		}
	} else {
		k.k8sClient = k8sClient
		k.helmClient = helmClient
	}
	k.inCluster = true // it helps us to identify if we are inside the cluster or not

	return k, nil
}

func NewKubeconfigClient(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	storage types.StorageFactory,
	kubeconfig string,
	useMock bool,
	k8sClient K8sClient,
	helmClient HelmClient,
) (k *K8sClusterClient, err error) {

	kubernetesCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	log = parentLog

	k = &K8sClusterClient{
		storageDriver: storage,
	}

	if !useMock {
		rawKubeconfig := []byte(kubeconfig)

		config := &rest.Config{}
		config, err = clientcmd.BuildConfigFromKubeconfigGetter(
			"",
			func() (*api.Config, error) {
				return clientcmd.Load(rawKubeconfig)
			})
		if err != nil {
			return
		}

		k.k8sClient, err = k8sclient.NewK8sClient(config)
		if err != nil {
			return
		}
		k.c = config

		k.helmClient, err = helmclient.NewKubeconfigHelmClient(
			kubernetesCtx,
			log,
			kubeconfig,
		)
		if err != nil {
			return
		}
	} else {
		k.k8sClient = k8sClient
		k.helmClient = helmClient
	}

	return k, nil
}
