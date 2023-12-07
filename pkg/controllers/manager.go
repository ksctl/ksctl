package controllers

import (
	"errors"
	"github.com/kubesimplify/ksctl/pkg/helpers"
	"os"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/kubesimplify/ksctl/internal/cloudproviders/azure"
	azurePkg "github.com/kubesimplify/ksctl/internal/cloudproviders/azure"
	localPkg "github.com/kubesimplify/ksctl/internal/cloudproviders/local"
	"github.com/kubesimplify/ksctl/internal/k8sdistros/universal"

	civoPkg "github.com/kubesimplify/ksctl/internal/cloudproviders/civo"

	localstate "github.com/kubesimplify/ksctl/internal/storage/local"
	"github.com/kubesimplify/ksctl/pkg/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/controllers/kubernetes"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
)

var (
	log resources.LoggerFactory
)

type KsctlControllerClient struct{}

func GenKsctlController() *KsctlControllerClient {
	return &KsctlControllerClient{}
}

func validationFields(meta resources.Metadata) error {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName("ksctl-manager")

	if !helpers.ValidateCloud(meta.Provider) {
		return errors.New("invalid cloud provider")
	}
	if !helpers.ValidateDistro(meta.K8sDistro) {
		return errors.New("invalid kubernetes distro")
	}
	if !helpers.ValidateStorage(meta.StateLocation) {
		return errors.New("invalid storage driver")
	}
	log.Debug("Valid fields from user")
	return nil
}

func InitializeStorageFactory(client *resources.KsctlClient) error {

	if log == nil {
		log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
		log.SetPackageName("ksctl-manager")
	}

	switch client.Metadata.StateLocation {
	case consts.StoreLocal:
		client.Storage = localstate.InitStorage()
	default:
		return log.NewError("Currently Local state is supported!")
	}
	log.Debug("initialized storageFactory")
	return nil
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

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		return log.NewError("invalid CNI plugin")
	}

	if err := cloud.HydrateCloud(client, consts.OperationStateCreate, fakeClient); err != nil {
		return log.NewError(err.Error())
	}
	// it gets supportForApps, supportForCNI, error
	externalApp, externalCNI, cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		log.Error(cloudResErr.Error())
	}

	kubeconfigPath := client.Cloud.GetKubeconfigPath()
	if len(os.Getenv(string(consts.KsctlFeatureFlagApplications))) > 0 {

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return log.NewError(err.Error())
		}

		if externalCNI {
			if err := kubernetesClient.InstallCNI(client.Metadata.CNIPlugin); err != nil {
				return log.NewError(err.Error())
			}
		}

		if len(client.Metadata.Applications) != 0 && externalApp {
			apps := strings.Split(client.Metadata.Applications, ",")
			if err := kubernetesClient.InstallApplications(apps); err != nil {
				return log.NewError(err.Error())
			}
		}
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

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, consts.OperationStateDelete, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	cloudResErr := cloud.DeleteManagedCluster(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}
	log.Success("successfully deleted managed cluster")
	return nil
}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster(client *resources.KsctlClient) error {
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		client.Cloud, err = civoPkg.ReturnCivoStruct(client.Metadata, civoPkg.ProvideClient)
		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.CloudAzure:
		client.Cloud, err = azurePkg.ReturnAzureStruct(client.Metadata, azurePkg.ProvideClient)
		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.CloudLocal:
		client.Cloud, err = localPkg.ReturnLocalStruct(client.Metadata, localPkg.ProvideClient)
		if err != nil {
			return log.NewError(err.Error())
		}
	}

	if err := client.Cloud.SwitchCluster(client.Storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("successfully switched cluster")
	return nil
}

func (ksctlControlCli *KsctlControllerClient) GetCluster(client *resources.KsctlClient) error {
	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}
	if err := validationFields(client.Metadata); err != nil {
		return log.NewError(err.Error())
	}

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

	if client.Storage == nil {
		return log.NewError("Initalize the storage driver")
	}

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !helpers.ValidCNIPlugin(consts.KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		return log.NewError("invalid CNI plugin")
	}

	if err := cloud.HydrateCloud(client, consts.OperationStateCreate, fakeClient); err != nil {
		return log.NewError(err.Error())
	}
	err := kubernetes.HydrateK8sDistro(client)
	if err != nil {
		return log.NewError(err.Error())
	}

	cloudResErr := cloud.CreateHACluster(client)
	if cloudResErr != nil {
		return log.NewError(cloudResErr.Error())
	}
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

	err = client.Distro.InitState(payload, client.Storage, consts.OperationStateCreate)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Warn("only cloud resources are having replay!")
	// Kubernetes controller
	externalCNI, err := kubernetes.ConfigureCluster(client)
	if err != nil {
		return log.NewError(err.Error())
	}

	cloudstate, err := client.Cloud.GetStateFile(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	k8sstate, err := client.Distro.GetStateFile(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	kubeconfigPath, kubeconfig, err := client.Distro.GetKubeConfig(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	var cloudSecret map[string][]byte
	cloudSecret, err = client.Cloud.GetSecretTokens(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	////// EXPERIMENTAL Features //////
	if len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return log.NewError(err.Error())
		}

		if err = kubernetesClient.KsctlConfigForController(kubeconfig, kubeconfigPath, cloudstate, k8sstate, cloudSecret); err != nil {
			return log.NewError(err.Error())
		}
	}

	if len(os.Getenv(string(consts.KsctlFeatureFlagApplications))) > 0 {

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return log.NewError(err.Error())
		}

		if externalCNI {
			if err := kubernetesClient.InstallCNI(client.Metadata.CNIPlugin); err != nil {
				return log.NewError(err.Error())
			}
		}

		if len(client.Metadata.Applications) != 0 {
			apps := strings.Split(client.Metadata.Applications, ",")
			if err := kubernetesClient.InstallApplications(apps); err != nil {
				return log.NewError(err.Error())
			}
		}
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

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, consts.OperationStateDelete, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	if len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {

		// find a better way to get the kubeconfig location

		err := kubernetes.HydrateK8sDistro(client)
		if err != nil {
			return log.NewError(err.Error())
		}
		var payload cloudController.CloudResourceState
		payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

		err = client.Distro.InitState(payload, client.Storage, consts.OperationStateGet)
		if err != nil {
			return log.NewError(err.Error())
		}
		kubeconfigPath, _, err := client.Distro.GetKubeConfig(client.Storage)
		if err != nil {
			return log.NewError(err.Error())
		}

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
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

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, consts.OperationStateGet, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	err := kubernetes.HydrateK8sDistro(client)
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

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, consts.OperationStateGet, fakeClient); err != nil {
		return log.NewError(err.Error())
	}

	err := kubernetes.HydrateK8sDistro(client)
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
		if err := kubernetes.DelWorkerPlanes(client, hostnames); err != nil {
			return log.NewError(err.Error())
		}
	}
	log.Success("Successfully deleted workerNodes")

	return nil
}
