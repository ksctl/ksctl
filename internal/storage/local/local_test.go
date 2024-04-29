package local

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/goccy/go-json"
	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

var (
	db resources.StorageFactory

	dir = fmt.Sprintf("%s ksctl-local-store-test", os.TempDir())
)

func TestMain(m *testing.M) {
	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)

	_ = m.Run()

	_ = os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-local-store-test")
}

func TestInitStorage(t *testing.T) {
	db = InitStorage(-1, os.Stdout)
	err := db.Setup(consts.CloudAzure, "region", "name", consts.ClusterTypeHa)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Connect(context.WithValue(context.Background(), "USERID", "XYz")); err != nil {
		t.Fatal(err)
	}
}

func TestReader(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	v := &types.StorageDocument{}

	d, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(dir+helpers.PathSeparator+"fake.json", d, 0755); err != nil {
		t.Fatal(err)
	}
}

func TestGenOsClusterPath(t *testing.T) {
	expectedHa := strings.Join([]string{os.TempDir(), "ksctl-local-store-test", ".ksctl", "state", "civo", "ha", "name region"}, helpers.PathSeparator)
	expectedManaged := strings.Join([]string{os.TempDir(), "ksctl-local-store-test", ".ksctl", "state", "azure", "managed", "name region"}, helpers.PathSeparator)
	expectedCreds := strings.Join([]string{os.TempDir(), "ksctl-local-store-test", ".ksctl", "credentials"}, helpers.PathSeparator)

	gotH := genOsClusterPath(false, string(consts.CloudCivo), string(consts.ClusterTypeHa), "name region")
	if gotH != expectedHa {
		t.Fatalf("expected %s; but got %s\n", expectedHa, gotH)
	}

	gotM := genOsClusterPath(false, string(consts.CloudAzure), string(consts.ClusterTypeMang), "name region")
	if gotM != expectedManaged {
		t.Fatalf("expected %s; but got %s\n", expectedManaged, gotM)
	}

	gotC := genOsClusterPath(true)
	if gotC != expectedCreds {
		t.Fatalf("expected %s; but got %s\n", expectedCreds, gotC)
	}
}

func TestStore_RWD(t *testing.T) {
	if _, err := db.Read(); err == nil {
		t.Fatal("Error should occur as there is no folder created")
	}
	if err := db.DeleteCluster(); err == nil {
		t.Fatalf("Error should happen on deleting cluster info")
	}

	if err := db.AlreadyCreated(consts.CloudAzure, "region", "name", consts.ClusterTypeHa); err == nil {
		t.Fatalf("Error should happen on checking for presence of the cluster")
	}

	fakeData := &types.StorageDocument{
		Region:      "region",
		ClusterName: "name",
		ClusterType: "ha",
	}
	err := db.Write(fakeData)
	if err != nil {
		t.Fatalf("Error shouln't happen: %v", err)
	}

	if err := db.AlreadyCreated(consts.CloudAzure, "region", "name", consts.ClusterTypeHa); err != nil {
		t.Fatalf("Error shouldn't happen on checking for presence of the cluster: %v", err)
	}

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

	if err := db.DeleteCluster(); err != nil {
		t.Fatalf("Error shouln't happen on deleting cluster info: %v", err)
	}
}

func TestStore_RWDCredentials(t *testing.T) {
	if _, err := db.ReadCredentials(consts.CloudAzure); err == nil {
		t.Fatal("Error should occur as there is no folder created")
	}

	fakeData := &types.CredentialsDocument{
		Azure: &types.CredentialsAzure{
			ClientID: "client_id",
		},
		InfraProvider: consts.CloudAzure,
	}
	err := db.WriteCredentials(consts.CloudAzure, fakeData)
	if err != nil {
		t.Fatalf("Error shouln't happen: %v", err)
	}

	if gotFakeData, err := db.ReadCredentials(consts.CloudAzure); err != nil {
		t.Fatalf("Error shouln't happen on reading file: %v", err)
	} else {
		fmt.Printf("%#+v\n", gotFakeData)

		if !reflect.DeepEqual(gotFakeData, fakeData) {
			t.Fatalf("Written data doesn't match Reading")
		}
	}
}

func TestGetClusterInfo(t *testing.T) {

	t.Run("Setup some demo clusters", func(t *testing.T) {

		func() {

			if err := db.Setup(consts.CloudAzure, "regionAzure", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}

			fakeData := &types.StorageDocument{
				Region:        "regionAzure",
				ClusterName:   "name_managed",
				ClusterType:   "managed",
				InfraProvider: consts.CloudAzure,
				CloudInfra:    &types.InfrastructureState{Azure: &types.StateConfigurationAzure{}},
			}

			err := db.Write(fakeData)
			if err != nil {
				t.Fatalf("Error shouln't happen: %v", err)
			}

		}()

		func() {

			if err := db.Setup(consts.CloudCivo, "regionCivo", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}

			fakeData := &types.StorageDocument{
				Region:        "regionCivo",
				ClusterName:   "name_managed",
				ClusterType:   "managed",
				InfraProvider: consts.CloudCivo,
				CloudInfra:    &types.InfrastructureState{Civo: &types.StateConfigurationCivo{}},
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

			if err := db.Setup(consts.CloudCivo, "regionCivo", "name_ha", consts.ClusterTypeHa); err != nil {
				t.Fatal(err)
			}

			fakeData := &types.StorageDocument{
				Region:        "regionCivo",
				ClusterName:   "name_ha",
				ClusterType:   "ha",
				InfraProvider: consts.CloudCivo,
				CloudInfra:    &types.InfrastructureState{Civo: &types.StateConfigurationCivo{}},
				K8sBootstrap:  &types.KubernetesBootstrapState{K3s: &types.StateConfigurationK3s{}},
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
			assert.Check(t, len(m[consts.ClusterTypeHa]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 2)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "civo", "clusterType": "ha"})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
			assert.Check(t, len(m[consts.ClusterTypeHa]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 0)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "azure", "clusterType": "managed"})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
			assert.Check(t, len(m[consts.ClusterTypeHa]) == 0)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 1)
		}(t)
	})
}

