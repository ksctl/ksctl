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
	"github.com/ksctl/ksctl/pkg/addons"
	"github.com/ksctl/ksctl/pkg/utilities"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"
	localstate "github.com/ksctl/ksctl/pkg/storage/host"
	"gotest.tools/v3/assert"
)

var (
	fakeClientManaged *Provider
	storeManaged      storage.Storage = nil

	fakeClientVars *Provider
	parentCtx      context.Context
	parentLogger   logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-local-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo",
			Region:      "LOCAL",
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
					Name:  "kind",
					IsCNI: true,
				},
			}, false, "kind", nil,
		},
		{
			addons.ClusterAddons{}, false, "kind", nil,
		},
		{
			nil, false, "kind", nil,
		},
		{
			addons.ClusterAddons{
				{
					Label:  "kind",
					Name:   "heheheh",
					Config: utilities.Ptr(`{"key":"value"}`),
				},
			}, false, "kind", nil,
		},
		{
			addons.ClusterAddons{
				{
					Label: "kind",
					Name:  "heheheh",
				},
			}, false, "kind", nil,
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
	_ = storeManaged.Connect()

	fakeClientManaged, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo-managed",
			Region:      "LOCAL",
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

		assert.Equal(t, fakeClientManaged.NewManagedCluster(2), nil, "managed cluster should be created")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Local.Nodes, 2, "missmatch of no of nodes")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Local.B.KubernetesVer, fakeClientManaged.K8sVersion, "k8s version does not match")
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientManaged.GetStateFile()
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(fakeClientManaged.state)
		assert.DeepEqual(t, string(got), expected)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []logger.ClusterDataForLogging{
			{
				Name:          fakeClientManaged.ClusterName,
				CloudProvider: consts.CloudLocal,
				ClusterType:   consts.ClusterTypeMang,
				Region:        fakeClientManaged.Region,
				NoMgt:         fakeClientManaged.state.CloudInfra.Local.Nodes,
				Mgt:           logger.VMData{VMSize: fakeClientManaged.state.CloudInfra.Local.ManagedNodeSize},

				K8sDistro:  "kind",
				K8sVersion: fakeClientManaged.state.CloudInfra.Local.B.KubernetesVer,
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	assert.Equal(t, fakeClientManaged.DelManagedCluster(), nil, "managed cluster should be deleted")
}
