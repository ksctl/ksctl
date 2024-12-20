package managed

import (
	"context"

	bootstrapController "github.com/ksctl/ksctl/pkg/controllers/bootstrap"
	cloudController "github.com/ksctl/ksctl/pkg/controllers/cloud"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

type ManagerClusterManaged struct {
	managerInfo
}

func NewManagerClusterManaged(ctx context.Context, log types.LoggerFactory, client *types.KsctlClient) (*ManagerClusterManaged, error) {
	defer panicCatcher(log)

	stateDocument = new(storageTypes.StorageDocument)
	controllerCtx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-manager")

	cloudController.InitLogger(controllerCtx, log)
	bootstrapController.InitLogger(controllerCtx, log)

	manager := new(ManagerClusterManaged)
	manager.log = log
	manager.client = client

	if err := manager.initStorage(controllerCtx); err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	if err := manager.startPoller(controllerCtx); err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	return manager, nil
}

func (manager *ManagerClusterManaged) CreateCluster() error {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.setupConfigurations(); err != nil {
		log.Error("handled error", "catch", err)
		return err
	}

	if client.Metadata.Provider == consts.CloudLocal {
		client.Metadata.Region = "LOCAL"
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeMang); err != nil {

		log.Error("handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin.StackName)) {
		err := log.NewError(controllerCtx, "invalid CNI plugin")
		log.Error("handled error", "catch", err)
		return err
	}

	if err := cloudController.InitCloud(client, stateDocument, consts.OperationCreate); err != nil {
		log.Error("handled error", "catch", err)
		return err
	}

	// it gets supportForApps, supportForCNI, error
	_, externalCNI, cloudResErr := cloudController.CreateManagedCluster(client)
	if cloudResErr != nil {
		log.Error("handled error", "catch", cloudResErr)
		return cloudResErr
	}

	if err := bootstrapController.InstallAdditionalTools(
		externalCNI,
		client,
		stateDocument); err != nil {

		log.Error("handled error", "catch", err)
		return err
	}

	log.Success(controllerCtx, "successfully created managed cluster")
	return nil
}

func (manager *ManagerClusterManaged) DeleteCluster() error {

	client := manager.client
	log := manager.log

	defer panicCatcher(log)
	if err := manager.setupConfigurations(); err != nil {
		log.Error("handled error", "catch", err)
		return err
	}

	if client.Metadata.Provider == consts.CloudLocal {
		client.Metadata.Region = "LOCAL"
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeMang); err != nil {

		log.Error("handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	if err := cloudController.InitCloud(client, stateDocument, consts.OperationDelete); err != nil {
		log.Error("handled error", "catch", err)
		return err
	}

	cloudResErr := cloudController.DeleteManagedCluster(client)
	if cloudResErr != nil {
		log.Error("handled error", "catch", cloudResErr)
		return cloudResErr
	}

	log.Success(controllerCtx, "successfully deleted managed cluster")
	return nil
}
