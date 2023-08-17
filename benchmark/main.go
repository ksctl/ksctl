package benchmark

import (
	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"
)

func ha() {

	cmd := &resources.CobraCmd{}

	cmd.Client.Metadata.Provider = utils.CLOUD_CIVO
	cmd.Client.Metadata.K8sDistro = utils.K8S_K3S
	cmd.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cmd.Client.Metadata.ClusterName = "benchmark"

	// managed
	// cmd.Client.Metadata.ManagedNodeType = "g4s.kube.small"

	// ha
	cmd.Client.Metadata.LoadBalancerNodeType = "g3.small"
	cmd.Client.Metadata.ControlPlaneNodeType = "g3.small"
	cmd.Client.Metadata.WorkerPlaneNodeType = "g3.small"
	cmd.Client.Metadata.DataStoreNodeType = "g3.medium"

	// cmd.Client.Metadata.CNIPlugin = "cilium"

	cmd.Client.Metadata.Region = "FRA1"
	cmd.Client.Metadata.NoCP = 3
	cmd.Client.Metadata.NoWP = 1
	cmd.Client.Metadata.NoDS = 1

	var controller controllers.Controller = control_pkg.GenKsctlController()
	// verbosity set to true
	if _, err := control_pkg.InitializeStorageFactory(&cmd.Client, true); err != nil {
		panic(err)
	}

	cmd.Client.Metadata.IsHA = true
	stat, err := controller.CreateHACluster(&cmd.Client)
	if err != nil {
		cmd.Client.Storage.Logger().Err(err.Error())
		return
	}
	cmd.Client.Storage.Logger().Success(stat)

}
