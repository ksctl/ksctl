package kubernetes

import (
	"strings"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubernetes struct {
	Metadata            resources.Metadata
	StorageDriver       resources.StorageFactory
	config              *rest.Config
	clientset           *kubernetes.Clientset
	apiextensionsClient *clientset.Clientset
	helmClient          *HelmClient
	InCluster           bool
}

var (
	log resources.LoggerFactory
)

func (k *Kubernetes) DeleteWorkerNodes(nodeName string) error {

	nodes, err := k.nodesList()
	if err != nil {
		return log.NewError(err.Error())
	}

	kNodeName := ""
	for _, node := range nodes.Items {
		log.Debug("string compariazion", "nodeToDelete", nodeName, "kubernetesNodeName", node.Name)
		if strings.HasPrefix(node.Name, nodeName) {
			kNodeName = node.Name
			break
		}
	}

	if len(kNodeName) == 0 {
		return log.NewError("Not found!")
	}
	err = k.nodeDelete(kNodeName)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("Deleted Node", "name", kNodeName)
	return nil
}

func (k *Kubernetes) NewInClusterClient() (err error) {
	log = logger.NewStructuredLogger(k.Metadata.LogVerbosity, k.Metadata.LogWritter)
	log.SetPackageName("kubernetes-client")

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
	k.InCluster = true // it helps us to identify if we are inside the cluster or not

	initApps()

	return nil
}

func (k *Kubernetes) NewKubeconfigClient(kubeconfig string) (err error) {
	log = logger.NewStructuredLogger(k.Metadata.LogVerbosity, k.Metadata.LogWritter)
	log.SetPackageName("kubernetes-client")

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

	return nil
}
