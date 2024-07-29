//go:build testing_k8sclient

package k8sclient

import (
	fake2 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func NewK8sClient(_ *rest.Config) (k *K8sClient, err error) {
	k = new(K8sClient)

	k.apiextensionsClient = fake2.NewSimpleClientset()
	if err != nil {
		return nil, err
	}

	k.clientset = fake.NewSimpleClientset()
	if err != nil {
		return nil, err
	}
	return k, nil
}
