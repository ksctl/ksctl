package k8sclient

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

type K8sClient struct {
	clientset           kubernetes.Interface
	apiextensionsClient clientset.Interface
}
