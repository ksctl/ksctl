package test

import (
	"fmt"
	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
)

var (
	cli        *resources.CobraCmd
	controller controllers.Controller
	dir        = fmt.Sprintf("%s/ksctl-black-box-test", os.TempDir())
)

func StartCloud() {
	cli = &resources.CobraCmd{}
	controller = control_pkg.GenKsctlController()

	cli.Client.Metadata.ClusterName = "fake"
	cli.Client.Metadata.StateLocation = utils.STORE_LOCAL
	cli.Client.Metadata.K8sDistro = utils.K8S_K3S

	if _, err := control_pkg.InitializeStorageFactory(&cli.Client, false); err != nil {
		panic(err)
	}
}

func ExecuteManagedRun() error {

	var err error

	_, err = controller.CreateManagedCluster(&cli.Client)
	if err != nil {
		return err
	}

	_, err = controller.DeleteManagedCluster(&cli.Client)
	if err != nil {
		return err
	}
	return nil
}

func AzureTestingManaged() error {
	cli.Client.Metadata.Region = "fake"
	cli.Client.Metadata.Provider = utils.CLOUD_AZURE
	cli.Client.Metadata.ManagedNodeType = "fake"
	cli.Client.Metadata.NoMP = 2
	cli.Client.Metadata.K8sVersion = "1.27"

	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	azManaged := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_MANG)

	if err := os.MkdirAll(azManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	return ExecuteManagedRun()
}

func CivoTestingManaged() error {
	cli.Client.Metadata.Region = "LON1"
	cli.Client.Metadata.Provider = utils.CLOUD_CIVO
	cli.Client.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Client.Metadata.NoMP = 2

	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	azManaged := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_MANG)

	if err := os.MkdirAll(azManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	return ExecuteManagedRun()
}

func CivoTestingHA() error {
	var err error
	cli.Client.Metadata.LoadBalancerNodeType = "fake.small"
	cli.Client.Metadata.ControlPlaneNodeType = "fake.small"
	cli.Client.Metadata.WorkerPlaneNodeType = "fake.small"
	cli.Client.Metadata.DataStoreNodeType = "fake.small"

	cli.Client.Metadata.IsHA = true

	cli.Client.Metadata.Region = "LON1"
	cli.Client.Metadata.Provider = utils.CLOUD_CIVO
	cli.Client.Metadata.NoCP = 5
	cli.Client.Metadata.NoWP = 1
	cli.Client.Metadata.NoDS = 3
	cli.Client.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	azHA := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA)

	if err := os.MkdirAll(azHA, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Created tmp directories")

	_, err = controller.CreateHACluster(&cli.Client)
	if err != nil {
		return err
	}

	_, err = controller.DeleteHACluster(&cli.Client)
	return nil
}

func AzureTestingHA() error {
	var err error
	cli.Client.Metadata.LoadBalancerNodeType = "fake"
	cli.Client.Metadata.ControlPlaneNodeType = "fake"
	cli.Client.Metadata.WorkerPlaneNodeType = "fake"
	cli.Client.Metadata.DataStoreNodeType = "fake"

	cli.Client.Metadata.IsHA = true

	cli.Client.Metadata.Region = "fake"
	cli.Client.Metadata.Provider = utils.CLOUD_AZURE
	cli.Client.Metadata.NoCP = 3
	cli.Client.Metadata.NoWP = 1
	cli.Client.Metadata.NoDS = 1
	cli.Client.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	azHA := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_HA)

	if err := os.MkdirAll(azHA, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Created tmp directories")

	_, err = controller.CreateHACluster(&cli.Client)
	if err != nil {
		return err
	}

	_, err = controller.DeleteHACluster(&cli.Client)
	return nil
}
