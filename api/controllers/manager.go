package controllers

import (
	"fmt"

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

func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(client *resources.KsctlClient) {
	fmt.Println("Create HA Cluster triggered successfully")
	// Builder methods directly called
	cloud.HydrateCloud(client)
	kubernetes.HydrateK8sDistro(client)
	switch client.Metadata.StateLocation {
	case "local":
		client.State = localstate.InitStorage()
	default:
		panic("Currently Local state is supported!")
	}

	cloudResErr := cloud.CreateManagedCluster(client)
	fmt.Println("Called Create Cloud managed cluster; Err->", cloudResErr)
}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(*resources.KsctlClient) {}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster() {}

func (ksctlControlCli *KsctlControllerClient) GetCluster() {}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) {
	fmt.Println("Create HA Cluster triggered successfully")
	// Builder methods directly called
	cloud.HydrateCloud(client)
	kubernetes.HydrateK8sDistro(client)
	switch client.Metadata.StateLocation {
	case "local":
		client.State = &localstate.LocalStorageProvider{}
	default:
		panic("Currently Local state is supported!")
	}

	cloudResErr := cloud.CreateHACluster(client)
	fmt.Println("Callled Create Cloud resources for HA setup; Err->", cloudResErr)

	// Cloud done
	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.State)
	// transfer the state
	client.Distro.InitState(payload)

	// Kubernetes controller
	kubernetes.ConfigureCluster(client)
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(*resources.KsctlClient) {}
