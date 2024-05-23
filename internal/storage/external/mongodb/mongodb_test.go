package mongodb

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"gotest.tools/v3/assert"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	db           types.StorageFactory
	parentCtx    context.Context
	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
	parentCtx = context.WithValue(parentCtx, "USERID", "fake")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(parentCtx, "mongo", image.PullOptions{})
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		panic(err)
	}

	defer reader.Close()

	resp, err := cli.ContainerCreate(parentCtx, &container.Config{
		Image: "mongo",
		Env:   []string{"MONGO_INITDB_ROOT_USERNAME=root", "MONGO_INITDB_ROOT_PASSWORD=1234"},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"27017/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "27017",
				},
			},
		},
	}, nil, nil, "mongodb")
	if err != nil {
		panic(err)
	}

	// Start the container
	if err := cli.ContainerStart(parentCtx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	err = os.Setenv("MONGODB_URI", "mongodb://root:1234@localhost:27017")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := cli.ContainerRemove(parentCtx, resp.ID, container.RemoveOptions{Force: true}); err != nil {
			panic(err)
		}
	}()
	_ = m.Run()

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

	fakeData := &storageTypes.StorageDocument{
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

	fakeData := &storageTypes.CredentialsDocument{
		InfraProvider: consts.CloudAzure,
		Azure: &storageTypes.CredentialsAzure{
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

			fakeData := &storageTypes.StorageDocument{
				Region:        "regionAzure",
				ClusterName:   "name_managed",
				InfraProvider: consts.CloudAzure,
				ClusterType:   "managed",
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
				InfraProvider: consts.CloudCivo,
				ClusterType:   "managed",
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
				InfraProvider: consts.CloudCivo,
				ClusterType:   "ha",
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
				&storageTypes.CredentialsDocument{
					Azure: &storageTypes.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
			Clusters: []*storageTypes.StorageDocument{
				&storageTypes.StorageDocument{
					Region:        "regionCivo",
					ClusterName:   "name_ha",
					ClusterType:   "ha",
					InfraProvider: consts.CloudCivo,
					CloudInfra:    &storageTypes.InfrastructureState{Civo: &storageTypes.StateConfigurationCivo{}},
					K8sBootstrap:  &storageTypes.KubernetesBootstrapState{K3s: &storageTypes.StateConfigurationK3s{}},
				},
				&storageTypes.StorageDocument{
					Region:        "regionCivo",
					ClusterName:   "name_managed",
					ClusterType:   "managed",
					InfraProvider: consts.CloudCivo,
					CloudInfra:    &storageTypes.InfrastructureState{Civo: &storageTypes.StateConfigurationCivo{}},
				},

				&storageTypes.StorageDocument{
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
			bkpData = _got // storing the data
			assert.Check(t, _got != nil && _got.Clusters != nil && _got.Credentials != nil)
			assert.Check(t, len(_got.Credentials) == len(_expect.Credentials))
			assert.Check(t, len(_got.Clusters) == len(_expect.Clusters))

			for _, g := range _got.Clusters {
				// make the _ID as 0
				g.ID = [12]byte{}
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
				g.ID = [12]byte{}
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

		f := func(factory types.StorageFactory) *Store {
			switch o := factory.(type) {
			case *Store:
				return o
			default:
				return nil
			}
		}(db)

		err := f.databaseClient.Drop(storeCtx)
		if err != nil {
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
				&storageTypes.CredentialsDocument{
					Azure: &storageTypes.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
			Clusters: []*storageTypes.StorageDocument{
				&storageTypes.StorageDocument{
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
				g.ID = [12]byte{}
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
				g.ID = [12]byte{}
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

func TestDelete(t *testing.T) {

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
