package local

import (
	"github.com/kubesimplify/ksctl/pkg/resources"
	"sigs.k8s.io/kind/pkg/cluster"
	"time"
)

type LocalGo interface {
	NewProvider(log resources.LoggerFactory, storage resources.StorageFactory, options ...cluster.ProviderOption)
	Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error
	Delete(name string, explicitKubeconfigPath string) error
}

type LocalGoClient struct {
	provider *cluster.Provider
	log      resources.LoggerFactory
}

type LocalGoMockClient struct {
	log   resources.LoggerFactory
	store resources.StorageFactory
}

func ProvideMockClient() LocalGo {
	return &LocalGoMockClient{}
}

func ProvideClient() LocalGo {
	return &LocalGoClient{}
}

func (l *LocalGoClient) NewProvider(log resources.LoggerFactory, _ resources.StorageFactory, options ...cluster.ProviderOption) {
	logger := CustomLogger{log}
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

func (l *LocalGoMockClient) NewProvider(log resources.LoggerFactory, storage resources.StorageFactory, options ...cluster.ProviderOption) {
	log.Debug("NewProvider initialized", "options", options)
	l.store = storage
	l.log = log
}

func (l *LocalGoMockClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	path, err := createNecessaryConfigs(name)
	l.log.Debug("Printing", "path", path, "error", err)
	l.log.Success("Created the cluster",
		"name", name, "config", config,
		"image", image, "wait", wait.String(),
		"configPath", explictPath())
	return nil
}

func (l *LocalGoMockClient) Delete(name string, explicitKubeconfigPath string) error {
	l.log.Success("Deleted the cluster", "name", name, "kubeconfigPath", explicitKubeconfigPath)
	return nil
}
