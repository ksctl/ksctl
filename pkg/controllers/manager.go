package controllers

import (
	"context"
	"os"
	"runtime/debug"
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	awsPkg "github.com/ksctl/ksctl/internal/cloudproviders/aws"
	azurePkg "github.com/ksctl/ksctl/internal/cloudproviders/azure"
	civoPkg "github.com/ksctl/ksctl/internal/cloudproviders/civo"
	localPkg "github.com/ksctl/ksctl/internal/cloudproviders/local"

	bootstrapController "github.com/ksctl/ksctl/pkg/controllers/bootstrap"
	"github.com/ksctl/ksctl/pkg/controllers/cloud"
	cloudController "github.com/ksctl/ksctl/pkg/controllers/cloud"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	cloudControllerResource "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

var (
	controllerCtx context.Context

	stateDocument *storageTypes.StorageDocument
)

type KsctlControllerClient struct {
	log    types.LoggerFactory
	client *types.KsctlClient
}

func GenKsctlController(
	ctx context.Context,
	log types.LoggerFactory,
	client *types.KsctlClient,
) (*KsctlControllerClient, error) {

	defer panicCatcher(log)

	stateDocument = new(storageTypes.StorageDocument)
	controllerCtx = context.WithValue(ctx, consts.ContextModuleNameKey, "ksctl-manager")

	cloudController.InitLogger(controllerCtx, log)
	bootstrapController.InitLogger(controllerCtx, log)

	manager := &KsctlControllerClient{
		log:    log,
		client: client,
	}

	err := manager.initStorage(controllerCtx)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	return manager, nil
}

func panicCatcher(log types.LoggerFactory) {
	if r := recover(); r != nil {
		log.Error(controllerCtx, "Failed to recover stack trace", "error", r)
		debug.PrintStack()
	}
}

func (manager *KsctlControllerClient) setupConfigurations() error {

	if err := validationFields(manager.client.Metadata); err != nil {
		return manager.log.NewError(controllerCtx, "field validation failed", "Reason", err)
	}

	if err := helpers.IsValidName(controllerCtx, manager.log, manager.client.Metadata.ClusterName); err != nil {
		return err
	}
	return nil
}

func (manager *KsctlControllerClient) Applications(op consts.KsctlOperation) error {

	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
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
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationGet, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if op != consts.OperationCreate && op != consts.OperationDelete {

		err := log.NewError(controllerCtx, "Invalid operation")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := bootstrapController.ApplicationsInCluster(client, stateDocument, op); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}
	return nil
}

func (manager *KsctlControllerClient) Credentials() error {
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

func (manager *KsctlControllerClient) CreateManagedCluster() error {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
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

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		err := log.NewError(controllerCtx, "invalid CNI plugin")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationCreate, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	// it gets supportForApps, supportForCNI, error
	externalApp, externalCNI, cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		log.Error(controllerCtx, "handled error", "catch", cloudResErr)
		return cloudResErr
	}

	if err := bootstrapController.InstallAdditionalTools(
		externalCNI,
		externalApp,
		client,
		stateDocument); err != nil {

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	log.Success(controllerCtx, "successfully created managed cluster")
	return nil
}

func (manager *KsctlControllerClient) DeleteManagedCluster() error {

	client := manager.client
	log := manager.log

	defer panicCatcher(log)
	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
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

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationDelete, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	cloudResErr := cloud.DeleteManagedCluster(client)
	if cloudResErr != nil {
		log.Error(controllerCtx, "handled error", "catch", cloudResErr)
		return cloudResErr
	}

	log.Success(controllerCtx, "successfully deleted managed cluster")
	return nil
}

func (manager *KsctlControllerClient) SwitchCluster() (*string, error) {
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

	path, err := helpers.WriteKubeConfig(kubeconfig)
	log.Debug(controllerCtx, "data", "kubeconfigPath", path)

	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return nil, err
	}

	printKubeConfig(manager.log, path)

	return &kubeconfig, nil
}

func (manager *KsctlControllerClient) GetCluster() error {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := validationFields(client.Metadata); err != nil {
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

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)

	case consts.CloudAzure:
		client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)

	case consts.CloudAws:
		client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)

	case consts.CloudLocal:
		client.Cloud, err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)

	default:
		err = log.NewError(controllerCtx, "Currently not supported!")
	}

	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	var printerTable []cloudControllerResource.AllClusterData
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		data, err := client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

	case consts.CloudLocal:
		data, err := client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAws:
		data, err := client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAzure:
		data, err := client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAll:
		data, err := client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

		data, err = client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

		data, err = client.Cloud.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
		printerTable = append(printerTable, data...)

		data, err = client.Cloud.GetRAWClusterInfos(client.Storage)
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

