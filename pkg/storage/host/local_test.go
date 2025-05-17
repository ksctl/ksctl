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

package host

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	"gotest.tools/v3/assert"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
)

var (
	db *Store

	parentCtx    context.Context
	parentLogger logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
	dir                        = filepath.Join(os.TempDir(), "ksctl-local-store-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
	parentCtx = context.WithValue(parentCtx, consts.KsctlCustomDirLoc, dir)

	_ = m.Run()

	_ = os.RemoveAll(dir)
}

func TestInitStorage(t *testing.T) {
	db = NewClient(parentCtx, parentLogger)
	err := db.Setup(consts.CloudAzure, "region", "name", consts.ClusterTypeSelfMang)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Connect(context.TODO()); err != nil {
		t.Fatal(err)
	}
}

func TestReader(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	v := &statefile.StorageDocument{}

	d, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "fake.json"), d, 0755); err != nil {
		t.Fatal(err)
	}
}

func TestGenOsClusterPath(t *testing.T) {
	expectedHa := filepath.Join([]string{
		os.TempDir(),
		"ksctl-local-store-test",
		".ksctl",
		"state",
		"aws",
		"selfmanaged",
		"name region"}...)
	expectedManaged := filepath.Join([]string{
		os.TempDir(),
		"ksctl-local-store-test",
		".ksctl",
		"state",
		"azure",
		"managed",
		"name region"}...)

	gotH, err := db.genOsClusterPath(string(consts.CloudAws), string(consts.ClusterTypeSelfMang), "name region")
	assert.NilError(t, err)
	if gotH != expectedHa {
		t.Fatalf("expected %s; but got %s\n", expectedHa, gotH)
	}

	gotM, err := db.genOsClusterPath(string(consts.CloudAzure), string(consts.ClusterTypeMang), "name region")
	assert.NilError(t, err)
	if gotM != expectedManaged {
		t.Fatalf("expected %s; but got %s\n", expectedManaged, gotM)
	}
}

func TestStore_RWD(t *testing.T) {
	if _, err := db.Read(); err == nil {
		t.Fatal("Error should occur as there is no folder created")
	}

	if err := db.DeleteCluster(); err == nil {
		t.Fatal("Error should happen on deleting cluster info")
	}

	if err := db.AlreadyCreated(consts.CloudAzure, "region", "name", consts.ClusterTypeSelfMang); err == nil {
		t.Fatalf("Error should happen on checking for presence of the cluster")
	}

	fakeData := &statefile.StorageDocument{
		Region:      "region",
		ClusterName: "name",
		ClusterType: "selfmanaged",
	}
	err := db.Write(fakeData)
	assert.NilError(t, err)

	err = db.AlreadyCreated(consts.CloudAzure, "region", "name", consts.ClusterTypeSelfMang)
	assert.NilError(t, err, fmt.Sprintf("Error shouldn't happen on checking for presence of the cluster: %v", err))

	if gotFakeData, err := db.Read(); err != nil {
		t.Fatalf("Error shouln't happen on reading file: %v", err)
	} else {
		if _, err := db.Read(); err != nil {
			t.Fatalf("Second Read failed")
		}
		fmt.Printf("%#+v\n", gotFakeData)

		if !reflect.DeepEqual(gotFakeData, fakeData) {
			t.Fatalf("Written data doesn't match Reading")
		}
	}

	err = db.DeleteCluster()
	assert.NilError(t, err, fmt.Sprintf("Error shouln't happen on deleting cluster info: %v", err))
}

func TestGetClusterInfo(t *testing.T) {

	t.Run("Setup some demo clusters", func(t *testing.T) {

		func() {

			if err := db.Setup(consts.CloudAzure, "regionAzure", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}

			fakeData := &statefile.StorageDocument{
				Region:        "regionAzure",
				ClusterName:   "name_managed",
				ClusterType:   "managed",
				InfraProvider: consts.CloudAzure,
				CloudInfra:    &statefile.InfrastructureState{Azure: &statefile.StateConfigurationAzure{}},
			}

			err := db.Write(fakeData)
			if err != nil {
				t.Fatalf("Error shouln't happen: %v", err)
			}

		}()

		func() {

			if err := db.Setup(consts.CloudAws, "regionAws", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}

			fakeData := &statefile.StorageDocument{
				Region:        "regionAws",
				ClusterName:   "name_managed",
				ClusterType:   "managed",
				InfraProvider: consts.CloudAws,
				CloudInfra:    &statefile.InfrastructureState{Aws: &statefile.StateConfigurationAws{}},
			}

			err := db.Write(fakeData)
			if err != nil {
				t.Fatalf("Error shouln't happen: %v", err)
			}
			err = db.Write(fakeData)
			if err != nil {
				t.Fatalf("Error shouln't happen on second Write: %v", err)
			}

		}()

		func() {

			if err := db.Setup(consts.CloudAws, "regionAws", "name_ha", consts.ClusterTypeSelfMang); err != nil {
				t.Fatal(err)
			}

			fakeData := &statefile.StorageDocument{
				Region:        "regionAws",
				ClusterName:   "name_ha",
				ClusterType:   "selfmanaged",
				InfraProvider: consts.CloudAws,
				CloudInfra:    &statefile.InfrastructureState{Aws: &statefile.StateConfigurationAws{}},
				K8sBootstrap:  &statefile.KubernetesBootstrapState{K3s: &statefile.StateConfigurationK3s{}},
			}

			err := db.Write(fakeData)
			if err != nil {
				t.Fatalf("Error shouln't happen: %v", err)
			}

		}()
	})

	t.Run("fetch cluster Infos", func(t *testing.T) {
		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "all", "clusterType": ""})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
			assert.Check(t, len(m[consts.ClusterTypeSelfMang]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 2)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "aws", "clusterType": "selfmanaged"})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
			assert.Check(t, len(m[consts.ClusterTypeSelfMang]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 0)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "azure", "clusterType": "managed"})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
			assert.Check(t, len(m[consts.ClusterTypeSelfMang]) == 0)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 1)
		}(t)
	})
}

