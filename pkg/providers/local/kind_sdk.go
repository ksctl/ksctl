package local

import (
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

type KindSDK interface {
	NewProvider(b *Provider, options ...cluster.ProviderOption)
	Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error
	Delete(name string, explicitKubeconfigPath string) error
}
