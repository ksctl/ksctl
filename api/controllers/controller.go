package controllers

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/resources"
)

type KsctlControllerClient struct{}

func GenKsctlController() *KsctlControllerClient {
	return &KsctlControllerClient{}
}

// TODO: make the cloud related or kubernetes related call have a function parameter of statInterface
// by that we can use them inside that instead to binding to multiple different things
// we can simplify things and just share the interface itself
func (ksctlControlCli *KsctlControllerClient) CreateManagedCluster(*resources.KsctlClient) {}

func (ksctlControlCli *KsctlControllerClient) DeleteManagedCluster(*resources.KsctlClient) {}

func (ksctlControlCli *KsctlControllerClient) SwitchCluster() {}

func (ksctlControlCli *KsctlControllerClient) GetCluster() {}

func (ksctlControlCli *KsctlControllerClient) CreateHACluster(client *resources.KsctlClient) {
	fmt.Println("CreateHACLuster triggered successfully")
	cloud.HydrateCloud(client)
	fmt.Println("Cloud", client.Cloud)
	fmt.Println("Distro", client.Distro)
	fmt.Println("State", client.State)
	fmt.Println("Metadata", client.Metadata)
	// act as flag for provider to distingush whether the request was for HA
	client.Metadata.IsHA = true
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(*resources.KsctlClient) {}
