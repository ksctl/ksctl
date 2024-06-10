//go:build testing_local

package local

import (
	"time"

	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
)

type LocalClient struct {
	log   types.LoggerFactory
	store types.StorageFactory
}

func ProvideClient() LocalGo {
	return &LocalClient{}
}

func (l *LocalClient) NewProvider(log types.LoggerFactory, storage types.StorageFactory, options ...cluster.ProviderOption) {
	log.Debug(localCtx, "NewProvider initialized", "options", options)
	l.store = storage
	l.log = log
}

func (l *LocalClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	path, _ := createNecessaryConfigs(name)
	l.log.Debug(localCtx, "Printing", "path", path)
	l.log.Success(localCtx, "Created the cluster",
		"name", name, "config", config,
		"image", image, "wait", wait.String(),
		"configPath", explictPath())
	return nil
}

func (l *LocalClient) Delete(name string, explicitKubeconfigPath string) error {
	l.log.Success(localCtx, "Deleted the cluster", "name", name, "kubeconfigPath", explicitKubeconfigPath)
	return nil
}
