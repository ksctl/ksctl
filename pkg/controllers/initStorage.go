package controllers

import (
	"context"

	externalmongostate "github.com/kubesimplify/ksctl/internal/storage/external/mongodb"
	localstate "github.com/kubesimplify/ksctl/internal/storage/local"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
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
