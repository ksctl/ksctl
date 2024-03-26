package controllers

import (
	"os"
	"strings"
	"time"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"

	awsPkg "github.com/ksctl/ksctl/internal/cloudproviders/aws"
	"github.com/ksctl/ksctl/internal/cloudproviders/azure"
	azurePkg "github.com/ksctl/ksctl/internal/cloudproviders/azure"
	localPkg "github.com/ksctl/ksctl/internal/cloudproviders/local"
	"github.com/ksctl/ksctl/internal/k8sdistros/universal"

	civoPkg "github.com/ksctl/ksctl/internal/cloudproviders/civo"

	"github.com/ksctl/ksctl/pkg/controllers/cloud"
	"github.com/ksctl/ksctl/pkg/controllers/kubernetes"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	cloudController "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	log resources.LoggerFactory
)

type KsctlControllerClient struct{}

func GenKsctlController() *KsctlControllerClient {
	return &KsctlControllerClient{}
}

func (ksctlControlCli *KsctlControllerClient) Credentials(client *resources.KsctlClient) error {

	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}

	switch client.Metadata.Provider {
	case consts.CloudCivo:
		err := civoPkg.GetInputCredential(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.CloudAzure:
		err := azure.GetInputCredential(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.CloudAws:
		err := awsPkg.GetInputCredential(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
	default:
		return log.NewError("Currently not supported!")
	}
	log.Success("successfully Credential Added")

	return nil
}

func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(client *resources.KsctlClient) error {
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}
	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
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
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		return log.NewError("invalid CNI plugin")
	}
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	if err := cloud.HydrateCloud(client, stateDocument, consts.OperationStateCreate, fakeClient); err != nil {
		return log.NewError(err.Error())
	}
	// it gets supportForApps, supportForCNI, error
	externalApp, externalCNI, cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}

	kubeconfig := stateDocument.ClusterKubeConfig

	if err := kubernetes.InstallAdditionalTools(kubeconfig, externalCNI, externalApp, client, stateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("successfully created managed cluster")
	return nil
}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(client *resources.KsctlClient) error {

	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
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
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, stateDocument, consts.OperationStateDelete, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	cloudResErr := cloud.DeleteManagedCluster(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}
	log.Success("successfully deleted managed cluster")
	return nil
}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster(client *resources.KsctlClient) (*string, error) {
	if client.Storage == nil {
		return nil, log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return nil, log.NewError(err.Error())
	}

	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
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
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	var err error
	switch client.Metadata.Provider {
	case consts.CloudAws:
		client.Cloud, err = awsPkg.ReturnAwsStruct(client.Metadata, stateDocument, awsPkg.ProvideClient)
		if err != nil {
			return nil, log.NewError(err.Error())
		}
	case consts.CloudCivo:
		client.Cloud, err = civoPkg.ReturnCivoStruct(client.Metadata, stateDocument, civoPkg.ProvideClient)
		if err != nil {
			return nil, log.NewError(err.Error())
		}
	case consts.CloudAzure:
		client.Cloud, err = azurePkg.ReturnAzureStruct(client.Metadata, stateDocument, azurePkg.ProvideClient)
		if err != nil {
			return nil, log.NewError(err.Error())
		}
	case consts.CloudLocal:
		client.Cloud, err = localPkg.ReturnLocalStruct(client.Metadata, stateDocument, localPkg.ProvideClient)
		if err != nil {
			return nil, log.NewError(err.Error())
		}
	}

	if err := client.Cloud.IsPresent(client.Storage); err != nil {
		return nil, err
	}

	read, err := client.Storage.Read()
	if err != nil {
		return nil, err
	}
	log.Debug("data", "read", read)

	kubeconfig := read.ClusterKubeConfig
	log.Debug("data", "kubeconfig", kubeconfig)
	path, err := helpers.WriteKubeConfig(kubeconfig)
	log.Debug("data", "kubeconfigPath", path)
	if err != nil {
		return nil, err
	}

	printKubeConfig(path)

	return &kubeconfig, nil
}

