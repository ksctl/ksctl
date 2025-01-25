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

package azure

import (
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	localstate "github.com/ksctl/ksctl/pkg/storage/host"
	"gotest.tools/v3/assert"
)

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(
		consts.CloudAzure,
		fakeClientManaged.state.Region,
		fakeClientManaged.state.ClusterName,
		consts.ClusterTypeMang,
	); err != nil {
		t.Fatal(err)
	}
	read, err := storeManaged.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, fakeClientManaged.state, read)
}

func TestManagedCluster(t *testing.T) {
	storeManaged = localstate.NewClient(parentCtx, parentLogger)
	_ = storeManaged.Setup(consts.CloudAzure, "fake", "demo-managed", consts.ClusterTypeMang)
	_ = storeManaged.Connect()

	fakeClientManaged, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo-managed",
			Region:      "fake",
			Provider:    consts.CloudAzure,
		},
		&statefile.StorageDocument{},
		storeManaged,
		ProvideClient,
	)

	fakeClientManaged.ManagedK8sVersion("1.27")
	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, fakeClientManaged.clusterType, consts.ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-data-will-not-be-used").NewNetwork(), nil, "resource grp should be created")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(fakeClientManaged.state.CloudInfra.Azure.ResourceGroupName) > 0)
		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		_ = fakeClientManaged.ManagedAddons(nil)
		assert.Equal(t, fakeClientManaged.Name("fake-managed").VMType("fake").NewManagedCluster(5), nil, "managed cluster should be created")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Azure.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, fakeClientManaged.state.CloudInfra.Azure.NoManagedNodes, 5)
		assert.Equal(t, *fakeClientManaged.state.Versions.Aks, fakeClientManaged.K8sVersion)
		assert.Assert(t, len(fakeClientManaged.state.CloudInfra.Azure.ManagedClusterName) > 0, "Managed cluster Name not saved")

		_, err := storeManaged.Read()
		if err != nil {
			t.Fatalf("kubeconfig should be present: %v", err)
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []logger.ClusterDataForLogging{
			{
				Name:            fakeClientManaged.ClusterName,
				CloudProvider:   consts.CloudAzure,
				ClusterType:     consts.ClusterTypeMang,
				ResourceGrpName: generateResourceGroupName(fakeClientManaged.ClusterName, string(consts.ClusterTypeMang)),
				Region:          fakeClientManaged.Region,
				ManagedK8sName:  "fake-managed",
				NoMgt:           fakeClientManaged.state.CloudInfra.Azure.NoManagedNodes,
				Mgt:             logger.VMData{VMSize: "fake"},
				K8sDistro:       consts.K8sAks,
				K8sVersion:      *fakeClientManaged.state.Versions.Aks,
				Apps:            nil,
				Cni:             "Name: azure, For: aks, Version: <nil>, KsctlSpecificComponents: map[]",
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(), nil, "managed cluster should be deleted")

		assert.Equal(t, len(fakeClientManaged.state.CloudInfra.Azure.ManagedClusterName), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(), nil, "Network should be deleted")

		assert.Equal(t, len(fakeClientManaged.state.CloudInfra.Azure.ResourceGroupName), 0, "resource grp still present")
		// at this moment the file is not present
		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}
