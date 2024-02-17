package local

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

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
				Region:      "regionAzure",
				ClusterName: "name_managed",
				ClusterType: "managed",
				CloudInfra:  &types.InfrastructureState{Azure: &types.StateConfigurationAzure{}},
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
				Region:      "regionCivo",
				ClusterName: "name_managed",
				ClusterType: "managed",
				CloudInfra:  &types.InfrastructureState{Civo: &types.StateConfigurationCivo{}},
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
				Region:       "regionCivo",
				ClusterName:  "name_ha",
				ClusterType:  "ha",
				CloudInfra:   &types.InfrastructureState{Civo: &types.StateConfigurationCivo{}},
				K8sBootstrap: &types.KubernetesBootstrapState{K3s: &types.StateConfigurationK3s{}},
			}

			err := db.Write(fakeData)
			if err != nil {
				t.Fatalf("Error shouln't happen: %v", err)
			}

		}()
	})

	t.Run("fetch cluster Infos", func(t *testing.T) {
		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[string]string{"cloud": "all", "clusterType": ""})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[string]string{"cloud": "civo", "clusterType": "ha"})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[string]string{"cloud": "azure", "clusterType": "managed"})

			if err != nil {
				t.Fatal(err)
			}
			dump.Println(m)
		}(t)
	})
}

func TestKill(t *testing.T) {
	if err := db.Kill(); err != nil {
		t.Fatal(err)
	}
}
