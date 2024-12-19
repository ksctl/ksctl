package local

import (
	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
	"time"
)

type metadata struct {
	resName           string
	version           string
	tempDirKubeconfig string
	cni               string
}

type LocalProvider struct {
	clusterName string
	region      string
	vmType      string
	metadata

	client LocalGo
}

type LocalGo interface {
	NewProvider(log types.LoggerFactory, storage types.StorageFactory, options ...cluster.ProviderOption)
	Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error
	Delete(name string, explicitKubeconfigPath string) error
}
