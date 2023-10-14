package controllers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/utils"

	"github.com/kubesimplify/ksctl/internal/cloudproviders/azure"
	azurePkg "github.com/kubesimplify/ksctl/internal/cloudproviders/azure"
	localPkg "github.com/kubesimplify/ksctl/internal/cloudproviders/local"
	"github.com/kubesimplify/ksctl/internal/k8sdistros/universal"

	civoPkg "github.com/kubesimplify/ksctl/internal/cloudproviders/civo"

	localstate "github.com/kubesimplify/ksctl/internal/storagelogger/local"
	"github.com/kubesimplify/ksctl/pkg/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/controllers/kubernetes"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

type KsctlControllerClient struct{}

func GenKsctlController() *KsctlControllerClient {
	return &KsctlControllerClient{}
}

func InitializeStorageFactory(client *resources.KsctlClient, verbosity bool) (string, error) {
	switch client.Metadata.StateLocation {
	case StoreLocal:
		client.Storage = localstate.InitStorage(verbosity)
	default:
		return "", fmt.Errorf("Currently Local state is supported!")
	}
	return "[ksctl] initialized storageFactory", nil
}

func (ksctlControlCli *KsctlControllerClient) Credentials(client *resources.KsctlClient) (string, error) {

	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	switch client.Metadata.Provider {
	case CloudCivo:
		err := civoPkg.GetInputCredential(client.Storage)
		if err != nil {
			return "", err
		}
	case CloudAzure:
		err := azure.GetInputCredential(client.Storage)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("Currently not supported!")
	}
	return "[ksctl] Credential added", nil
}

