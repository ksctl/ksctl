package controllers

import (
	"context"

	awsPkg "github.com/ksctl/ksctl/internal/cloudproviders/aws"
	azurePkg "github.com/ksctl/ksctl/internal/cloudproviders/azure"
	civoPkg "github.com/ksctl/ksctl/internal/cloudproviders/civo"
	localPkg "github.com/ksctl/ksctl/internal/cloudproviders/local"
	bootstrapController "github.com/ksctl/ksctl/pkg/controllers/bootstrap"
	cloudController "github.com/ksctl/ksctl/pkg/controllers/cloud"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	cloudControllerResource "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

type ManagerClusterKsctl struct {
	managerInfo
}

func NewManagerClusterKsctl(ctx context.Context, log types.LoggerFactory, client *types.KsctlClient) (*ManagerClusterKsctl, error) {
	defer panicCatcher(log)

	stateDocument = new(storageTypes.StorageDocument)
	controllerCtx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-manager")

	cloudController.InitLogger(controllerCtx, log)
	bootstrapController.InitLogger(controllerCtx, log)

	manager := new(ManagerClusterKsctl)
	manager.log = log
	manager.client = client

	err := manager.initStorage(controllerCtx)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	return manager, nil
}

func (manager *ManagerClusterKsctl) Credentials() error {
	log := manager.log
	client := manager.client

	defer panicCatcher(log)

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)

	case consts.CloudAzure:
		client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)

	case consts.CloudAws:
		client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)

	default:
		err = log.NewError(controllerCtx, "Currently not supported!")
	}

	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	err = client.Cloud.Credential(client.Storage)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}
	log.Success(controllerCtx, "Successfully Credential Added")

	return nil
}

func (manager *ManagerClusterKsctl) SwitchCluster() (*string, error) {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	if client.Metadata.Provider == consts.CloudLocal {
		client.Metadata.Region = "LOCAL"
	}

	clusterType := consts.ClusterTypeMang
	if client.Metadata.IsHA {
		clusterType = consts.ClusterTypeHa
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		clusterType); err != nil {

		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, civoPkg.ProvideClient)

	case consts.CloudAzure:
		client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, azurePkg.ProvideClient)

	case consts.CloudAws:
		client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, awsPkg.ProvideClient)

	case consts.CloudLocal:
		client.Cloud, err = localPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, localPkg.ProvideClient)

	default:
		err = log.NewError(controllerCtx, "Currently not supported!")
	}

	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	if err := client.Cloud.IsPresent(client.Storage); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	read, err := client.Storage.Read()
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}
	log.Debug(controllerCtx, "data", "read", read)

	kubeconfig := read.ClusterKubeConfig
	log.Debug(controllerCtx, "data", "kubeconfig", kubeconfig)

	path, err := helpers.WriteKubeConfig(controllerCtx, kubeconfig)
	log.Debug(controllerCtx, "data", "kubeconfigPath", path)

	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	printKubeConfig(manager.log, path)

	return &kubeconfig, nil
}

func (manager *ManagerClusterKsctl) GetCluster() error {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.validationFields(client.Metadata); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if client.Metadata.Provider == consts.CloudLocal {
		client.Metadata.Region = "LOCAL"
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	log.Note(controllerCtx, "Filter", "cloudProvider", string(client.Metadata.Provider))

	var (
		cloudMapper = map[consts.KsctlCloud]types.CloudFactory{
			consts.CloudCivo:  nil,
			consts.CloudAzure: nil,
			consts.CloudAws:   nil,
			consts.CloudLocal: nil,
		}
	)

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		cloudMapper[consts.CloudCivo], err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)

	case consts.CloudAzure:
		cloudMapper[consts.CloudAzure], err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)

	case consts.CloudAws:
		cloudMapper[consts.CloudAws], err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)

	case consts.CloudLocal:
		cloudMapper[consts.CloudLocal], err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)

	case consts.CloudAll:
		cloudMapper[consts.CloudCivo], err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		cloudMapper[consts.CloudAzure], err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		cloudMapper[consts.CloudAws], err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		cloudMapper[consts.CloudLocal], err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}

	default:
		err = log.NewError(controllerCtx, "Currently not supported!")
	}

	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	var printerTable []cloudControllerResource.AllClusterData
	for _, v := range cloudMapper {
		if v == nil {
			continue
		}
		data, err := v.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)
	}

	log.Table(controllerCtx, printerTable)

	log.Success(controllerCtx, "successfully get clusters")

	return nil
}
