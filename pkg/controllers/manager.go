package controllers

import (
	"context"
	"os"
	"strings"

	"github.com/ksctl/ksctl/pkg/resources/controllers"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"

	awsPkg "github.com/ksctl/ksctl/internal/cloudproviders/aws"
	azurePkg "github.com/ksctl/ksctl/internal/cloudproviders/azure"
	civoPkg "github.com/ksctl/ksctl/internal/cloudproviders/civo"
	localPkg "github.com/ksctl/ksctl/internal/cloudproviders/local"

	"github.com/ksctl/ksctl/pkg/controllers/cloud"
	kubernetesController "github.com/ksctl/ksctl/pkg/controllers/kubernetes"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	cloudControllerResource "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

// TODO: need to pass the same context with no child creation to
// 1. cloud controller
// 2. bootstrap controller

var (
	controllerCtx context.Context
)

type KsctlControllerClient struct {
	log    resources.LoggerFactory
	client *resources.KsctlClient
}

func GenKsctlController(
	ctx context.Context,
	log resources.LoggerFactory,
	client *resources.KsctlClient,
) controllers.Controller {

	controllerCtx = context.WithValue(ctx, consts.ContextModuleNameKey, "ksctl-manager")

	return &KsctlControllerClient{
		log:    log,
		client: client,
	}
}

func (manager *KsctlControllerClient) setupConfigurations() error {

	if manager.client.Storage == nil {
		return manager.log.NewError(controllerCtx, "Initalize the storage driver")
	}
	if err := validationFields(manager.client.Metadata); err != nil {
		return manager.log.NewError(controllerCtx, "field validation failed", "Reason", err)
	}

	if err := helpers.IsValidName(manager.client.Metadata.ClusterName); err != nil {
		return err
	}
	return nil
}

func (manager *KsctlControllerClient) Applications(op consts.KsctlOperation) error {

	client := manager.client
	log := manager.log
	if err := manager.setupConfigurations(); err != nil {
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
		return err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationGet, fakeClient); err != nil {
		return err
	}

	if op != consts.OperationCreate && op != consts.OperationDelete {
		return log.NewError(controllerCtx, "Invalid operation")
	}

	return kubernetesController.ApplicationsInCluster(client, stateDocument, op)
}

func (manager *KsctlControllerClient) Credentials() error {
	log := manager.log
	client := manager.client

	if client.Storage == nil {
		return log.NewError(controllerCtx, "Initalize the storage driver")
	}

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
		return err
	}

	err = client.Cloud.Credential(client.Storage)
	if err != nil {
		return err
	}
	log.Success(controllerCtx, "Successfully Credential Added")

	return nil
}

func (manager *KsctlControllerClient) CreateManagedCluster() error {
	client := manager.client
	log := manager.log
	if err := manager.setupConfigurations(); err != nil {
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
		return log.NewError(controllerCtx, "invalid CNI plugin")
	}
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	if err := cloud.InitCloud(client, stateDocument, consts.OperationCreate, fakeClient); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}
	// it gets supportForApps, supportForCNI, error
	externalApp, externalCNI, cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		return log.NewError(controllerCtx, cloudResErr.Error())
	}

	if err := kubernetesController.InstallAdditionalTools(externalCNI, externalApp, client, stateDocument); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	log.Success(controllerCtx, "successfully created managed cluster")
	return nil
}

