package controllers

import (
	"context"

	externalmongostate "github.com/ksctl/ksctl/internal/storage/external/mongodb"
	kubernetesstate "github.com/ksctl/ksctl/internal/storage/kubernetes"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

// InitializeStorageFactory it initializes the storage class
func (manager *KsctlControllerClient) initStorage(ctx context.Context) error {

	switch manager.client.Metadata.StateLocation {
	case consts.StoreLocal:
		manager.client.Storage = localstate.NewClient(ctx, manager.log)
	case consts.StoreExtMongo:
		manager.client.Storage = externalmongostate.NewClient(ctx, manager.log)
	case consts.StoreK8s:
		manager.client.Storage = kubernetesstate.NewClient(ctx, manager.log)
	default:
		return manager.log.NewError(ctx, "invalid storage provider")
	}

	if err := manager.client.Storage.Connect(); err != nil {
		return err
	}
	manager.log.Debug(ctx, "initialized storageFactory")
	return nil
}
