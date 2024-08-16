//go:build testing_k8s_manifest

package kubernetes

import (
	"context"
	"fmt"
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
	ksctlK8sClient *K8sClusterClient
	parentCtx      context.Context
	dir            = filepath.Join(os.TempDir(), "ksctl-k9s-manifest-test")
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
	kubeconfigLoc := os.Getenv("KUBECONFIG")
	if len(kubeconfigLoc) == 0 {
		t.Error("KUBECONFIG env variable not set")
	}
	v, err := os.ReadFile(kubeconfigLoc)
	if err != nil {
		t.Error(err)
	}
	ksctlK8sClient, err = NewKubeconfigClient(
		parentCtx,
		parentLogger,
		storeVars,
		string(v),
		false, nil, nil,
	)
	if err != nil {
		t.Error(err)
	}
}

var componentUrl = []string{
	"spin-operator.crds.yaml",
	"spin-operator.runtime-class.yaml",
	"spin-operator.shim-executor.yaml",
}

func TestTryOutApply(t *testing.T) {
	baseUri := "https://github.com/spinkube/spin-operator/releases/download/v0.2.0/"
	for _, component := range componentUrl {
		if err := installKubectl(ksctlK8sClient, &metadata.KubectlHandler{
			Url:             baseUri + component,
			Version:         "v0.2.0",
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("KubeSpin (ver:) is an open source project that streamlines developing, deploying and operating WebAssembly workloads in Kubernetes - resulting in delivering smaller, more portable applications and incredible compute performance benefits"),
			PostInstall:     "https://www.spinkube.dev/docs/topics/",
		}); err != nil {
			t.Error(err)
		}
	}
}

func TestTryOutDelete(t *testing.T) {
	baseUri := "https://github.com/spinkube/spin-operator/releases/download/v0.2.0/"

	for idx := len(componentUrl) - 1; idx >= 0; idx-- {
		component := componentUrl[idx]
		if err := deleteKubectl(ksctlK8sClient, &metadata.KubectlHandler{
			Url:             baseUri + component,
			Version:         "v0.2.0",
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("KubeSpin (ver: ) is an open source project that streamlines developing, deploying and operating WebAssembly workloads in Kubernetes - resulting in delivering smaller, more portable applications and incredible compute performance benefits"),
			PostInstall:     "https://www.spinkube.dev/docs/topics/",
		}); err != nil {
			t.Error(err)
		}
	}
}