func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(client *resources.KsctlClient) (string, error) {
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	fakeClient := false
	if str := os.Getenv(string(KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !utils.ValidCNIPlugin(KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		return "", errors.New("invalid CNI plugin")
	}

	if err := cloud.HydrateCloud(client, OperationStateCreate, fakeClient); err != nil {
		return "", err
	}
	// it gets supportForApps, supportForCNI, error
	externalApp, externalCNI, cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		client.Storage.Logger().Err(cloudResErr.Error())
	}

	kubeconfigPath := client.Cloud.GetKubeconfigPath()
	if len(os.Getenv(string(KsctlFeatureFlagApplications))) > 0 {

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return "", err
		}

		if externalCNI {
			if err := kubernetesClient.InstallCNI(client.Metadata.CNIPlugin); err != nil {
				return "", err
			}
		}

		if len(client.Metadata.Applications) != 0 && externalApp {
			apps := strings.Split(client.Metadata.Applications, ",")
			if err := kubernetesClient.InstallApplications(apps); err != nil {
				return "", err
			}
		}
	}
	return "[ksctl] created managed cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(client *resources.KsctlClient) (string, error) {

	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}
	fakeClient := false
	if str := os.Getenv(string(KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, OperationStateDelete, fakeClient); err != nil {
		return "", err
	}

	cloudResErr := cloud.DeleteManagedCluster(client)
	if cloudResErr != nil {
		client.Storage.Logger().Err(cloudResErr.Error())
	}
	return "[ksctl] deleted managed cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster(client *resources.KsctlClient) (string, error) {
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}
	var err error
	switch client.Metadata.Provider {
	case CloudCivo:
		client.Cloud, err = civoPkg.ReturnCivoStruct(client.Metadata, civoPkg.ProvideClient)
		if err != nil {
			return "", fmt.Errorf("[cloud] " + err.Error())
		}
	case CloudAzure:
		client.Cloud, err = azurePkg.ReturnAzureStruct(client.Metadata, azurePkg.ProvideClient)
		if err != nil {
			return "", fmt.Errorf("[cloud] " + err.Error())
		}
	case CloudLocal:
		client.Cloud, err = localPkg.ReturnLocalStruct(client.Metadata)
		if err != nil {
			return "", fmt.Errorf("[cloud] " + err.Error())
		}
	}

	if err := client.Cloud.SwitchCluster(client.Storage); err != nil {
		return "", err
	}
	return "[ksctl] switched cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) GetCluster(client *resources.KsctlClient) (string, error) {
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	var printerTable []cloudController.AllClusterData
	switch client.Metadata.Provider {
	case CloudCivo:
		data, err := civoPkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

	case CloudLocal:
		data, err := localPkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

	case CloudAzure:
		data, err := azurePkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

	case CloudAll:
		data, err := civoPkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

		data, err = localPkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

		data, err = azurePkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)
	}
	client.Storage.Logger().Table(printerTable)
	return "[ksctl] get clusters", nil
}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) (string, error) {
	if client.Metadata.Provider == CloudLocal {
		return "", fmt.Errorf("ha not supported")
	}

	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	fakeClient := false
	if str := os.Getenv(string(KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !utils.ValidCNIPlugin(KsctlValidCNIPlugin(client.Metadata.CNIPlugin)) {
		return "", errors.New("invalid CNI plugin")
	}

	if err := cloud.HydrateCloud(client, OperationStateCreate, fakeClient); err != nil {
		return "", err
	}
	err := kubernetes.HydrateK8sDistro(client)
	if err != nil {
		return "", err
	}

	cloudResErr := cloud.CreateHACluster(client)
	if cloudResErr != nil {
		return "", cloudResErr
	}
	// Cloud done
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

	err = client.Distro.InitState(payload, client.Storage, OperationStateCreate)
	if err != nil {
		return "", err
	}

	client.Storage.Logger().Warn("[ksctl] only cloud resources are having replay!\n")
	// Kubernetes controller
	externalCNI, err := kubernetes.ConfigureCluster(client)
	if err != nil {
		return "", err
	}

	//////// Done with cluster setup
	cloudstate, err := client.Cloud.GetStateFile(client.Storage)
	if err != nil {
		return "", err
	}

	k8sstate, err := client.Distro.GetStateFile(client.Storage)
	if err != nil {
		return "", err
	}

	kubeconfigPath, kubeconfig, err := client.Distro.GetKubeConfig(client.Storage)
	if err != nil {
		return "", err
	}

	var cloudSecret map[string][]byte
	cloudSecret, err = client.Cloud.GetSecretTokens(client.Storage)
	if err != nil {
		return "", err
	}

	////// EXPERIMENTAL Features //////
	if len(os.Getenv(string(KsctlFeatureFlagHaAutoscale))) > 0 {

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return "", err
		}

		if err = kubernetesClient.KsctlConfigForController(kubeconfig, kubeconfigPath, cloudstate, k8sstate, cloudSecret); err != nil {
			return "", err
		}
	}

	if len(os.Getenv(string(KsctlFeatureFlagApplications))) > 0 {

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return "", err
		}

		if externalCNI {
			if err := kubernetesClient.InstallCNI(client.Metadata.CNIPlugin); err != nil {
				return "", err
			}
		}

		if len(client.Metadata.Applications) != 0 {
			apps := strings.Split(client.Metadata.Applications, ",")
			if err := kubernetesClient.InstallApplications(apps); err != nil {
				return "", err
			}
		}
	}

	return "[ksctl] created HA cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(client *resources.KsctlClient) (string, error) {

	if client.Metadata.Provider == CloudLocal {
		return "", fmt.Errorf("ha not supported")
	}
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	fakeClient := false
	if str := os.Getenv(string(KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, OperationStateDelete, fakeClient); err != nil {
		return "", err
	}

	if len(os.Getenv(string(KsctlFeatureFlagHaAutoscale))) > 0 {

		// find a better way to get the kubeconfig location

		err := kubernetes.HydrateK8sDistro(client)
		if err != nil {
			return "", err
		}
		var payload cloudController.CloudResourceState
		payload, _ = client.Cloud.GetStateForHACluster(client.Storage)

		err = client.Distro.InitState(payload, client.Storage, OperationStateGet)
		if err != nil {
			return "", err
		}
		kubeconfigPath, _, err := client.Distro.GetKubeConfig(client.Storage)
		if err != nil {
			return "", err
		}

		kubernetesClient := universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
			return "", err
		}

		if err = kubernetesClient.DeleteResourcesFromController(); err != nil {
			return "", err
		}

		// NOTE: explict make the count of the workernodes as 0 as we need one schedulable workload to test of the operation was successful
		if _, err := client.Cloud.NoOfWorkerPlane(client.Storage, 0, true); err != nil {
			return "", err
		}
	}

	cloudResErr := cloud.DeleteHACluster(client)
	if cloudResErr != nil {
		return "", cloudResErr
	}

	return "[ksctl] deleted HA cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) AddWorkerPlaneNode(client *resources.KsctlClient) (string, error) {

	if client.Metadata.IsHA && len(os.Getenv(string(KsctlFeatureFlagHaAutoscale))) > 0 {
		// disable add AddWorkerPlaneNode when this feature is being used
		return "", fmt.Errorf("This Functionality is diabled for {HA type clusters} due to FEATURE_FLAG [%s]", KsctlFeatureFlagHaAutoscale)
	}
	if client.Metadata.Provider == CloudLocal {
		return "", fmt.Errorf("ha not supported")
	}
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}
	if !client.Metadata.IsHA {
		return "", fmt.Errorf("this feature is only for ha clusters (for now)")
	}

	fakeClient := false
	if str := os.Getenv(string(KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, OperationStateGet, fakeClient); err != nil {
		return "", err
	}

	err := kubernetes.HydrateK8sDistro(client)
	if err != nil {
		return "", err
	}

	currWP, cloudResErr := cloud.AddWorkerNodes(client)
	if cloudResErr != nil {
		return "", cloudResErr
	}

	// Cloud done
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
	// transfer the state

	err = client.Distro.InitState(payload, client.Storage, OperationStateGet)
	if err != nil {
		return "", err
	}

	client.Storage.Logger().Warn("\n[ksctl] only cloud resources are having replay!\n")
	// Kubernetes controller
	err = kubernetes.JoinMoreWorkerPlanes(client, currWP, client.Metadata.NoWP)
	if err != nil {
		return "", err
	}

	return "[ksctl] added worker node(s)", nil
}

