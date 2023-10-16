package main

import (
	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func main() {
	cli := new(resources.CobraCmd)
	controller := control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = "example-cluster"
	cli.Client.Metadata.StateLocation = consts.StoreLocal
	cli.Client.Metadata.K8sDistro = consts.K8sK3s

	cli.Client.Metadata.K8sVersion = "1.27.1"
	cli.Client.Metadata.IsHA = true
	cli.Client.Metadata.Provider = consts.CloudAzure

	cli.Client.Metadata.NoWP = 2
	cli.Client.Metadata.NoCP = 2
	cli.Client.Metadata.NoDS = 1

	cli.Client.Metadata.WorkerPlaneNodeType = "Standard_F2s"
	cli.Client.Metadata.ControlPlaneNodeType = "Standard_F2s"
	cli.Client.Metadata.LoadBalancerNodeType = "Standard_F2s"
	cli.Client.Metadata.DataStoreNodeType = "Standard_F2s"

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		panic(err)
	}

	_, err := controller.CreateHACluster(&cli.Client)
	if err != nil {
		panic(err)
	}
}
