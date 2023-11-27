package main

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func main() {
	client := new(resources.KsctlClient)
	controller := control_pkg.GenKsctlController()

	client.Metadata.ClusterName = "example-cluster"
	client.Metadata.StateLocation = consts.StoreLocal
	client.Metadata.K8sDistro = consts.K8sK3s

	client.Metadata.K8sVersion = "1.27.1"
	client.Metadata.IsHA = true
	client.Metadata.Provider = consts.CloudAzure

	client.Metadata.NoWP = 2
	client.Metadata.NoCP = 2
	client.Metadata.NoDS = 1

	client.Metadata.WorkerPlaneNodeType = "Standard_F2s"
	client.Metadata.ControlPlaneNodeType = "Standard_F2s"
	client.Metadata.LoadBalancerNodeType = "Standard_F2s"
	client.Metadata.DataStoreNodeType = "Standard_F2s"

	client.Metadata.LogVerbosity = 0
	client.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(client); err != nil {
		panic(err)
	}

	err := controller.CreateHACluster(client)
	if err != nil {
		panic(err)
	}
}
