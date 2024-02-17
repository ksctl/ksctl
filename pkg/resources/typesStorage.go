package resources

import (
	"context"
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type StorageFactory interface {
	// Kill to achieve graceful termination we can store a boolean flag in the
	// storagedriver that whether there was any write operation if yes and a reference
	//always present in the storagedriver we can make the driver write the struct once termination is triggered
	Kill() error

	// Connect TODO: transfer the Logverbosity and LogWriter and all other stuff
	Connect(ctx context.Context) error

	Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	Write(*types.StorageDocument) error

	WriteCredentials(consts.KsctlCloud, *types.CredentialsDocument) error

	Read() (*types.StorageDocument, error)

	ReadCredentials(consts.KsctlCloud) (*types.CredentialsDocument, error)

	DeleteCluster() error

	AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	GetOneOrMoreClusters(filters map[string]string) (map[consts.KsctlClusterType][]*types.StorageDocument, error)
}
