package universal

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubesimplify/ksctl/api/resources"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubernetes struct {
	Metadata            resources.Metadata
	config              *rest.Config
	clientset           *kubernetes.Clientset
	apiextensionsClient *clientset.Clientset
}

type Data struct {
	Url string
	//.....
}

var (
	apps map[string]Data
)

func DeleteNode(storage resources.StorageFactory, nodeName string, kubeconfigPath string) error {

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
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
	err = clientset.CoreV1().Nodes().Delete(context.TODO(), kNodeName, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	storage.Logger().Success("[kubernetes] Deleted Node", kNodeName)
	return nil
}

func initApps() {
	apps = map[string]Data{
		"cilium":  {},
		"flannel": {},
		"argocd":  {},
	}
}

func GetApps(storage resources.StorageFactory, name string) (Data, error) {
	if apps == nil {
		initApps()
	}

	val, present := apps[name]

	if !present {
		return Data{}, fmt.Errorf("[kubernetes] app not found %s", name)
	}
	return val, nil
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

	return nil
}
