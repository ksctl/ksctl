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

func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(*resources.KsctlClient) {}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(*resources.KsctlClient) {}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster() {}

func (ksctlControlCli *KsctlControllerClient) GetCluster() {}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) {
	fmt.Println("CreateHACLuster triggered successfully")
	cloud.HydrateCloud(client)
	client.State = &localstate.LocalStorageProvider{}
	kubernetes.HydrateK8sDistro(client)
	fmt.Println("Cloud", client.Cloud)
	fmt.Println("Distro", client.Distro)
	fmt.Println("State", client.State)
	fmt.Println("Metadata", client.Metadata)
	fmt.Println("Callled Create Cloud resources for HA setup", cloud.CreateHACluster(client))

	var payload cloudController.CloudResourceState
	payload, _ = client.Cloud.GetStateForHACluster(client.State)
	client.Distro.InitState(payload)
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(*resources.KsctlClient) {}
