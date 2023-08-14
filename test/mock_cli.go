package test

import (
	"fmt"
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"
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

func AzureInit(cmd *resources.CobraCmd) {
	cmd.Client.Metadata.Provider = utils.CLOUD_AZURE
	cmd.Client.Metadata.K8sDistro = utils.K8S_K3S
	cmd.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cmd.Client.Metadata.ClusterName = "benchmark"

	// managed
	cmd.Client.Metadata.ManagedNodeType = "Standard_DS2_v2"

	// ha
	cmd.Client.Metadata.LoadBalancerNodeType = "Standard_F2s"
	cmd.Client.Metadata.ControlPlaneNodeType = "Standard_F2s"

	cmd.Client.Metadata.WorkerPlaneNodeType = "Standard_F2s"

	cmd.Client.Metadata.DataStoreNodeType = "Standard_F2s"

	cmd.Client.Metadata.CNIPlugin = "cilium"

	cmd.Client.Metadata.Region = "eastus"
}

func CivoInit(cmd *resources.CobraCmd) {
	cmd.Client.Metadata.Provider = utils.CLOUD_CIVO
	cmd.Client.Metadata.K8sDistro = utils.K8S_K3S
	cmd.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cmd.Client.Metadata.ClusterName = "benchmark"

	// managed
	cmd.Client.Metadata.ManagedNodeType = "g4s.kube.small"

	// ha
	cmd.Client.Metadata.LoadBalancerNodeType = "g3.small"
	cmd.Client.Metadata.ControlPlaneNodeType = "g3.small"

	cmd.Client.Metadata.WorkerPlaneNodeType = "g3.small"

	cmd.Client.Metadata.DataStoreNodeType = "g3.medium"

	cmd.Client.Metadata.CNIPlugin = "cilium"

	cmd.Client.Metadata.Region = "FRA1"
}

func LocalInit(cmd *resources.CobraCmd) {
	// cmd.Client.Metadata.Provider = utils.CLOUD_LOCAL
	// cmd.Client.Metadata.K8sDistro = utils.K8S_K3S
	cmd.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cmd.Client.Metadata.ClusterName = "benchmark"

	// managed
	// cmd.Client.Metadata.ManagedNodeType = "g4s.kube.small"
	//
	// // ha
	// cmd.Client.Metadata.LoadBalancerNodeType = "g3.small"
	// cmd.Client.Metadata.ControlPlaneNodeType = "g3.small"
	//
	// cmd.Client.Metadata.WorkerPlaneNodeType = "g3.small"
	//
	// cmd.Client.Metadata.DataStoreNodeType = "g3.medium"
	//
	// cmd.Client.Metadata.CNIPlugin = "cilium"
	//
	// cmd.Client.Metadata.Region = "FRA1"
}

func main() {
	cmd := &resources.CobraCmd{}
	NewCli(cmd)

	// // AZURE
	// // CIVO
	// cmd.Client.Metadata.Provider = utils.CLOUD_AZURE
	// cmd.Client.Metadata.K8sDistro = utils.K8S_K3S
	// cmd.Client.Metadata.StateLocation = utils.STORE_LOCAL
	// cmd.Client.Metadata.ClusterName = "benchmark"
	// // managed
	// //cmd.Client.Metadata.ManagedNodeType = "g4s.kube.small"
	// cmd.Client.Metadata.ManagedNodeType = "Standard_DS2_v2"
	// // ha
	// //cmd.Client.Metadata.LoadBalancerNodeType = "g3.small"
	// cmd.Client.Metadata.LoadBalancerNodeType = "Standard_F2s"
	// //cmd.Client.Metadata.ControlPlaneNodeType = "g3.small"
	// cmd.Client.Metadata.ControlPlaneNodeType = "Standard_F2s"
	// //cmd.Client.Metadata.WorkerPlaneNodeType = "g3.small"
	// cmd.Client.Metadata.WorkerPlaneNodeType = "Standard_F2s"
	// //cmd.Client.Metadata.DataStoreNodeType = "g3.medium"
	// cmd.Client.Metadata.DataStoreNodeType = "Standard_F2s"
	// cmd.Client.Metadata.CNIPlugin = "cilium"
	// //cmd.Client.Metadata.Region = "FRA1"
	// cmd.Client.Metadata.Region = "eastus"

	// CivoInit(cmd)
	// AzureInit(cmd)
	LocalInit(cmd)

	cmd.Client.Metadata.NoCP = 3
	cmd.Client.Metadata.NoWP = 1
	cmd.Client.Metadata.NoDS = 1

	var controller controllers.Controller = control_pkg.GenKsctlController()
	// NOTE: verbosity set to true
	if _, err := control_pkg.InitializeStorageFactory(&cmd.Client, true); err != nil {
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

		cmd.Client.Metadata.K8sVersion = "1.27.1"
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
		cmd.Client.Metadata.NoMP = 2
		cmd.Client.Metadata.K8sVersion = "1.27.1" // TODO: for civo managed and local
		// cmd.Client.Metadata.K8sVersion = "1.27" // TODO: for Azure managed
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

		cmd.Client.Metadata.NoWP = 1
		stat, err := controller.AddWorkerPlaneNode(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	case 7:
		cmd.Client.Metadata.IsHA = true

		cmd.Client.Metadata.NoWP = 0
		stat, err := controller.DelWorkerPlaneNode(&cmd.Client)
		if err != nil {
			cmd.Client.Storage.Logger().Err(err.Error())
			return
		}
		cmd.Client.Storage.Logger().Success(stat)
	}
}
