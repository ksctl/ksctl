//go:build !testing_local

package local

import (
	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
	"time"
)

type LocalClient struct {
	provider *cluster.Provider
	log      types.LoggerFactory
}

func ProvideClient() LocalGo {
	return &LocalClient{}
}

func (l *LocalClient) NewProvider(log types.LoggerFactory, _ types.StorageFactory, options ...cluster.ProviderOption) {
	logger := &customLogger{Logger: log}
	options = append(options, cluster.ProviderWithLogger(logger))
	l.log = log
	l.provider = cluster.NewProvider(options...)
}

func (l *LocalClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
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

func (l *LocalClient) Delete(name string, explicitKubeconfigPath string) error {
	return l.provider.Delete(name, explicitKubeconfigPath)
}
