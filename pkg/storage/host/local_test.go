package host

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"gotest.tools/v3/assert"

	"github.com/goccy/go-json"
	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	db types.StorageFactory

	parentCtx    context.Context
	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
	dir                              = filepath.Join(os.TempDir(), "ksctl-local-store-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
	parentCtx = context.WithValue(parentCtx, consts.KsctlCustomDirLoc, dir)

	_ = m.Run()

	_ = os.RemoveAll(dir)
}

func TestInitStorage(t *testing.T) {
	db = NewClient(parentCtx, parentLogger)
	err := db.Setup(consts.CloudAzure, "region", "name", consts.ClusterTypeHa)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Connect(); err != nil {
		t.Fatal(err)
	}
}

func TestReader(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	v := &storageTypes.StorageDocument{}

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
		"civo",
		"ha",
		"name region"}...)
	expectedManaged := filepath.Join([]string{
		os.TempDir(),
		"ksctl-local-store-test",
		".ksctl",
		"state",
		"azure",
		"managed",
		"name region"}...)
	expectedCreds := filepath.Join([]string{
		os.TempDir(),
		"ksctl-local-store-test",
		".ksctl",
		"credentials"}...)

	gotH, err := genOsClusterPath(false, string(consts.CloudCivo), string(consts.ClusterTypeHa), "name region")
	assert.NilError(t, err)
	if gotH != expectedHa {
		t.Fatalf("expected %s; but got %s\n", expectedHa, gotH)
	}

	gotM, err := genOsClusterPath(false, string(consts.CloudAzure), string(consts.ClusterTypeMang), "name region")
	assert.NilError(t, err)
	if gotM != expectedManaged {
		t.Fatalf("expected %s; but got %s\n", expectedManaged, gotM)
	}

	gotC, err := genOsClusterPath(true)
	assert.NilError(t, err)
	if gotC != expectedCreds {
		t.Fatalf("expected %s; but got %s\n", expectedCreds, gotC)
	}
}

