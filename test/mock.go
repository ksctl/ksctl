package test

import (
	"fmt"
	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"
)

var (
	cli        *resources.CobraCmd
	controller controllers.Controller
)

func StartCloud() {
	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = "fake"
	cli.Client.Metadata.Region = "LON1"
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.Provider = utils.CLOUD_CIVO
	cli.Client.Metadata.K8sDistro = utils.K8S_K3S

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, false); err != nil {
		panic(err)
	}
}

func ProvideMockCivoClient() civo.CivoGo {
	return &civo.CivoGoMockClient{}
}

func CivoTestingManaged() error {
	var err error
	cli.Client.Cloud, err = civo.ReturnCivoStruct(cli.Client.Metadata, ProvideMockCivoClient)
	cli.Client.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Client.Metadata.NoMP = 2

	err = cli.Client.Cloud.InitState(cli.Client.Storage, utils.OPERATION_STATE_CREATE)
	if err != nil {
		return err
	}

	_, err = controller.CreateManagedCluster(&cli.Client)
	if err != nil {
		return err
	}

	_, err = controller.DeleteManagedCluster(&cli.Client)
	return nil
}

func CivoTestingHA() error {
	var err error
	cli.Client.Cloud, err = civo.ReturnCivoStruct(cli.Client.Metadata, ProvideMockCivoClient)
	cli.Client.Metadata.LoadBalancerNodeType = "fake.small"
	cli.Client.Metadata.ControlPlaneNodeType = "fake.small"
	cli.Client.Metadata.WorkerPlaneNodeType = "fake.small"
	cli.Client.Metadata.DataStoreNodeType = "fake.small"

	cli.Client.Metadata.IsHA = true
	// cmd.Client.Metadata.CNIPlugin = "cilium"

	cli.Client.Metadata.Region = "FRA1"
	cli.Client.Metadata.NoCP = 3
	cli.Client.Metadata.NoWP = 1
	cli.Client.Metadata.NoDS = 1
	cli.Client.Metadata.K8sVersion = "1.27.4"

	// it will faile in kubernetes part as it expectes real IP addresses

	err = cli.Client.Cloud.InitState(cli.Client.Storage, utils.OPERATION_STATE_CREATE)
	if err != nil {
		return err
	}

	_, err = controller.CreateHACluster(&cli.Client)
	if err != nil {
		return err
	}

	_, err = controller.DeleteHACluster(&cli.Client)
	return nil
}

func Cleanup() error {
	path := fmt.Sprintf("%s/.ksctl", utils.GetUserName())
	return cli.Client.Storage.Path(path).DeleteDir()
}
