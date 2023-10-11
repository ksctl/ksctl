package universal

import (
	"fmt"
	"strings"

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

func (this *Kubernetes) DeleteWorkerNodes(nodeName string) error {

	nodes, err := this.nodesList()
	if err != nil {
		return err
	}

	kNodeName := ""
	for _, node := range nodes.Items {
		if strings.HasPrefix(node.Name, nodeName) {
			kNodeName = node.Name
			break
		}
	}

	if len(kNodeName) == 0 {
		return fmt.Errorf("Not found!")
	}
	err = this.nodeDelete(kNodeName)
	if err != nil {
		return err
	}
	this.StorageDriver.Logger().Success("[client-go] Deleted Node", kNodeName)
	return nil
}

func (this *Kubernetes) ClientInit(kubeconfigPath string) (err error) {
	this.config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	this.apiextensionsClient, err = clientset.NewForConfig(this.config)
	if err != nil {
		return err
	}

	this.clientset, err = kubernetes.NewForConfig(this.config)
	if err != nil {
		return err
	}

	this.helmClient = new(HelmClient)
	if err := this.helmClient.InitClient(kubeconfigPath); err != nil {
		return err
	}

	initApps()

	return nil
}
