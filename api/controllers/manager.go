package controllers

import (
	"fmt"
	"strings"

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

func (ksctlControlCli *KsctlControllerClient) Credentials(client *resources.KsctlClient) {
	switch client.Metadata.StateLocation {
	case "local":
		client.Storage = localstate.InitStorage()
	default:
		panic("Currently Local state is supported!")
	}

	switch client.Metadata.Provider {
	case "civo":
		err := civo_pkg.GetInputCredential(client.Storage)
		if err != nil {
			panic(err)
		}
	case "azure", "aws":
		panic("Currently not supported!")
	default:
		panic("Currently not supported!")
	}
}

func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(client *resources.KsctlClient) {
	fmt.Println("Create Managed Cluster triggered successfully")
	switch client.Metadata.StateLocation {
	case "local":
		client.Storage = localstate.InitStorage()
	default:
		panic("Currently Local state is supported!")
	}

	cloud.HydrateCloud(client, "create")

	cloudResErr := cloud.CreateManagedCluster(client)
	if cloudResErr != nil {
		panic(cloudResErr)
	}
	client.Storage.Logger().Success("[ksctl] Created the managed cluster")
}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(client *resources.KsctlClient) {
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
			return
		}
	}
	switch client.Metadata.StateLocation {
	case "local":
		client.Storage = localstate.InitStorage()
	default:
		panic("Currently Local state is supported!")
	}
	cloud.HydrateCloud(client, "delete")
	cloudResErr := cloud.DeleteManagedCluster(client)
	if cloudResErr != nil {
		panic(cloudResErr)
	}
	client.Storage.Logger().Success("[ksctl] Deleted the managed cluster")
}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster() {}

func (ksctlControlCli *KsctlControllerClient) GetCluster(client *resources.KsctlClient) {
	switch client.Metadata.StateLocation {
	case "local":
		client.Storage = &localstate.LocalStorageProvider{}
	default:
		panic("Currently Local state is supported!")
	}

	var printerTable []cloudController.AllClusterData
	switch client.Metadata.Provider {
	case "civo":
		data, err := civo_pkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			panic(err)
		}
		printerTable = append(printerTable, data...)
	case "azure", "aws", "local":
		panic("Not yet implemtned")
	case "all":
		data, err := civo_pkg.GetRAWClusterInfos(client.Storage)
		if err != nil {
			panic(err)
		}
		printerTable = append(printerTable, data...)
	}
	fmt.Println(printerTable)
}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) {
	fmt.Println("Create HA Cluster triggered successfully")
	// Builder methods directly called
	cloud.HydrateCloud(client, "create")

	kubernetes.HydrateK8sDistro(client)

	switch client.Metadata.StateLocation {
	case "local":
		client.Storage = &localstate.LocalStorageProvider{}
	default:
		panic("Currently Local state is supported!")
	}

	cloudResErr := cloud.CreateHACluster(client)
	fmt.Println("Called Create Cloud resources for HA setup; Err->", cloudResErr)

	// Cloud done
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.Storage)
	// transfer the state
	client.Distro.InitState(payload)

	// Kubernetes controller
	kubernetes.ConfigureCluster(client)
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(client *resources.KsctlClient) {
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
			return
		}
	}
	fmt.Println("Create HA delete triggered successfully")
	switch client.Metadata.StateLocation {
	case "local":
		client.Storage = &localstate.LocalStorageProvider{}
	default:
		panic("Currently Local state is supported!")
	}
	cloud.HydrateCloud(client, "delete")

	kubernetes.HydrateK8sDistro(client)

	cloudResErr := cloud.DeleteHACluster(client)
	fmt.Println("Called Delete Cloud resources for HA setup; Err->", cloudResErr)
}
