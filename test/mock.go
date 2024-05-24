package test

import (
	"context"
	"fmt"
	"os"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers"
)

var (
	cli        *types.KsctlClient
	controller controllers.Controller
	dir        = fmt.Sprintf("%s ksctl-black-box-test", os.TempDir())
)

func InitCore() (err error) {
	ctx := context.WithValue(
		context.Background(),
		"USERID",
		"demo",
	)
	ctx = context.WithValue(
		ctx,
		consts.KsctlCustomDirLoc,
		dir,
	)
	ctx = context.WithValue(
		ctx,
		consts.KsctlTestFlagKey,
		"true",
	)

	cli = new(types.KsctlClient)

	cli.Metadata.ClusterName = "fake"
	cli.Metadata.StateLocation = consts.StoreLocal
	cli.Metadata.K8sDistro = consts.K8sK3s

	log := logger.NewGeneralLogger(-1, os.Stdout)

	controller, err = control_pkg.GenKsctlController(
		ctx,
		log,
		cli,
	)
	return
}

func ExecuteManagedRun() error {

	if err := controller.CreateManagedCluster(); err != nil {
		return err
	}

	if _, err := controller.SwitchCluster(); err != nil {
		return err
	}

	if err := controller.GetCluster(); err != nil {
		return err
	}

	if err := controller.DeleteManagedCluster(); err != nil {
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

	return ExecuteManagedRun()
}

func CivoTestingManaged() error {
	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.ManagedNodeType = "g4s.kube.small"
	cli.Metadata.NoMP = 2

	return ExecuteManagedRun()
}

func LocalTestingManaged() error {
	cli.Metadata.Provider = consts.CloudLocal
	cli.Metadata.NoMP = 5

	return ExecuteManagedRun()
}

func ExecuteHARun() error {

	if err := controller.CreateHACluster(); err != nil {
		return err
	}

	if _, err := controller.SwitchCluster(); err != nil {
		return err
	}

	if err := controller.GetCluster(); err != nil {
		return err
	}

	cli.Metadata.NoWP = 0
	if err := controller.DelWorkerPlaneNode(); err != nil {
		return err
	}

	if err := controller.GetCluster(); err != nil {
		return err
	}

	cli.Metadata.NoWP = 1
	if err := controller.AddWorkerPlaneNode(); err != nil {
		return err
	}

	if err := controller.GetCluster(); err != nil {
		return err
	}

	if err := controller.DeleteHACluster(); err != nil {
		return err
	}
	return nil
}

func CivoTestingHAKubeadm() error {
	cli.Metadata.LoadBalancerNodeType = "fake.small"
	cli.Metadata.ControlPlaneNodeType = "fake.small"
	cli.Metadata.WorkerPlaneNodeType = "fake.small"
	cli.Metadata.DataStoreNodeType = "fake.small"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.K8sDistro = consts.K8sKubeadm
	cli.Metadata.NoCP = 5
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.28"

	return ExecuteHARun()
}

func CivoTestingHAK3s() error {
	cli.Metadata.LoadBalancerNodeType = "fake.small"
	cli.Metadata.ControlPlaneNodeType = "fake.small"
	cli.Metadata.WorkerPlaneNodeType = "fake.small"
	cli.Metadata.DataStoreNodeType = "fake.small"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "LON1"
	cli.Metadata.Provider = consts.CloudCivo
	cli.Metadata.K8sDistro = consts.K8sK3s
	cli.Metadata.NoCP = 5
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.27.4"

	return ExecuteHARun()
}

func AzureTestingHAKubeadm() error {
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.K8sDistro = consts.K8sKubeadm
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.28"

	return ExecuteHARun()
}

func AzureTestingHAK3s() error {
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAzure
	cli.Metadata.K8sDistro = consts.K8sK3s
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.27.4"

	return ExecuteHARun()
}

func AwsTestingHA() error {
	cli.Metadata.LoadBalancerNodeType = "fake"
	cli.Metadata.ControlPlaneNodeType = "fake"
	cli.Metadata.WorkerPlaneNodeType = "fake"
	cli.Metadata.DataStoreNodeType = "fake"

	cli.Metadata.IsHA = true

	cli.Metadata.Region = "fake"
	cli.Metadata.Provider = consts.CloudAws
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.27.4"

	return ExecuteHARun()
}
