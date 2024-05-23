package types

import (
	"github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type StorageFactory interface {
	// Kill to achieve graceful termination we can store a boolean flag in the
	// storagedriver that whether there was any write operation if yes and a reference
	//always present in the storagedriver we can make the driver write the struct once termination is triggered
	Kill() error

	Connect() error

	Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	Write(*storage.StorageDocument) error

	WriteCredentials(consts.KsctlCloud, *storage.CredentialsDocument) error

	Read() (*storage.StorageDocument, error)

	ReadCredentials(consts.KsctlCloud) (*storage.CredentialsDocument, error)

	DeleteCluster() error

	AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	GetOneOrMoreClusters(filters map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*storage.StorageDocument, error)

	// Export is not goroutine safe, but the child process it calls is!
	Export(filters map[consts.KsctlSearchFilter]string) (*StorageStateExportImport, error)

	// Import is not goroutine safe, but the child process it calls is!
	Import(*StorageStateExportImport) error
}

type StorageStateExportImport struct {
	Clusters    []*storage.StorageDocument
	Credentials []*storage.CredentialsDocument
}
