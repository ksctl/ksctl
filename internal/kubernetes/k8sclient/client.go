package k8sclient

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	clientset           *kubernetes.Clientset
	apiextensionsClient *clientset.Clientset
}

func New(c *rest.Config) (k *K8sClient, err error) {
	k = new(K8sClient)
	k.apiextensionsClient, err = clientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	k.clientset, err = kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return k, nil
}
