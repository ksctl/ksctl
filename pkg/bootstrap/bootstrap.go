package bootstrap

import "github.com/ksctl/ksctl/pkg/consts"

type Bootstrap interface {
	Setup(cloud.CloudResourceState, StorageFactory, consts.KsctlOperation) error

	ConfigureDataStore(int, StorageFactory) error

	ConfigureLoadbalancer(StorageFactory) error
}