func TestExportImport(t *testing.T) {

	var bkpData *storage.StateExportImport

	t.Run("Export all", func(t *testing.T) {
		var _expect storage.StateExportImport = storage.StateExportImport{
			Clusters: []*statefile.StorageDocument{
				{
					Region:        "regionAws",
					ClusterName:   "name_ha",
					ClusterType:   "selfmanaged",
					InfraProvider: consts.CloudAws,
					CloudInfra:    &statefile.InfrastructureState{Aws: &statefile.StateConfigurationAws{}},
					K8sBootstrap:  &statefile.KubernetesBootstrapState{K3s: &statefile.StateConfigurationK3s{}},
				},
				{
					Region:        "regionAws",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAws,
					CloudInfra:    &statefile.InfrastructureState{Aws: &statefile.StateConfigurationAws{}},
				},

				{
					Region:        "regionAzure",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAzure,
					CloudInfra:    &statefile.InfrastructureState{Azure: &statefile.StateConfigurationAzure{}},
				},
			},
		}

		dump.Println(_expect)
		if _got, err := db.Export(map[consts.KsctlSearchFilter]string{}); err != nil {
			t.Fatal(err)
		} else {
			dump.Println(_got)
			bkpData = _got // storing the exported data

			assert.Check(t, _got != nil && _got.Clusters != nil)
			assert.Check(t, len(_got.Clusters) == len(_expect.Clusters))

			for _, g := range _got.Clusters {
				assert.Check(t, g != nil)
				v := false
				for _, e := range _expect.Clusters {
					if reflect.DeepEqual(e, g) {
						v = true
					}
				}
				assert.Check(t, v == true, "didn't find the exepcted cluster state")
			}
		}
	})

	t.Run("delete all storage", func(t *testing.T) {

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "all", "clusterType": ""})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
		}(t)

		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("import data", func(t *testing.T) {

		if err := db.Import(bkpData); err != nil {
			t.Fatal(err)
		}

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "all", "clusterType": ""})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
			assert.Check(t, len(m[consts.ClusterTypeSelfMang]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 2)
		}(t)
	})

	t.Run("Export specific cluster", func(t *testing.T) {

		var _expect storage.StateExportImport = storage.StateExportImport{
			Clusters: []*statefile.StorageDocument{
				{
					Region:        "regionAzure",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAzure,
					CloudInfra:    &statefile.InfrastructureState{Azure: &statefile.StateConfigurationAzure{}},
				},
			},
		}

		if _got, err := db.Export(map[consts.KsctlSearchFilter]string{
			consts.Cloud:       string(consts.CloudAzure),
			consts.ClusterType: string(consts.ClusterTypeMang),
			consts.Name:        "name_managed",
			consts.Region:      "regionAzure",
		}); err != nil {
			t.Fatal(err)
		} else {
			dump.Println(_got)
			assert.Check(t, _got != nil && _got.Clusters != nil)
			assert.Check(t, len(_got.Clusters) == len(_expect.Clusters))

			for _, g := range _got.Clusters {
				assert.Check(t, g != nil)
				v := false
				for _, e := range _expect.Clusters {
					if reflect.DeepEqual(e, g) {
						v = true
					}
				}
				assert.Check(t, v == true, "didn't find the exepcted cluster state")
			}
		}
	})
}

func TestExternalUsageFunctions(t *testing.T) {

	locGot, okGot := db.PresentDirectory([]string{"demo03", "234"})
	assert.Equal(t, locGot, filepath.Join("demo03", "234"))
	assert.Equal(t, okGot, false)

	home, _ := os.UserHomeDir()
	locGot, okGot = db.PresentDirectory([]string{home})
	assert.Equal(t, locGot, home)
	assert.Equal(t, okGot, true)

	err := db.CreateDirectory([]string{os.TempDir(), "ksctl-local-store-test"})
	assert.NilError(t, err)

	locGot, err = db.CreateFileIfNotPresent([]string{os.TempDir(), "ksctl-local-store-test", "abcd.yml"})
	assert.NilError(t, err)

	assert.Equal(t, locGot, filepath.Join(os.TempDir(), "ksctl-local-store-test", "abcd.yml"))
}

func TestKill(t *testing.T) {
	if err := db.Kill(); err != nil {
		t.Fatal(err)
	}
}
