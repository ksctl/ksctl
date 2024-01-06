package mongodb

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/kubesimplify/ksctl/internal/storage/types"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	db resources.StorageFactory
)

func TestMain(m *testing.M) {
	_ = m.Run()
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
		Region:        "region",
		ClusterName:   "name",
		ClusterType:   "ha",
		InfraProvider: consts.CloudAzure,
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

		gotFakeData.ID = primitive.ObjectID([12]byte{}) // to make the ID same as out fake as it gets updated
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
		t.Fatalf("Error should occur as there is no folder created, %v", err)
	}

	fakeData := &types.CredentialsDocument{
		InfraProvider: consts.CloudAzure,
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

		gotFakeData.ID = primitive.ObjectID([12]byte{}) // to make the ID same as out fake as it gets updated
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
				InfraProvider: consts.CloudAzure,
				ClusterType:   "managed",
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
				InfraProvider: consts.CloudCivo,
				ClusterType:   "managed",
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
				InfraProvider: consts.CloudCivo,
				ClusterType:   "ha",
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

	t.Run("delete all", func(t *testing.T) {
		func() {
			if err := db.Setup(consts.CloudCivo, "regionCivo", "name_ha", consts.ClusterTypeHa); err != nil {
				t.Fatal(err)
			}
			if err := db.DeleteCluster(); err != nil {
				t.Fatal(err)
			}
		}()

		func() {
			if err := db.Setup(consts.CloudCivo, "regionCivo", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}
			if err := db.DeleteCluster(); err != nil {
				t.Fatal(err)
			}
		}()

		func() {
			if err := db.Setup(consts.CloudAzure, "regionAzure", "name_managed", consts.ClusterTypeMang); err != nil {
				t.Fatal(err)
			}
			if err := db.DeleteCluster(); err != nil {
				t.Fatal(err)
			}
		}()
	})
}

func TestKill(t *testing.T) {
	if err := db.Kill(); err != nil {
		t.Fatal(err)
	}
}