func TestStore_RWD(t *testing.T) {
	if _, err := db.Read(); err == nil {
		t.Fatal("Error should occur as there is no folder created")
	}

	if err := db.DeleteCluster(); err == nil {
		t.Fatal("Error should happen on deleting cluster info")
	}

	if err := db.AlreadyCreated(consts.CloudAzure, "region", "name", consts.ClusterTypeHa); err == nil {
		t.Fatalf("Error should happen on checking for presence of the cluster")
	}

	fakeData := &storageTypes.StorageDocument{
		Region:      "region",
		ClusterName: "name",
		ClusterType: "ha",
	}
	err := db.Write(fakeData)
	assert.NilError(t, err)

	err = db.AlreadyCreated(consts.CloudAzure, "region", "name", consts.ClusterTypeHa)
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

func TestStore_RWDCredentials(t *testing.T) {
	if _, err := db.ReadCredentials(consts.CloudAzure); err == nil {
		t.Fatal("Error should occur as there is no folder created")
	}

	t.Run("azure", func(t *testing.T) {
		fakeDataAzure := &storageTypes.CredentialsDocument{
			Azure: &storageTypes.CredentialsAzure{
				ClientID: "client_id",
			},
			InfraProvider: consts.CloudAzure,
		}
		err := db.WriteCredentials(consts.CloudAzure, fakeDataAzure)
		if err != nil {
			t.Fatalf("Error shouln't happen: %v", err)
		}

		if gotFakeData, err := db.ReadCredentials(consts.CloudAzure); err != nil {
			t.Fatalf("Error shouln't happen on reading file: %v", err)
		} else {
			fmt.Printf("%#+v\n", gotFakeData)

			if !reflect.DeepEqual(gotFakeData, fakeDataAzure) {
				t.Fatalf("Written data doesn't match Reading")
			}
		}
	})

	t.Run("aws", func(t *testing.T) {
		fakeDataAws := &storageTypes.CredentialsDocument{
			Aws: &storageTypes.CredentialsAws{
				AccessKeyId:     "access_key",
				SecretAccessKey: "secret",
			},
			InfraProvider: consts.CloudAws,
		}
		err := db.WriteCredentials(consts.CloudAws, fakeDataAws)
		if err != nil {
			t.Fatalf("Error shouln't happen: %v", err)
		}

		if gotFakeData, err := db.ReadCredentials(consts.CloudAws); err != nil {
			t.Fatalf("Error shouln't happen on reading file: %v", err)
		} else {
			fmt.Printf("%#+v\n", gotFakeData)

			if !reflect.DeepEqual(gotFakeData, fakeDataAws) {
				t.Fatalf("Written data doesn't match Reading")
			}
		}
	})
}

func TestGetClusterInfo(t *testing.T) {

	t.Run("Setup some demo clusters", func(t *testing.T) {

		func() {

			if err := db.Setup(consts.CloudAzure, "regionAzure", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}

			fakeData := &storageTypes.StorageDocument{
				Region:        "regionAzure",
				ClusterName:   "name_managed",
				ClusterType:   "managed",
				InfraProvider: consts.CloudAzure,
				CloudInfra:    &storageTypes.InfrastructureState{Azure: &storageTypes.StateConfigurationAzure{}},
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

			fakeData := &storageTypes.StorageDocument{
				Region:        "regionCivo",
				ClusterName:   "name_managed",
				ClusterType:   "managed",
				InfraProvider: consts.CloudCivo,
				CloudInfra:    &storageTypes.InfrastructureState{Civo: &storageTypes.StateConfigurationCivo{}},
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

			fakeData := &storageTypes.StorageDocument{
				Region:        "regionCivo",
				ClusterName:   "name_ha",
				ClusterType:   "ha",
				InfraProvider: consts.CloudCivo,
				CloudInfra:    &storageTypes.InfrastructureState{Civo: &storageTypes.StateConfigurationCivo{}},
				K8sBootstrap:  &storageTypes.KubernetesBootstrapState{K3s: &storageTypes.StateConfigurationK3s{}},
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

	var bkpData *types.StorageStateExportImport

	t.Run("Export all", func(t *testing.T) {
		var _expect types.StorageStateExportImport = types.StorageStateExportImport{
			Credentials: []*storageTypes.CredentialsDocument{
				{
					Azure: &storageTypes.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
				{
					Aws: &storageTypes.CredentialsAws{
						AccessKeyId:     "access_key",
						SecretAccessKey: "secret",
					},
					InfraProvider: consts.CloudAws,
				},
			},
			Clusters: []*storageTypes.StorageDocument{
				{
					Region:        "regionCivo",
					ClusterName:   "name_ha",
					ClusterType:   "ha",
					InfraProvider: consts.CloudCivo,
					CloudInfra:    &storageTypes.InfrastructureState{Civo: &storageTypes.StateConfigurationCivo{}},
					K8sBootstrap:  &storageTypes.KubernetesBootstrapState{K3s: &storageTypes.StateConfigurationK3s{}},
				},
				{
					Region:        "regionCivo",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudCivo,
					CloudInfra:    &storageTypes.InfrastructureState{Civo: &storageTypes.StateConfigurationCivo{}},
				},

				{
					Region:        "regionAzure",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAzure,
					CloudInfra:    &storageTypes.InfrastructureState{Azure: &storageTypes.StateConfigurationAzure{}},
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
			assert.Check(t, len(m[consts.ClusterTypeHa]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 2)
		}(t)
	})

	t.Run("Export specific cluster", func(t *testing.T) {

		var _expect types.StorageStateExportImport = types.StorageStateExportImport{
			Credentials: []*storageTypes.CredentialsDocument{
				{
					Azure: &storageTypes.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
			Clusters: []*storageTypes.StorageDocument{
				{
					Region:        "regionAzure",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudAzure,
					CloudInfra:    &storageTypes.InfrastructureState{Azure: &storageTypes.StateConfigurationAzure{}},
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
	assert.Equal(t, locGot, filepath.Join("demo03", "234"))
	assert.Equal(t, okGot, false)

	home, _ := os.UserHomeDir()
	locGot, okGot = extDb.PresentDirectory([]string{home})
	assert.Equal(t, locGot, home)
	assert.Equal(t, okGot, true)

	err := extDb.CreateDirectory([]string{os.TempDir(), "ksctl-local-store-test"})
	assert.NilError(t, err)

	locGot, err = extDb.CreateFileIfNotPresent([]string{os.TempDir(), "ksctl-local-store-test", "abcd.yml"})
	assert.NilError(t, err)

	assert.Equal(t, locGot, filepath.Join(os.TempDir(), "ksctl-local-store-test", "abcd.yml"))
}

func TestKill(t *testing.T) {
	if err := db.Kill(); err != nil {
		t.Fatal(err)
	}
}
