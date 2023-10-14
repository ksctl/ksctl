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
	cli.Client.Metadata.StateLocation = consts.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = consts.K8S_K3S

	cli.Client.Metadata.K8sVersion = "1.27.1"

	cli.Client.Metadata.Provider = consts.CLOUD_LOCAL

	cli.Client.Metadata.NoMP = 2

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, true); err != nil {
		panic(err)
	}

	_, err := controller.CreateManagedCluster(&cli.Client)
	if err != nil {
		panic(err)
	}
}
