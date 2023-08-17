package universal

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubesimplify/ksctl/api/resources"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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
