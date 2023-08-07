package main

import (
	"fmt"
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
)

func NewCli(cmd *resources.CobraCmd) {
	cmd.Version = os.Getenv("KSCTL_VERSION")

	if len(cmd.Version) == 0 {
		cmd.Version = "dummy v11001.2"
	}
}

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cmd := &resources.CobraCmd{}
	NewCli(cmd)

	cmd.Client.Metadata.Provider = "civo"
	cmd.Client.Metadata.K8sDistro = "k3s"
	cmd.Client.Metadata.StateLocation = "local"
	cmd.Client.Metadata.ClusterName = "dummy-name"

	// managed
	cmd.Client.Metadata.ManagedNodeType = "g4s.kube.small"

	// ha
	cmd.Client.Metadata.LoadBalancerNodeType = "g3.small"
	cmd.Client.Metadata.ControlPlaneNodeType = "g3.small"
	cmd.Client.Metadata.WorkerPlaneNodeType = "g3.small"
	cmd.Client.Metadata.DataStoreNodeType = "g3.large"

	cmd.Client.Metadata.CNIPlugin = "cilium"

	cmd.Client.Metadata.Region = "LON1"
	cmd.Client.Metadata.NoCP = 3
	cmd.Client.Metadata.NoWP = 1
	cmd.Client.Metadata.NoDS = 1

	var controller controllers.Controller = control_pkg.GenKsctlController()
	if _, err := control_pkg.InitializeStorageFactory(&cmd.Client); err != nil {
		panic(err)
	}
	choice := -1
	fmt.Println(`
[0] enter credential
[1] create HA
[2] Delete HA
[3] Create Managed
[4] Delete Managed
[5] Get Clusters
[6] add workerplane node(s)
[7] delete workerplane node(s)

Your Choice`)
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return
	}
	switch choice {
	case 0:
		stat, err := controller.Credentials(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 1:
		cmd.Client.Metadata.IsHA = true
		stat, err := controller.CreateHACluster(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 2:
		cmd.Client.Metadata.IsHA = true

		stat, err := controller.DeleteHACluster(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 3:
		cmd.Client.Metadata.NoMP = 1
		cmd.Client.Metadata.K8sVersion = "1.27.1"
		stat, err := controller.CreateManagedCluster(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 4:
		stat, err := controller.DeleteManagedCluster(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)

	case 5:
		stat, err := controller.GetCluster(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 6:
		cmd.Client.Metadata.IsHA = true

		stat, err := controller.AddWorkerPlaneNode(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 7:
		cmd.Client.Metadata.IsHA = true

		stat, err := controller.DelWorkerPlaneNode(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	}
}
