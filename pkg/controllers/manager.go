package controllers

import (
	"context"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"runtime/debug"

	"github.com/ksctl/ksctl/pkg/helpers"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	externalmongostate "github.com/ksctl/ksctl/internal/storage/external/mongodb"
	kubernetesstate "github.com/ksctl/ksctl/internal/storage/kubernetes"
	localstate "github.com/ksctl/ksctl/internal/storage/local"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	controllerCtx context.Context

	stateDocument *storageTypes.StorageDocument
)

type managerInfo struct {
	log    types.LoggerFactory
	client *types.KsctlClient
}

func (manager *managerInfo) initStorage(ctx context.Context) error {
	if !helpers.ValidateStorage(manager.client.Metadata.StateLocation) {
		return ksctlErrors.ErrInvalidStorageProvider.Wrap(
			manager.log.NewError(
				controllerCtx, "Problem in validation", "storage", manager.client.Metadata.StateLocation,
			),
		)
	}
	switch manager.client.Metadata.StateLocation {
	case consts.StoreLocal:
		manager.client.Storage = localstate.NewClient(ctx, manager.log)
	case consts.StoreExtMongo:
		manager.client.Storage = externalmongostate.NewClient(ctx, manager.log)
	case consts.StoreK8s:
		manager.client.Storage = kubernetesstate.NewClient(ctx, manager.log)
	}

	if err := manager.client.Storage.Connect(); err != nil {
		return err
	}
	manager.log.Debug(ctx, "initialized storageFactory")
	return nil
}

func panicCatcher(log types.LoggerFactory) {
	if r := recover(); r != nil {
		log.Error(controllerCtx, "Failed to recover stack trace", "error", r)
		debug.PrintStack()
	}
}

func (manager *managerInfo) setupConfigurations() error {

	if err := manager.validationFields(manager.client.Metadata); err != nil {
		return err
	}

	if err := helpers.IsValidName(controllerCtx, manager.log, manager.client.Metadata.ClusterName); err != nil {
		return err
	}
	return nil
}