func (ksctlControlCli *KsctlControllerClient) DelWorkerPlaneNode(client *resources.KsctlClient) (string, error) {

	if client.Metadata.IsHA && len(os.Getenv(string(KsctlFeatureFlagHaAutoscale))) > 0 {
		return "", fmt.Errorf("This Functionality is diabled for {HA type cluster} due to FEATURE_FLAG [%s]", KsctlFeatureFlagHaAutoscale)
	}

	if client.Metadata.Provider == CloudLocal {
		return "", fmt.Errorf("ha not supported")
	}
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}
	if !client.Metadata.IsHA {
		return "", fmt.Errorf("this feature is only for ha clusters (for now)")
	}

	fakeClient := false
	if str := os.Getenv(string(KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}
	if err := cloud.HydrateCloud(client, OperationStateGet, fakeClient); err != nil {
		return "", err
	}

	err := kubernetes.HydrateK8sDistro(client)
	if err != nil {
		return "", err
	}

	hostnames, err := cloud.DelWorkerNodes(client)
	if err != nil {
		return "", err
	}

	client.Storage.Logger().Note("Hostnames to remove", strings.Join(hostnames, ";"))

	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
	// transfer the state

	err = client.Distro.InitState(payload, client.Storage, OperationStateGet)
	if err != nil {
		return "", err
	}

	// move it to kubernetes controller
	if err := kubernetes.DelWorkerPlanes(client, hostnames); err != nil {
		return "", err
	}

	return "[ksctl] deleted worker node(s)", nil
}
