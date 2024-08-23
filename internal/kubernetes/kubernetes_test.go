//go:build !testing_k8s_manifest

package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ksctl/ksctl/commons"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"github.com/ksctl/ksctl/poller"
)

var (
	ksctlK8sClient *K8sClusterClient
	parentCtx      context.Context
	dir            = filepath.Join(os.TempDir(), "ksctl-kubernetes-test")
	parentLogger   = logger.NewStructuredLogger(-1, os.Stdout)
	stateDocument  = &storageTypes.StorageDocument{}

	storeVars types.StorageFactory
)

func initPoller() {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		if org == "spinkube" {
			if repo == "spin-operator" {
				vers = append(vers, "v0.2.0")
			} else if repo == "containerd-shim-spin" {
				vers = append(vers, "v0.15.1")
			}
		}
		if org == "cert-manager" && repo == "cert-manager" {
			vers = append(vers, "v1.15.3")
		}
		if org == "cilium" && repo == "cilium" {
			vers = append(vers, "v1.16.1")
		}
		if org == "flannel-io" && repo == "flannel" {
			vers = append(vers, "v0.25.5")
		}

		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		return vers, nil
	})
}

func TestMain(m *testing.M) {

	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudCivo, "LON1", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	initPoller()

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

	tc := []struct {
		ver    string
		suffix string
	}{
		{"main", ""}, // by default
		{"v0.1.1", ""},
		{"f14cd9094b2160c40ef8734e90141df81c22999e", "/pr-134"},
	}

	for _, tt := range tc {
		t.Run(fmt.Sprintf("Deploy Required Controllers: ver %s and prefix %s", tt.ver, tt.suffix), func(t *testing.T) {
			commons.OCIVersion = tt.ver
			commons.OCIImgSuffix = tt.suffix

			var _app types.KsctlApp
			var ctx context.Context

			loc := filepath.Join(os.TempDir(), "deploy.yml")

			if tt.suffix != "" {
				_app = types.KsctlApp{
					StackName: string(metadata.KsctlOperatorsID),
					Overrides: map[string]map[string]any{
						string(metadata.KsctlApplicationComponentID): {
							"url":     "file:::" + loc,
							"version": tt.ver,
						},
					},
				}
				ctx = context.WithValue(parentCtx, consts.KsctlComponentOverrides, "application=file:::"+loc)

				data := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo
data:
  key: value
`)
				if err := os.WriteFile(loc, data, 0644); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.Remove(loc); err != nil {
						t.Fatal(err)
					}
				})
			} else {

				_app = types.KsctlApp{
					StackName: string(metadata.KsctlOperatorsID),
					Overrides: map[string]map[string]any{
						string(metadata.KsctlApplicationComponentID): {
							"version": tt.ver,
						},
					},
				}
				ctx = parentCtx
			}

			_client, err := NewInClusterClient(
				ctx,
				parentLogger,
				storeVars,
				true,
				&k8sClientMock{},
				&helmClientMock{},
			)
			if err != nil {
				t.Error(err)
			}

			if err := _client.DeployRequiredControllers(
				stateDocument,
			); err != nil {
				t.Error(err)
			}

			// need to uninstall the ksctl operator
			if err := _client.Applications(
				[]types.KsctlApp{
					_app,
				},
				stateDocument,
				consts.OperationDelete,
			); err != nil {
				t.Error(err)
			}
		})
	}
}
