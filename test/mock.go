package test

import (
	"context"
	"fmt"
	"os"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers"
)

var (
	cli        *resources.KsctlClient
	controller controllers.Controller
	dir        = fmt.Sprintf("%s ksctl-black-box-test", os.TempDir())
)

func StartCloud() {
	cli = new(resources.KsctlClient)
	controller = control_pkg.GenKsctlController()

	cli.Metadata.ClusterName = "fake"
	cli.Metadata.StateLocation = consts.StoreLocal
	cli.Metadata.K8sDistro = consts.K8sK3s
	cli.Metadata.LogVerbosity = -1
	cli.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "demo"), cli); err != nil {
		panic(err)
	}
}

func ExecuteManagedRun() error {

	var err error

	err = controller.CreateManagedCluster(cli)
	if err != nil {
		return err
	}
	_, err = controller.SwitchCluster(cli)
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

	return ExecuteManagedRun()
}

func CivoTestingManaged() error {
	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Metadata.NoMP = 2

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)

	return ExecuteManagedRun()
}

func LocalTestingManaged() error {
	cli.Metadata.Provider = consts.CloudLocal
	cli.Metadata.NoMP = 5

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)

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

	err = controller.CreateHACluster(cli)
	if err != nil {
		return err
	}
	_, err = controller.SwitchCluster(cli)
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

	err = controller.CreateHACluster(cli)
	if err != nil {
		return err
	}

	_, err = controller.SwitchCluster(cli)
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

	err = controller.CreateHACluster(cli)
	if err != nil {
		return err
	}

	_, err = controller.SwitchCluster(cli)
	if err != nil {
		return err
	}

	err = controller.DeleteHACluster(cli)
	if err != nil {
		return err
	}
	return nil
}
