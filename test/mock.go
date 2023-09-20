package test

import (
	"fmt"
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
	"github.com/kubesimplify/ksctl/api/utils"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
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
	cli.Client.Metadata.StateLocation = STORE_LOCAL
	cli.Client.Metadata.K8sDistro = K8S_K3S

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
	cli.Client.Metadata.Provider = CLOUD_AZURE
	cli.Client.Metadata.ManagedNodeType = "fake"
	cli.Client.Metadata.NoMP = 2
	cli.Client.Metadata.K8sVersion = "1.27"

	_ = os.Setenv(string(KSCTL_CUSTOM_DIR_ENABLED), dir)
	azManaged := utils.GetPath(CLUSTER_PATH, CLOUD_AZURE, CLUSTER_TYPE_MANG)

	if err := os.MkdirAll(azManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	return ExecuteManagedRun()
}

func CivoTestingManaged() error {
	cli.Client.Metadata.Region = "LON1"
	cli.Client.Metadata.Provider = CLOUD_CIVO
	cli.Client.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Client.Metadata.NoMP = 2

	_ = os.Setenv(string(KSCTL_CUSTOM_DIR_ENABLED), dir)
	azManaged := utils.GetPath(CLUSTER_PATH, CLOUD_CIVO, CLUSTER_TYPE_MANG)

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
	cli.Client.Metadata.Provider = CLOUD_CIVO
	cli.Client.Metadata.NoCP = 5
	cli.Client.Metadata.NoWP = 1
	cli.Client.Metadata.NoDS = 3
	cli.Client.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(string(KSCTL_CUSTOM_DIR_ENABLED), dir)
	azHA := utils.GetPath(CLUSTER_PATH, CLOUD_CIVO, CLUSTER_TYPE_HA)

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
	cli.Client.Metadata.Provider = CLOUD_AZURE
	cli.Client.Metadata.NoCP = 3
	cli.Client.Metadata.NoWP = 1
	cli.Client.Metadata.NoDS = 1
	cli.Client.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(string(KSCTL_CUSTOM_DIR_ENABLED), dir)
	azHA := utils.GetPath(CLUSTER_PATH, CLOUD_AZURE, CLUSTER_TYPE_HA)

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