func TestExportImport(t *testing.T) {

	var bkpData *resources.StorageStateExportImport

	t.Run("Export all", func(t *testing.T) {
		var _expect resources.StorageStateExportImport = resources.StorageStateExportImport{
			Credentials: []*types.CredentialsDocument{
				&types.CredentialsDocument{
					Azure: &types.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
			Clusters: []*types.StorageDocument{
				&types.StorageDocument{
					Region:        "regionCivo",
					ClusterName:   "name_ha",
					ClusterType:   "ha",
					InfraProvider: consts.CloudCivo,
					CloudInfra:    &types.InfrastructureState{Civo: &types.StateConfigurationCivo{}},
					K8sBootstrap:  &types.KubernetesBootstrapState{K3s: &types.StateConfigurationK3s{}},
				},
				&types.StorageDocument{
					Region:        "regionCivo",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudCivo,
					CloudInfra:    &types.InfrastructureState{Civo: &types.StateConfigurationCivo{}},
				},

				&types.StorageDocument{
					Region:        "regionAzure",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAzure,
					CloudInfra:    &types.InfrastructureState{Azure: &types.StateConfigurationAzure{}},
				},
			},
		}

		dump.Println(_expect)
		if _got, err := db.Export(map[consts.KsctlSearchFilter]string{}); err != nil {
			t.Fatal(err)
		} else {
			dump.Println(_got)
			bkpData = _got // storing the exported data

			assert.Check(t, _got != nil && _got.Clusters != nil && _got.Credentials != nil)
			assert.Check(t, len(_got.Credentials) == len(_expect.Credentials))
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
			for _, g := range _got.Credentials {
				assert.Check(t, g != nil)
				v := false
				for _, e := range _expect.Credentials {
					if reflect.DeepEqual(e, g) {
						v = true
					}
				}
				assert.Check(t, v == true, "didn't find the exepcted credentials state")
			}
		}
	})

	t.Run("delete all resources", func(t *testing.T) {

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "all", "clusterType": ""})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
		}(t)

		if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-local-store-test"); err != nil {
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
			assert.Check(t, len(m[consts.ClusterTypeHa]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 2)
		}(t)
	})

	t.Run("Export specific cluster", func(t *testing.T) {

		var _expect resources.StorageStateExportImport = resources.StorageStateExportImport{
			Credentials: []*types.CredentialsDocument{
				&types.CredentialsDocument{
					Azure: &types.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
			Clusters: []*types.StorageDocument{
				&types.StorageDocument{
					Region:        "regionAzure",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAzure,
					CloudInfra:    &types.InfrastructureState{Azure: &types.StateConfigurationAzure{}},
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
			assert.Check(t, _got != nil && _got.Clusters != nil && _got.Credentials != nil)
			assert.Check(t, len(_got.Credentials) == len(_expect.Credentials))
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
			for _, g := range _got.Credentials {
				assert.Check(t, g != nil)
				v := false
				for _, e := range _expect.Credentials {
					if reflect.DeepEqual(e, g) {
						v = true
					}
				}
				assert.Check(t, v == true, "didn't find the exepcted credentials state")
			}
		}
	})
}

func TestExternalUsageFunctions(t *testing.T) {
	extDb := db.(*Store)

	locGot, okGot := extDb.PresentDirectory([]string{"demo03", "234"})
	assert.Equal(t, locGot, "demo03"+helpers.PathSeparator+"234")
	assert.Equal(t, okGot, false)

	locGot, okGot = extDb.PresentDirectory([]string{helpers.GetUserName()})
	assert.Equal(t, locGot, helpers.GetUserName())
	assert.Equal(t, okGot, true)

	err := extDb.CreateDirectory([]string{os.TempDir(), "ksctl-local-store-test"})
	assert.Assert(t, err == nil)

	locGot, err = extDb.CreateFileIfNotPresent([]string{os.TempDir(), "ksctl-local-store-test", "abcd.yml"})
	assert.Assert(t, err == nil)
	assert.Equal(t, locGot, os.TempDir()+helpers.PathSeparator+"ksctl-local-store-test"+helpers.PathSeparator+"abcd.yml")
}

func TestKill(t *testing.T) {
	if err := db.Kill(); err != nil {
		t.Fatal(err)
	}
}
