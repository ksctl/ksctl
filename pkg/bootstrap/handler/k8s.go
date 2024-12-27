package handler

import (
	"context"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/k8s"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/storage"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type K8sClusterClient struct {
	ctx           context.Context
	l             logger.Logger
	storageDriver storage.Storage

	r *rest.Config

	helmClient *helm.Client
	k8sClient  *k8s.Client
	inCluster  bool
}

func NewClusterClient(
	parentCtx context.Context,
	parentLog logger.Logger,
	storage storage.Storage,
	kubeconfig string,
) (k *K8sClusterClient, err error) {
	k = &K8sClusterClient{
		storageDriver: storage,
	}

	k.ctx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	k.l = parentLog

	rawKubeconfig := []byte(kubeconfig)

	config := &rest.Config{}
	config, err = clientcmd.BuildConfigFromKubeconfigGetter(
		"",
		func() (*api.Config, error) {
			return clientcmd.Load(rawKubeconfig)
		})
	if err != nil {
		return
	}

	k.k8sClient, err = k8s.NewK8sClient(parentCtx, parentLog, config)
	if err != nil {
		return
	}
	k.r = config

	k.helmClient, err = helm.NewKubeconfigHelmClient(
		k.ctx,
		k.l,
		kubeconfig,
	)
	if err != nil {
		return
	}

	return k, nil
}

func getVersionIfItsNotNilAndLatest(ver *string, defaultVer string) string {
	if ver == nil {
		return defaultVer
	}
	if *ver == "latest" {
		return defaultVer
	}
	return *ver
}
