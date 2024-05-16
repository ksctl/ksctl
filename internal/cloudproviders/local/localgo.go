package local

import (
	"time"

	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
)

type LocalGo interface {
	NewProvider(log types.LoggerFactory, storage types.StorageFactory, options ...cluster.ProviderOption)
	Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error
	Delete(name string, explicitKubeconfigPath string) error
}

type LocalGoClient struct {
	provider *cluster.Provider
	log      types.LoggerFactory
}

type LocalGoMockClient struct {
	log   types.LoggerFactory
	store types.StorageFactory
}

func ProvideMockClient() LocalGo {
	return &LocalGoMockClient{}
}

func ProvideClient() LocalGo {
	return &LocalGoClient{}
}

func (l *LocalGoClient) NewProvider(log types.LoggerFactory, _ types.StorageFactory, options ...cluster.ProviderOption) {
	logger := &customLogger{Logger: log}
	options = append(options, cluster.ProviderWithLogger(logger))
	l.log = log
	l.provider = cluster.NewProvider(options...)
}

func (l *LocalGoClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	return l.provider.Create(
		name,
		config,
		cluster.CreateWithNodeImage(image),
		cluster.CreateWithWaitForReady(wait),
		cluster.CreateWithKubeconfigPath(explictPath()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
}

func (l *LocalGoClient) Delete(name string, explicitKubeconfigPath string) error {
	return l.provider.Delete(name, explicitKubeconfigPath)
}

func (l *LocalGoMockClient) NewProvider(log types.LoggerFactory, storage types.StorageFactory, options ...cluster.ProviderOption) {
	log.Debug(localCtx, "NewProvider initialized", "options", options)
	l.store = storage
	l.log = log
}

func (l *LocalGoMockClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	path, _ := createNecessaryConfigs(name)
	l.log.Debug(localCtx, "Printing", "path", path)
	l.log.Success(localCtx, "Created the cluster",
		"name", name, "config", config,
		"image", image, "wait", wait.String(),
		"configPath", explictPath())
	return nil
}

func (l *LocalGoMockClient) Delete(name string, explicitKubeconfigPath string) error {
	l.log.Success(localCtx, "Deleted the cluster", "name", name, "kubeconfigPath", explicitKubeconfigPath)
	return nil
}
