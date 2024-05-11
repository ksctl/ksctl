package main

import (
	"context"
	"os"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func main() {
	client := new(types.KsctlClient)
	controller := control_pkg.GenKsctlController()

	client.Metadata.ClusterName = "example-cluster"
	client.Metadata.StateLocation = consts.StoreLocal
	client.Metadata.K8sDistro = consts.K8sK3s

	client.Metadata.K8sVersion = "1.27.1"

	client.Metadata.Provider = consts.CloudLocal

	client.Metadata.NoMP = 2
	client.Metadata.LogVerbosity = 0
	client.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "scalar"), client); err != nil {
		panic(err)
	}

	err := controller.CreateManagedCluster(client)
	if err != nil {
		panic(err)
	}
}
