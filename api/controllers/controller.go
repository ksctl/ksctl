package controllers

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/resources"
)

// func NewController(client *resources.Builder) {
// 	ksctlCloudAPI := cloudController.WrapCloudEngineBuilder(client)
// 	abcd := cloudController.NewController(ksctlCloudAPI)
// 	reqForK8sDistro := abcd.FetchState()

// 	ksctlK8sAPI := k8sController.WrapK8sEngineBuilder(client)
// 	k8sController.NewController(ksctlK8sAPI, reqForK8sDistro)
// }

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
	fmt.Println("Cloud", client.Cloud)
	fmt.Println("Distro", client.Distro)
	fmt.Println("State", client.State)
	fmt.Println("Metadata", client.Metadata)
	// act as flag for provider to distingush whether the request was for HA
	client.Metadata.IsHA = true
}

func (ksctlControlCli *KsctlControllerClient) DeleteHACluster(*resources.KsctlClient) {}
