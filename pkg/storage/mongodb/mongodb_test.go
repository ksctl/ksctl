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

package mongodb

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"

	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"gotest.tools/v3/assert"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
)

var (
	db           storage.Storage
	parentCtx    context.Context
	parentLogger logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
	parentCtx = context.WithValue(parentCtx, consts.KsctlContextUserID, "fake")

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

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

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
	//recover()
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

	fakeData := &statefile.StorageDocument{
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

		gotFakeData.ID = [12]byte{} // to make the ID same as out fake as it gets updated
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

	fakeData := &statefile.CredentialsDocument{
		InfraProvider: consts.CloudAzure,
		Azure: &statefile.CredentialsAzure{
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

		gotFakeData.ID = [12]byte{} // to make the ID same as out fake as it gets updated
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

			fakeData := &statefile.StorageDocument{
				Region:        "regionAzure",
				ClusterName:   "name_managed",
				InfraProvider: consts.CloudAzure,
				ClusterType:   "managed",
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
				InfraProvider: consts.CloudAws,
				ClusterType:   "managed",
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

			if err := db.Setup(consts.CloudAws, "regionAws", "name_ha", consts.ClusterTypeHa); err != nil {
				t.Fatal(err)
			}

			fakeData := &statefile.StorageDocument{
				Region:        "regionAws",
				ClusterName:   "name_ha",
				InfraProvider: consts.CloudAws,
				ClusterType:   "ha",
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
			assert.Check(t, len(m[consts.ClusterTypeHa]) == 1)
			assert.Check(t, len(m[consts.ClusterTypeMang]) == 2)
		}(t)

		func(t *testing.T) {
			m, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{"cloud": "aws", "clusterType": "ha"})

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

	var bkpData *storage.StateExportImport

	t.Run("Export all", func(t *testing.T) {
		var _expect = storage.StateExportImport{
			Credentials: []*statefile.CredentialsDocument{
				{
					Azure: &statefile.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
			Clusters: []*statefile.StorageDocument{
				{
					Region:        "regionCivo",
					ClusterName:   "name_ha",
					ClusterType:   "ha",
					InfraProvider: consts.CloudAws,
					CloudInfra:    &statefile.InfrastructureState{Aws: &statefile.StateConfigurationAws{}},
					K8sBootstrap:  &statefile.KubernetesBootstrapState{K3s: &statefile.StateConfigurationK3s{}},
				},
				{
					Region:        "regionCivo",
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

		f := func(factory storage.Storage) *Store {
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

		var _expect = storage.StateExportImport{
			Credentials: []*statefile.CredentialsDocument{
				{
					Azure: &statefile.CredentialsAzure{
						ClientID: "client_id",
					},
					InfraProvider: consts.CloudAzure,
				},
			},
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
			if err := db.Setup(consts.CloudAws, "regionAws", "name_ha", consts.ClusterTypeHa); err != nil {
				t.Fatal(err)
			}
			if err := db.DeleteCluster(); err != nil {
				t.Fatal(err)
			}
		}()

		func() {
			if err := db.Setup(consts.CloudAws, "regionAws", "name_managed", consts.ClusterTypeMang); err != nil {
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
