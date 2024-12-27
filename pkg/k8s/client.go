package k8s

import (
	"context"

	"github.com/ksctl/ksctl/pkg/logger"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	ctx                 context.Context
	l                   logger.Logger
	clientset           kubernetes.Interface
	apiextensionsClient clientset.Interface
	r                   *rest.Config
}

type App struct {
	CreateNamespace bool
	Namespace       string

	Version     string   // how to get the version propogated?
	Urls        []string // make sure you traverse backwards to uninstall resources
	PostInstall string
	Metadata    string
}

func NewK8sClient(ctx context.Context, l logger.Logger, c *rest.Config) (k *Client, err error) {
	k = new(Client)
	k.ctx = context.WithValue(ctx, "module", "kubernetes-client")
	k.r = c
	k.l = l
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