func (manager *KsctlControllerClient) CreateHACluster() error {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if client.Metadata.Provider == consts.CloudLocal {
		err := log.NewError(controllerCtx, "ha not supported")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}
	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		err := log.NewError(controllerCtx, "invalid CNI plugin")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationCreate, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	err := bootstrapController.Setup(client, stateDocument)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	cloudResErr := cloud.CreateHACluster(client)
	if cloudResErr != nil {
		log.Error(controllerCtx, "handled error", "catch", cloudResErr)
		return cloudResErr
	}

	var payload cloudControllerResource.CloudResourceState
	payload, err = client.Cloud.GetStateForHACluster(client.Storage)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationCreate)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	log.Note(controllerCtx, "only cloud storage are having replay!")

	externalCNI, err := bootstrapController.ConfigureCluster(client)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := bootstrapController.InstallAdditionalTools(
		externalCNI,
		true,
		client,
		stateDocument); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	log.Success(controllerCtx, "successfully created ha cluster")

	return nil
}

func (manager *KsctlControllerClient) DeleteHACluster() error {

	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if client.Metadata.Provider == consts.CloudLocal {
		err := log.NewError(controllerCtx, "ha not supported")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationDelete, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	//if len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
	//
	//	// find a better way to get the kubeconfig location
	//
	//	err := kubernetesController.Setup(client, stateDocument)
	//	if err != nil {
	//		return log.NewError(controllerCtx,err.Error())
	//	}
	//	var payload cloudControllerResource.CloudResourceState
	//	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
	//
	//	err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
	//	if err != nil {
	//		return log.NewError(controllerCtx,err.Error())
	//	}
	//
	//	kubeconfig := stateDocument.ClusterKubeConfig
	//
	//	kubernetesClient := kubernetes.Kubernetes{
	//		Metadata:      client.Metadata,
	//		StorageDriver: client.Storage,
	//	}
	//	if err := kubernetesClient.NewKubeconfigClient(kubeconfig); err != nil {
	//		return log.NewError(controllerCtx,err.Error())
	//	}
	//
	//	if err = kubernetesClient.DeleteResourcesFromController(); err != nil {
	//		return log.NewError(controllerCtx,err.Error())
	//	}
	//
	//	// NOTE: explict make the count of the workernodes as 0 as we need one schedulable workload to test of the operation was successful
	//	if _, err := client.Cloud.NoOfWorkerPlane(client.Storage, 0, true); err != nil {
	//		return log.NewError(controllerCtx,err.Error())
	//	}
	//}

	cloudResErr := cloud.DeleteHACluster(client)
	if cloudResErr != nil {
		log.Error(controllerCtx, "handled error", "catch", cloudResErr)
		return cloudResErr
	}

	log.Success(controllerCtx, "successfully deleted HA cluster")

	return nil
}

func (manager *KsctlControllerClient) AddWorkerPlaneNode() error {
	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	//if client.Metadata.IsHA && len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
	//	// disable add AddWorkerPlaneNode when this feature is being used
	//	return log.NewError(controllerCtx,"This Functionality is diabled for {HA type clusters}", "FEATURE_FLAG", consts.KsctlFeatureFlagHaAutoscale)
	//}

	if client.Metadata.Provider == consts.CloudLocal {
		err := log.NewError(controllerCtx, "ha not supported")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if !client.Metadata.IsHA {
		err := log.NewError(controllerCtx, "this feature is only for ha clusters")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationGet, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	err := bootstrapController.Setup(client, stateDocument)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	currWP, cloudResErr := cloud.AddWorkerNodes(client)
	if cloudResErr != nil {
		log.Error(controllerCtx, "handled error", "catch", cloudResErr)
		return cloudResErr
	}

	var payload cloudControllerResource.CloudResourceState
	payload, err = client.Cloud.GetStateForHACluster(client.Storage)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	log.Note(controllerCtx, "Only cloud storage are having replay!")
	err = bootstrapController.JoinMoreWorkerPlanes(client, currWP, client.Metadata.NoWP)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	log.Success(controllerCtx, "successfully added workernodes")
	return nil
}

func (manager *KsctlControllerClient) DelWorkerPlaneNode() error {

	client := manager.client
	log := manager.log
	defer panicCatcher(log)

	if err := manager.setupConfigurations(); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	//if client.Metadata.IsHA && len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
	//	return log.NewError(controllerCtx,"This Functionality is diabled for {HA type cluster}", "FEATURE_FLAG", consts.KsctlFeatureFlagHaAutoscale)
	//}

	if client.Metadata.Provider == consts.CloudLocal {
		err := log.NewError(controllerCtx, "ha not supported for local")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if !client.Metadata.IsHA {
		err := log.NewError(controllerCtx, "this feature is only for ha clusters")
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {

		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationGet, fakeClient); err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	err := bootstrapController.Setup(client, stateDocument)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	hostnames, err := cloud.DelWorkerNodes(client)
	if err != nil {
		log.Error(controllerCtx, "handled error", "catch", err)
		return err
	}

	log.Debug(controllerCtx, "K8s nodes to be deleted", "hostnames", strings.Join(hostnames, ";"))

	if !fakeClient {
		var payload cloudControllerResource.CloudResourceState
		payload, err = client.Cloud.GetStateForHACluster(client.Storage)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}

		err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
		if err != nil {
			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}

		if err := bootstrapController.DelWorkerPlanes(
			client,
			stateDocument.ClusterKubeConfig,
			hostnames); err != nil {

			log.Error(controllerCtx, "handled error", "catch", err)
			return err
		}
	}
	log.Success(controllerCtx, "Successfully deleted workerNodes")

	return nil
}
