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

package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/provider"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/utilities"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"
	localstate "github.com/ksctl/ksctl/v2/pkg/storage/host"
	"gotest.tools/v3/assert"
)

var (
	fakeClientManaged *Provider
	storeManaged      storage.Storage = nil

	fakeClientVars *Provider
	parentCtx      context.Context
	ksc                          = context.Background()
	parentLogger   logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-local-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = NewClient(
		parentCtx,
		parentLogger,
		ksc,
		controller.Metadata{
			ClusterName: "demo",
			Region:      "LOCAL",
			ClusterType: consts.ClusterTypeMang,
			Provider:    consts.CloudLocal,
		},
		&statefile.StorageDocument{},
		storeManaged,
		ProvideClient,
	)

	exitVal := m.Run()
	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
	os.Exit(exitVal)
}

func TestCNIandApp(t *testing.T) {
	testCases := []struct {
		Addon           addons.ClusterAddons
		Valid           bool
		managedAddonCNI string
		managedAddonApp map[string]map[string]*string
	}{
		{
			addons.ClusterAddons{
				{
					Label: "ksctl",
					Name:  "cilium",
					IsCNI: true,
				},
				{
					Label: "kind",
					Name:  "none",
					IsCNI: true,
				},
			}, true, "none", nil,
		},
		{
			addons.ClusterAddons{
				{
					Label: "kind",
					Name:  "kindnet",
					IsCNI: true,
				},
			}, false, "kindnet", nil,
		},
		{
			addons.ClusterAddons{}, false, "kindnet", nil,
		},
		{
			nil, false, "kindnet", nil,
		},
		{
			addons.ClusterAddons{
				{
					Label:  "kind",
					Name:   "heheheh",
					Config: utilities.Ptr(`{"key":"value"}`),
				},
			}, false, "kindnet", nil,
		},
		{
			addons.ClusterAddons{
				{
					Label: "kind",
					Name:  "heheheh",
				},
			}, false, "kindnet", nil,
		},
	}

	for _, v := range testCases {
		got := fakeClientVars.ManagedAddons(v.Addon)
		assert.Equal(t, got, v.Valid, "missmatch in return value")
		assert.Equal(t, fakeClientVars.managedAddonCNI, v.managedAddonCNI, "missmatch in managedAddonCNI")
		assert.DeepEqual(t, fakeClientVars.managedAddonApp, v.managedAddonApp)
	}
}

func TestGenerateConfig(t *testing.T) {
	if _, err := fakeClientVars.generateConfig(0, 0, false); err == nil {
		t.Fatalf("It should throw err as no of controlplane is 0")
	}

	valid := map[string]string{
		strconv.Itoa(1) + " " + strconv.Itoa(1): `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
networking:
  disableDefaultCNI: false
...`,

		strconv.Itoa(0) + " " + strconv.Itoa(1): `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
networking:
  disableDefaultCNI: false
...`,
	}
	for key, val := range valid {
		inp := strings.Split(key, " ")
		noWP, _ := strconv.Atoi(inp[0])
		noCP, _ := strconv.Atoi(inp[1])
		if raw, _ := fakeClientVars.generateConfig(noWP, noCP, false); !reflect.DeepEqual(raw, []byte(val)) {
			t.Fatalf("Data missmatch for noCP: %d, noWP: %d expected %s but got %s", noCP, noWP, val, string(raw))
		}
	}
}

func TestManagedCluster(t *testing.T) {
	storeManaged = localstate.NewClient(parentCtx, parentLogger)
	_ = storeManaged.Setup(consts.CloudLocal, "LOCAL", "demo-managed", consts.ClusterTypeMang)

	fakeClientManaged, _ = NewClient(
		parentCtx,
		parentLogger,
		ksc,
		controller.Metadata{
			ClusterName: "demo-managed",
			Region:      "LOCAL",
			ClusterType: consts.ClusterTypeMang,
			Provider:    consts.CloudLocal,
		},
		&statefile.StorageDocument{},
		storeManaged,
		ProvideClient,
	)

	t.Run("initState", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.InitState(consts.OperationCreate), nil, "Init must work before")
	})
	t.Run("managed cluster", func(t *testing.T) {

		fakeClientManaged.ManagedK8sVersion("1.27.1")
		fakeClientManaged.Name("fake")
		fakeClientManaged.ManagedAddons(nil)

		assert.Equal(t, fakeClientManaged.NewManagedCluster(2), nil, "managed cluster should be created")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Local.Nodes, 2, "missmatch of no of nodes")
		assert.Equal(t, *fakeClientManaged.state.Versions.Kind, fakeClientManaged.K8sVersion, "k8s version does not match")
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientManaged.GetStateFile()
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(fakeClientManaged.state)
		assert.DeepEqual(t, string(got), expected)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []provider.ClusterData{
			{
				Name:          fakeClientManaged.ClusterName,
				CloudProvider: consts.CloudLocal,
				ClusterType:   consts.ClusterTypeMang,
				Region:        fakeClientManaged.Region,
				NoMgt:         fakeClientManaged.state.CloudInfra.Local.Nodes,
				Mgt:           provider.VMData{VMSize: fakeClientManaged.state.CloudInfra.Local.ManagedNodeSize},

				K8sDistro:  consts.K8sKind,
				K8sVersion: *fakeClientManaged.state.Versions.Kind,
				Apps:       nil,
				Cni:        "Name: kindnet, For: kind, Version: <nil>, KsctlSpecificComponents: map[]",
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	assert.Equal(t, fakeClientManaged.DelManagedCluster(), nil, "managed cluster should be deleted")
}
