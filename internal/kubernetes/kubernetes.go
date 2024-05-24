package kubernetes

import (
	"context"
	"strings"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubernetes struct {
	storageDriver       types.StorageFactory
	config              *rest.Config
	clientset           *kubernetes.Clientset
	apiextensionsClient *clientset.Clientset
	helmClient          *HelmClient
	inCluster           bool
}

var (
	log           types.LoggerFactory
	kubernetesCtx context.Context
)

func (k *Kubernetes) DeleteWorkerNodes(nodeName string) error {

	nodes, err := k.nodesList()
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
		return log.NewError(kubernetesCtx, "node not found!")
	}
	err = k.nodeDelete(kNodeName)
	if err != nil {
		return err
	}
	log.Success(kubernetesCtx, "Deleted Node", "name", kNodeName)
	return nil
}

func NewInClusterClient(parentCtx context.Context, parentLog types.LoggerFactory, storage types.StorageFactory) (k *Kubernetes, err error) {
	kubernetesCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	log = parentLog

	k = &Kubernetes{
		storageDriver: storage,
	}

	k.config, err = rest.InClusterConfig()
	if err != nil {
		return
	}

	k.apiextensionsClient, err = clientset.NewForConfig(k.config)
	if err != nil {
		return
	}

	k.clientset, err = kubernetes.NewForConfig(k.config)
	if err != nil {
		return
	}

	k.helmClient = new(HelmClient)
	if err = k.helmClient.NewInClusterHelmClient(); err != nil {
		return
	}
	k.inCluster = true // it helps us to identify if we are inside the cluster or not

	initApps()

	return k, nil
}

func NewKubeconfigClient(parentCtx context.Context, parentLog types.LoggerFactory, storage types.StorageFactory, kubeconfig string) (k *Kubernetes, err error) {
	kubernetesCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	log = parentLog

	k = &Kubernetes{
		storageDriver: storage,
	}

	rawKubeconfig := []byte(kubeconfig)

	k.config, err = clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		return clientcmd.Load(rawKubeconfig)
	})
	if err != nil {
		return
	}

	k.apiextensionsClient, err = clientset.NewForConfig(k.config)
	if err != nil {
		return
	}

	k.clientset, err = kubernetes.NewForConfig(k.config)
	if err != nil {
		return
	}

	k.helmClient = new(HelmClient)
	if err = k.helmClient.NewKubeconfigHelmClient(kubeconfig); err != nil {
		return
	}

	initApps()

	return k, nil
}