func (ksctlControlCli *KsctlControllerClient) GetCluster(client *resources.KsctlClient) error {
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	if client.Metadata.Provider == consts.CloudLocal {
		client.Metadata.Region = "LOCAL"
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	log.Note("Filter", "cloudProvider", string(client.Metadata.Provider))

	var printerTable []cloudController.AllClusterData
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		data, err := civoPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudLocal:
		data, err := localPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAws:
		data, err := awsPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAzure:
		data, err := azurePkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

	case consts.CloudAll:
		data, err := civoPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

		data, err = localPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

		data, err = azurePkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)

		data, err = awsPkg.GetRAWClusterInfos(client.Storage, client.Metadata)
		if err != nil {
			return log.NewError(err.Error())
		}
		printerTable = append(printerTable, data...)
	}
	log.Table(printerTable)

	log.Success("successfully get clusters")

	return nil
}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) error {
	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError("ha not supported")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
		return err
	}

	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
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
			log.Error("StorageClass Kill failed", "reason", err)
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
		return log.NewError("invalid CNI plugin")
	}

	if err := cloud.HydrateCloud(client, stateDocument, consts.OperationStateCreate, fakeClient); err != nil {
		return log.NewError(err.Error())
	}
	err := kubernetes.HydrateK8sDistro(client, stateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}

	cloudResErr := cloud.CreateHACluster(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}
	// Cloud done
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

	err = client.Distro.InitState(payload, client.Storage, consts.OperationStateCreate)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Warn("only cloud resources are having replay!")

	time.Sleep(30 * time.Second) // hack to wait for all the cloud specific resources to be in consistent state
	// Kubernetes controller
	externalCNI, err := kubernetes.ConfigureCluster(client)
	if err != nil {
		return log.NewError(err.Error())
	}

	kubeconfig := stateDocument.ClusterKubeConfig

	if err := kubernetes.InstallAdditionalTools(kubeconfig, externalCNI, true, client, stateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("successfully created ha cluster")

	return nil
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(client *resources.KsctlClient) error {

	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError("ha not supported")
	}
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
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
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)
	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, stateDocument, consts.OperationStateDelete, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	// TODO: move it kubernetes controller
	// FIXME: do we actually need it any more as storage driver can be in a single place thus no need to write the below magic
	if len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {

		// find a better way to get the kubeconfig location

		err := kubernetes.HydrateK8sDistro(client, stateDocument)
		if err != nil {
			return log.NewError(err.Error())
		}
		var payload cloudController.CloudResourceState
		payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

		err = client.Distro.InitState(payload, client.Storage, consts.OperationStateGet)
		if err != nil {
			return log.NewError(err.Error())
		}

		kubeconfig := stateDocument.ClusterKubeConfig

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfig); err != nil {
			return log.NewError(err.Error())
		}

		if err = kubernetesClient.DeleteResourcesFromController(); err != nil {
			return log.NewError(err.Error())
		}

		// NOTE: explict make the count of the workernodes as 0 as we need one schedulable workload to test of the operation was successful
		if _, err := client.Cloud.NoOfWorkerPlane(client.Storage, 0, true); err != nil {
			return log.NewError(err.Error())
		}
	}

	cloudResErr := cloud.DeleteHACluster(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}

	log.Success("successfully deleted HA cluster")

	return nil
}

func (ksctlControlCli *KsctlControllerClient) AddWorkerPlaneNode(client *resources.KsctlClient) error {
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
		return err
	}

	if client.Metadata.IsHA && len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
		// disable add AddWorkerPlaneNode when this feature is being used
		return log.NewError("This Functionality is diabled for {HA type clusters}", "FEATURE_FLAG", consts.KsctlFeatureFlagHaAutoscale)
	}
	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError("ha not supported")
	}
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if !client.Metadata.IsHA {
		return log.NewError("this feature is only for ha clusters (for now)")
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
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, stateDocument, consts.OperationStateGet, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	err := kubernetes.HydrateK8sDistro(client, stateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}

	currWP, cloudResErr := cloud.AddWorkerNodes(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}

	// Cloud done
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
	// transfer the state

	err = client.Distro.InitState(payload, client.Storage, consts.OperationStateGet)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Warn("[ksctl] only cloud resources are having replay!")
	// Kubernetes controller
	err = kubernetes.JoinMoreWorkerPlanes(client, currWP, client.Metadata.NoWP)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("successfully added workernodes")
	return nil
}

func (ksctlControlCli *KsctlControllerClient) DelWorkerPlaneNode(client *resources.KsctlClient) error {
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	if err := helpers.IsValidName(client.Metadata.ClusterName); err != nil {
		return err
	}

	if client.Metadata.IsHA && len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {
		return log.NewError("This Functionality is diabled for {HA type cluster}", "FEATURE_FLAG", consts.KsctlFeatureFlagHaAutoscale)
	}

	if client.Metadata.Provider == consts.CloudLocal {
		return log.NewError("ha not supported")
	}
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if !client.Metadata.IsHA {
		return log.NewError("this feature is only for ha clusters (for now)")
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
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()
	var (
		stateDocument *types.StorageDocument = &types.StorageDocument{}
	)
	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, stateDocument, consts.OperationStateGet, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	err := kubernetes.HydrateK8sDistro(client, stateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}

	hostnames, err := cloud.DelWorkerNodes(client)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Debug("K8s nodes to be deleted", "hostnames", strings.Join(hostnames, ";"))
	if !fakeClient {
		var payload cloudController.CloudResourceState
		payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
		// transfer the state

		err = client.Distro.InitState(payload, client.Storage, consts.OperationStateGet)
		if err != nil {
			return log.NewError(err.Error())
		}

		// move it to kubernetes controller
		if err := kubernetes.DelWorkerPlanes(client, stateDocument.ClusterKubeConfig, hostnames); err != nil {
			return log.NewError(err.Error())
		}
	}
	log.Success("Successfully deleted workerNodes")

	return nil
}
