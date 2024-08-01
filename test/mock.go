package test

import (
	"context"
	"os"
	"path/filepath"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	cli *types.KsctlClient
	dir = filepath.Join(os.TempDir(), "ksctl-black-box-test")
	ctx context.Context
)

func InitCore() (err error) {
	ctx = context.WithValue(
		context.Background(),
		consts.KsctlContextUserID,
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

	return
}

func ExecuteKsctlSpecificRun() error {

	log := logger.NewStructuredLogger(-1, os.Stdout)

	controller, err := control_pkg.NewManagerClusterKsctl(
		ctx,
		log,
		cli,
	)
	if err != nil {
		return err
	}
	if _, err := controller.SwitchCluster(); err != nil {
		return err
	}

	if _, err := controller.InfoCluster(); err != nil {
		return err
	}

	if err := controller.GetCluster(); err != nil {
		return err
	}
	return nil
}

func ExecuteManagedRun() error {
	log := logger.NewStructuredLogger(-1, os.Stdout)

	controller, err := control_pkg.NewManagerClusterManaged(
		ctx,
		log,
		cli,
	)
	if err != nil {
		return err
	}

	if err := controller.CreateCluster(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	if err := controller.DeleteCluster(); err != nil {
		return err
	}
	return nil
}

func ExecuteHARun() error {
	log := logger.NewStructuredLogger(-1, os.Stdout)

	controller, err := control_pkg.NewManagerClusterSelfManaged(
		ctx,
		log,
		cli,
	)
	if err != nil {
		return err
	}

	if err := controller.CreateCluster(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	cli.Metadata.NoWP = 0
	if err := controller.DelWorkerPlaneNodes(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	cli.Metadata.NoWP = 1
	if err := controller.AddWorkerPlaneNodes(); err != nil {
		return err
	}

	if err := ExecuteKsctlSpecificRun(); err != nil {
		return err
	}

	if err := controller.DeleteCluster(); err != nil {
		return err
	}
	return nil
}

func AwsTestingManaged() error {
	cli.Metadata.Region = "fake-region"
	cli.Metadata.Provider = consts.CloudAws
	cli.Metadata.ManagedNodeType = "fake"
	cli.Metadata.NoMP = 2
	cli.Metadata.K8sVersion = "1.30"

	return ExecuteManagedRun()
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
	cli.Metadata.K8sVersion = "1.30.0"

	return ExecuteManagedRun()
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

	cli.Metadata.Region = "fake-region"
	cli.Metadata.Provider = consts.CloudAws
	cli.Metadata.NoCP = 3
	cli.Metadata.NoWP = 1
	cli.Metadata.NoDS = 3
	cli.Metadata.K8sVersion = "1.27.4"

	return ExecuteHARun()
}
