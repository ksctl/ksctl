package kubernetes_test

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/internal/kubernetes"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"os"
	"path/filepath"
	"testing"
)

var (
	ksctlK8sClient *kubernetes.K8sClusterClient
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
	ksctlK8sClient, err = kubernetes.NewInClusterClient(
		parentCtx,
		parentLogger,
		storeVars,
		true,
	)
	if err != nil {
		t.Error(err)
	}
	ksctlK8sClient, err = kubernetes.NewKubeconfigClient(
		parentCtx,
		parentLogger,
		storeVars,
		"",
		true,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestInstallApps(t *testing.T) {
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

	if err := ksctlK8sClient.CNI(
		types.KsctlApp{
			StackName: string(metadata.CiliumStandardStackID),
		},
		stateDocument,
		consts.OperationCreate,
	); err != nil {
		t.Error(err)
	}
}

func TestUnInstallApps(t *testing.T) {
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

	if err := ksctlK8sClient.CNI(
		types.KsctlApp{
			StackName: string(metadata.CiliumStandardStackID),
		},
		stateDocument,
		consts.OperationDelete,
	); err != nil {
		t.Error(err)
	}
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
