//go:build testing_helm_oci

package helm_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/logger"
)

var (
	helmClient   *helm.Client
	parentCtx    context.Context
	dir          = filepath.Join(os.TempDir(), "ksctl-kubernetes-test")
	parentLogger = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestMain(m *testing.M) {

	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestInitClient(t *testing.T) {

	v, err := os.ReadFile("/Users/dipankardas/.kube/config")
	if err != nil {
		t.Fatal(err)
	}

	helmClient, err = helm.NewKubeconfigHelmClient(
		parentCtx,
		parentLogger,
		string(v),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestHelmOciInstall(t *testing.T) {

	releaseName := "helm-sdk-example"
	chartRef := "oci://ghcr.io/stefanprodan/charts/podinfo"
	releaseValues := map[string]interface{}{
		"replicaCount": "2",
	}
	chartVer := "6.4.1"
	releaseNamespace := "xyz"

	if err := helmClient.InstallChart(
		chartRef,
		chartVer,
		chartRef,
		releaseNamespace,
		releaseName,
		true,
		releaseValues,
	); err != nil {
		t.Fatal(err)
	}

	if err := helmClient.ListInstalledCharts(); err != nil {
		t.Fatal(err)
	}

	if err := helmClient.UninstallChart(releaseNamespace, releaseName); err != nil {
		t.Fatal(err)
	}
}
