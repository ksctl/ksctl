// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
