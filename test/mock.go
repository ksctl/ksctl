package test

import (
	"fmt"
	"os"

	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers"
)

var (
	cli        *resources.KsctlClient
	controller controllers.Controller
	dir        = fmt.Sprintf("%s/ksctl-black-box-test", os.TempDir())
)

func StartCloud() {
	cli = new(resources.KsctlClient)
	controller = control_pkg.GenKsctlController()

	cli.Metadata.ClusterName = "fake"
	cli.Metadata.StateLocation = consts.StoreLocal
	cli.Metadata.K8sDistro = consts.K8sK3s
	cli.Metadata.LogVerbosity = -1
	cli.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(cli); err != nil {
		panic(err)
	}
}

func ExecuteManagedRun() error {

	var err error

	err = controller.CreateManagedCluster(cli)
	if err != nil {
		return err
	}

	err = controller.DeleteManagedCluster(cli)
	if err != nil {
		return err
	}
	return nil
}

func AzureTestingManaged() error {
	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.ManagedNodeType = "fake"
	cli.Metadata.NoMP = 2
	cli.Metadata.K8sVersion = "1.27"

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	azManaged := helpers.GetPath(consts.UtilClusterPath, consts.CloudAzure, consts.ClusterTypeMang)

	if err := os.MkdirAll(azManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	return ExecuteManagedRun()
}

func CivoTestingManaged() error {
	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Metadata.NoMP = 2

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	azManaged := helpers.GetPath(consts.UtilClusterPath, consts.CloudCivo, consts.ClusterTypeMang)

	if err := os.MkdirAll(azManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	return ExecuteManagedRun()
}

func LocalTestingManaged() error {
	cli.Metadata.Provider = consts.CloudLocal
	cli.Metadata.NoMP = 5

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	localManaged := helpers.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang)

	if err := os.MkdirAll(localManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	return ExecuteManagedRun()
}

func CivoTestingHA() error {
	var err error
	cli.Metadata.LoadBalancerNodeType = "fake.small"
	cli.Metadata.ControlPlaneNodeType = "fake.small"
	cli.Metadata.WorkerPlaneNodeType = "fake.small"
	cli.Metadata.DataStoreNodeType = "fake.small"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.NoCP = 5
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	azHA := helpers.GetPath(consts.UtilClusterPath, consts.CloudCivo, consts.ClusterTypeHa)

	if err := os.MkdirAll(azHA, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Created tmp directories")

	err = controller.CreateHACluster(cli)
	if err != nil {
		return err
	}

	err = controller.DeleteHACluster(cli)
	if err != nil {
		return err
	}
	return nil
}

func AzureTestingHA() error {
	var err error
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 1
	cli.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	azHA := helpers.GetPath(consts.UtilClusterPath, consts.CloudAzure, consts.ClusterTypeHa)

	if err := os.MkdirAll(azHA, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Created tmp directories")

	err = controller.CreateHACluster(cli)
	if err != nil {
		return err
	}

	err = controller.DeleteHACluster(cli)
	if err != nil {
		return err
	}
	return nil
}

func AwsTestingHA() error {
	var err error
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAws
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 1
	cli.Metadata.K8sVersion = "1.27.4"

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	awsHA := helpers.GetPath(consts.UtilClusterPath, consts.CloudAws, consts.ClusterTypeHa)

	if err := os.MkdirAll(awsHA, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Created tmp directories")

	err = controller.CreateHACluster(cli)
	if err != nil {
		return err
	}

	err = controller.DeleteHACluster(cli)
	if err != nil {
		return err
	}
	return nil
}
