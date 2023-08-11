package controllers

import (
	"fmt"
	local_pkg "github.com/kubesimplify/ksctl/api/provider/local"
	"strings"

	"github.com/kubesimplify/ksctl/api/utils"

	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"

	"github.com/kubesimplify/ksctl/api/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/controllers/kubernetes"
	"github.com/kubesimplify/ksctl/api/resources"
	cloudController "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
)

type KsctlControllerClient struct{}

func GenKsctlController() *KsctlControllerClient {
	return &KsctlControllerClient{}
}

func InitializeStorageFactory(client *resources.KsctlClient, verbosity bool) (string, error) {
	switch client.Metadata.StateLocation {
	case utils.CLOUD_LOCAL:
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
	case utils.CLOUD_CIVO:
		err := civo_pkg.GetInputCredential(client.Storage)
		if err != nil {
			return "", err
		}
		return "[ksctl] Credential added", nil
	case utils.CLOUD_AZURE:
		return "", fmt.Errorf("Currently not supported!")
	default:
		return "", fmt.Errorf("Currently not supported!")
	}
}

// TODO: accept gate for creation as well for approval
// FIXME: move these scanf and approval to the cli it shouldn't be here

func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(client *resources.KsctlClient) (string, error) {
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	if err := cloud.HydrateCloud(client, utils.OPERATION_STATE_CREATE); err != nil {
		return "", err
	}

	cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		client.Storage.Logger().Err(cloudResErr.Error())
	}
	return "[ksctl] created managed cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(client *resources.KsctlClient) (string, error) {

	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	showMsg := true
	if showMsg {
		fmt.Println(fmt.Sprintf(`🚨 THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'
	`, client.ClusterName+" "+client.Region))

		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return "[ksctl] cancelled", nil
		}
	}

	if err := cloud.HydrateCloud(client, utils.OPERATION_STATE_DELETE); err != nil {
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
	return "[ksctl] switched cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) GetCluster(client *resources.KsctlClient) (string, error) {
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	var printerTable []cloudController.AllClusterData
	switch client.Metadata.Provider {
	case utils.CLOUD_CIVO:
		data, err := civo_pkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

	case utils.CLOUD_LOCAL:
		data, err := local_pkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

	case utils.CLOUD_AZURE:
		return "", fmt.Errorf("Not yet implemtned")

	case "all":
		data, err := civo_pkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)

		data, err = local_pkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			return "", err
		}
		printerTable = append(printerTable, data...)
	}
	client.Storage.Logger().Table(printerTable)
	return "[ksctl] get clusters", nil
}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) (string, error) {
	if client.Provider == utils.CLOUD_LOCAL {
		return "", fmt.Errorf("ha not supported")
	}

	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	if err := cloud.HydrateCloud(client, utils.OPERATION_STATE_CREATE); err != nil {
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

	err = client.Distro.InitState(payload, client.Storage, utils.OPERATION_STATE_CREATE)
	if err != nil {
		return "", err
	}

	client.Storage.Logger().Warn("[ksctl] only cloud resources are having replay!\n")
	// Kubernetes controller
	err = kubernetes.ConfigureCluster(client)
	if err != nil {
		return "", err
	}
	return "[ksctl] created HA cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(client *resources.KsctlClient) (string, error) {

	if client.Provider == utils.CLOUD_LOCAL {
		return "", fmt.Errorf("ha not supported")
	}
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}

	showMsg := true
	if showMsg {
		fmt.Println(fmt.Sprintf(`🚨 THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'
	`, client.ClusterName+" "+client.Region))

		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return "[ksctl] approval cancelled", nil
		}
	}
	if err := cloud.HydrateCloud(client, utils.OPERATION_STATE_DELETE); err != nil {
		return "", err
	}

	cloudResErr := cloud.DeleteHACluster(client)
	if cloudResErr != nil {
		return "", cloudResErr
	}
	return "[ksctl] deleted HA cluster", nil
}

func (ksctlControlCli *KsctlControllerClient) AddWorkerPlaneNode(client *resources.KsctlClient) (string, error) {
	if client.Provider == utils.CLOUD_LOCAL {
		return "", fmt.Errorf("ha not supported")
	}
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}
	if !client.IsHA {
		return "", fmt.Errorf("this feature is only for ha clusters (for now)")
	}
	if err := cloud.HydrateCloud(client, utils.OPERATION_STATE_GET); err != nil {
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

	err = client.Distro.InitState(payload, client.Storage, utils.OPERATION_STATE_GET)
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
	if client.Provider == utils.CLOUD_LOCAL {
		return "", fmt.Errorf("ha not supported")
	}
	if client.Storage == nil {
		return "", fmt.Errorf("Initalize the storage driver")
	}
	if !client.IsHA {
		return "", fmt.Errorf("this feature is only for ha clusters (for now)")
	}
	showMsg := true
	if showMsg {
		fmt.Println(fmt.Sprintf(`🚨 THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'
	`, client.ClusterName+" "+client.Region))

		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return "[ksctl] approval cancelled", nil
		}
	}

	if err := cloud.HydrateCloud(client, utils.OPERATION_STATE_GET); err != nil {
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

	err = client.Distro.InitState(payload, client.Storage, utils.OPERATION_STATE_GET)
	if err != nil {
		return "", err
	}

	// move it to kubernetes controller
	if err := kubernetes.DelWorkerPlanes(client, hostnames); err != nil {
		return "", err
	}

	return "[ksctl] deleted worker node(s)", nil
}
