package controllers

import (
	"context"

	externalmongostate "github.com/ksctl/ksctl/internal/storage/external/mongodb"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

// InitializeStorageFactory it initializes the storage class
func InitializeStorageFactory(ctx context.Context, client *resources.KsctlClient) error {

	if log == nil {
		log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
		log.SetPackageName("ksctl-manager")
	}

	switch client.Metadata.StateLocation {
	case consts.StoreLocal:
		client.Storage = localstate.InitStorage(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
	case consts.StoreExtMongo:
		client.Storage = externalmongostate.InitStorage(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
	default:
		return log.NewError("Currently Local state is supported!")
	}

	if err := client.Storage.Connect(ctx); err != nil {
		return err
	}
	log.Debug("initialized storageFactory")
	return nil
}