func (manager *KsctlControllerClient) DeleteManagedCluster() error {

	client := manager.client
	log := manager.log
	if err := manager.setupConfigurations(); err != nil {
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
		return err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()

	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.InitCloud(client, stateDocument, consts.OperationDelete, fakeClient); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	cloudResErr := cloud.DeleteManagedCluster(client)
	if cloudResErr != nil {
		return log.NewError(controllerCtx, cloudResErr.Error())
	}
	log.Success(controllerCtx, "successfully deleted managed cluster")
	return nil
}

func (manager *KsctlControllerClient) SwitchCluster() (*string, error) {
	client := manager.client
	log := manager.log
	if err := manager.setupConfigurations(); err != nil {
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
		return nil, err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

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
		return nil, err
	}

	if err := client.Cloud.IsPresent(client.Storage); err != nil {
		return nil, err
	}

	read, err := client.Storage.Read()
	if err != nil {
		return nil, err
	}
	log.Debug(controllerCtx, "data", "read", read)

	kubeconfig := read.ClusterKubeConfig
	log.Debug(controllerCtx, "data", "kubeconfig", kubeconfig)

	path, err := helpers.WriteKubeConfig(kubeconfig)
	log.Debug(controllerCtx, "data", "kubeconfigPath", path)
	if err != nil {
		return nil, err
	}

	printKubeConfig(path)

	return &kubeconfig, nil
}

func (manager *KsctlControllerClient) GetCluster() error {
	client := manager.client
	log := manager.log

	if client.Storage == nil {
		return log.NewError(controllerCtx, "Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(controllerCtx, err.Error())
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

	var printerTable []cloudControllerResource.AllClusterData
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		data, err := civoPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudLocal:
		data, err := localPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAws:
		data, err := awsPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAzure:
		data, err := azurePkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAll:
		data, err := civoPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

		data, err = localPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

		data, err = azurePkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
		printerTable = append(printerTable, data...)

		data, err = awsPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
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
	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError(controllerCtx, "ha not supported")
	}
	if err := manager.setupConfigurations(); err != nil {
		return err
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {
		return err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		return log.NewError(controllerCtx, "invalid CNI plugin")
	}

	if err := cloud.InitCloud(client, stateDocument, consts.OperationCreate, fakeClient); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}
	err := kubernetesController.Setup(client, stateDocument)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	cloudResErr := cloud.CreateHACluster(client)
	if cloudResErr != nil {
		return log.NewError(controllerCtx, cloudResErr.Error())
	}
	// Cloud done
	var payload cloudControllerResource.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

	err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationCreate)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	log.Note(controllerCtx, "only cloud resources are having replay!")

	externalCNI, err := kubernetesController.ConfigureCluster(client)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	if err := kubernetesController.InstallAdditionalTools(externalCNI, true, client, stateDocument); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	log.Success(controllerCtx, "successfully created ha cluster")

	return nil
}

func (manager *KsctlControllerClient) DeleteHACluster() error {

	client := manager.client
	log := manager.log
	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError(controllerCtx, "ha not supported")
	}
	if err := manager.setupConfigurations(); err != nil {
		return err
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {
		return err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)
	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.InitCloud(client, stateDocument, consts.OperationDelete, fakeClient); err != nil {
		return log.NewError(controllerCtx, err.Error())
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
		return log.NewError(controllerCtx, cloudResErr.Error())
	}

	log.Success(controllerCtx, "successfully deleted HA cluster")

	return nil
}

func (manager *KsctlControllerClient) AddWorkerPlaneNode() error {
	client := manager.client
	log := manager.log
	if err := manager.setupConfigurations(); err != nil {
		return err
	}

	//if client.Metadata.IsHA && len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
	//	// disable add AddWorkerPlaneNode when this feature is being used
	//	return log.NewError(controllerCtx,"This Functionality is diabled for {HA type clusters}", "FEATURE_FLAG", consts.KsctlFeatureFlagHaAutoscale)
	//}
	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError(controllerCtx, "ha not supported")
	}
	if !client.Metadata.IsHA {
		return log.NewError(controllerCtx, "this feature is only for ha clusters (for now)")
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {
		return err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.InitCloud(client, stateDocument, consts.OperationGet, fakeClient); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	err := kubernetesController.Setup(client, stateDocument)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	currWP, cloudResErr := cloud.AddWorkerNodes(client)
	if cloudResErr != nil {
		return log.NewError(controllerCtx, cloudResErr.Error())
	}

	// Cloud done
	var payload cloudControllerResource.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
	// transfer the state

	err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	log.Note(controllerCtx, "Only cloud resources are having replay!")
	err = kubernetesController.JoinMoreWorkerPlanes(client, currWP, client.Metadata.NoWP)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	log.Success(controllerCtx, "successfully added workernodes")
	return nil
}

func (manager *KsctlControllerClient) DelWorkerPlaneNode() error {

	client := manager.client
	log := manager.log
	if err := manager.setupConfigurations(); err != nil {
		return err
	}

	//if client.Metadata.IsHA && len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
	//	return log.NewError(controllerCtx,"This Functionality is diabled for {HA type cluster}", "FEATURE_FLAG", consts.KsctlFeatureFlagHaAutoscale)
	//}

	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError(controllerCtx, "ha not supported")
	}
	if !client.Metadata.IsHA {
		return log.NewError(controllerCtx, "this feature is only for ha clusters (for now)")
	}

	if err := client.Storage.Setup(
		client.Metadata.Provider,
		client.Metadata.Region,
		client.Metadata.ClusterName,
		consts.ClusterTypeHa); err != nil {
		return err
	}
	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error(controllerCtx, "StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)
	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.InitCloud(client, stateDocument, consts.OperationGet, fakeClient); err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	err := kubernetesController.Setup(client, stateDocument)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	hostnames, err := cloud.DelWorkerNodes(client)
	if err != nil {
		return log.NewError(controllerCtx, err.Error())
	}

	log.Debug(controllerCtx, "K8s nodes to be deleted", "hostnames", strings.Join(hostnames, ";"))
	if !fakeClient {
		var payload cloudControllerResource.CloudResourceState
		payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
		// transfer the state

		err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
		if err != nil {
			return log.NewError(controllerCtx, err.Error())
		}

		// move it to kubernetes controller
		if err := kubernetesController.DelWorkerPlanes(client, stateDocument.ClusterKubeConfig, hostnames); err != nil {
			return log.NewError(controllerCtx, err.Error())
		}
	}
	log.Success(controllerCtx, "Successfully deleted workerNodes")

	return nil
}
