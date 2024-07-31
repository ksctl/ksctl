package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

var (
	ksctlK8sClient *K8sClusterClient
	parentCtx      context.Context
	dir            = filepath.Join(os.TempDir(), "ksctl-kubernetes-test")
	parentLogger   = logger.NewStructuredLogger(-1, os.Stdout)
	stateDocument  = &storageTypes.StorageDocument{}

	storeVars types.StorageFactory
)

func TestMain(m *testing.M) {

	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudCivo, "LON1", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func TestInitClient(t *testing.T) {
	var err error
	ksctlK8sClient, err = NewInClusterClient(
		parentCtx,
		parentLogger,
		storeVars,
		true,
		&k8sClientMock{},
		&helmClientMock{},
	)
	if err != nil {
		t.Error(err)
	}
	ksctlK8sClient, err = NewKubeconfigClient(
		parentCtx,
		parentLogger,
		storeVars,
		"",
		true,
		&k8sClientMock{},
		&helmClientMock{},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestInstallApps(t *testing.T) {
	t.Run("InstallArgoCD", func(t *testing.T) {
		if err := ksctlK8sClient.Applications(
			[]types.KsctlApp{
				{
					StackName: string(metadata.ArgocdStandardStackID),
				},
			},
			stateDocument,
			consts.OperationCreate,
		); err != nil {
			t.Error(err)
		}
	})

	t.Run("InstallCilium", func(t *testing.T) {
		if err := ksctlK8sClient.CNI(
			types.KsctlApp{
				StackName: string(metadata.CiliumStandardStackID),
			},
			stateDocument,
			consts.OperationCreate,
		); err != nil {
			t.Error(err)
		}
	})
}

func TestUnInstallApps(t *testing.T) {
	t.Run("UnInstallArgoCD", func(t *testing.T) {
		if err := ksctlK8sClient.Applications(
			[]types.KsctlApp{
				{
					StackName: string(metadata.ArgocdStandardStackID),
				},
			},
			stateDocument,
			consts.OperationDelete,
		); err != nil {
			t.Error(err)
		}
	})

	t.Run("UnInstallCilium", func(t *testing.T) {
		if err := ksctlK8sClient.CNI(
			types.KsctlApp{
				StackName: string(metadata.CiliumStandardStackID),
			},
			stateDocument,
			consts.OperationDelete,
		); err != nil {
			t.Error(err)
		}
	})
}

func TestDeleteWorkerNodes(t *testing.T) {
	if err := ksctlK8sClient.DeleteWorkerNodes(
		"node1",
	); err != nil {
		t.Error(err)
	}
}

func TestDeployRequiredControllers(t *testing.T) {
	if err := ksctlK8sClient.DeployRequiredControllers(
		stateDocument,
	); err != nil {
		t.Error(err)
	}
}
