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

// TODO: create a interface so that we can have a mock test for this as well
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

func (this *Kubernetes) DeleteWorkerNodes(nodeName string) error {

	nodes, err := this.nodesList()
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
	err = this.nodeDelete(kNodeName)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("Deleted Node", "name", kNodeName)
	return nil
}

func (this *Kubernetes) NewInClusterClient() (err error) {
	log = logger.NewDefaultLogger(this.Metadata.LogVerbosity, this.Metadata.LogWritter)
	log.SetPackageName("kubernetes-client")

	this.config, err = rest.InClusterConfig()
	if err != nil {
		return
	}

	this.apiextensionsClient, err = clientset.NewForConfig(this.config)
	if err != nil {
		return
	}

	this.clientset, err = kubernetes.NewForConfig(this.config)
	if err != nil {
		return
	}

	this.helmClient = new(HelmClient)
	if err = this.helmClient.NewInClusterHelmClient(); err != nil {
		return
	}
	this.InCluster = true // it helps us to identify if we are inside the cluster or not

	initApps()

	return nil
}

func (this *Kubernetes) NewKubeconfigClient(kubeconfig string) (err error) {
	log = logger.NewDefaultLogger(this.Metadata.LogVerbosity, this.Metadata.LogWritter)
	log.SetPackageName("kubernetes-client")

	rawKubeconfig := []byte(kubeconfig)

	this.config, err = clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		return clientcmd.Load(rawKubeconfig)
	})
	if err != nil {
		return
	}

	this.apiextensionsClient, err = clientset.NewForConfig(this.config)
	if err != nil {
		return
	}

	this.clientset, err = kubernetes.NewForConfig(this.config)
	if err != nil {
		return
	}

	this.helmClient = new(HelmClient)
	if err = this.helmClient.NewKubeconfigHelmClient(kubeconfig); err != nil {
		return
	}

	initApps()

	return nil
}
