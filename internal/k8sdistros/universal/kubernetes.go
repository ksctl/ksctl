package universal

import (
	"strings"

	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
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

func (this *Kubernetes) ClientInit(kubeconfigPath string) (err error) {
	log = logger.NewDefaultLogger(this.Metadata.LogVerbosity, this.Metadata.LogWritter)
	log.SetPackageName("kubernetes-client")

	this.config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return log.NewError(err.Error())
	}

	this.apiextensionsClient, err = clientset.NewForConfig(this.config)
	if err != nil {
		return log.NewError(err.Error())
	}

	this.clientset, err = kubernetes.NewForConfig(this.config)
	if err != nil {
		return log.NewError(err.Error())
	}

	this.helmClient = new(HelmClient)
	if err := this.helmClient.InitClient(kubeconfigPath); err != nil {
		return log.NewError(err.Error())
	}

	initApps()

	return nil
}
