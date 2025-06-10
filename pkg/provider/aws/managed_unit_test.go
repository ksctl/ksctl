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

package aws

import (
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/provider"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	localstate "github.com/ksctl/ksctl/v2/pkg/storage/host"
	"gotest.tools/v3/assert"
)

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(consts.CloudAws, fakeClientManaged.state.Region, fakeClientManaged.state.ClusterName, consts.ClusterTypeMang); err != nil {
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
	_ = storeManaged.Setup(consts.CloudAws, "fake-region", "demo-managed", consts.ClusterTypeMang)

	fakeClientManaged, _ = NewClient(
		parentCtx,
		parentLogger,
		ksc,
		controller.Metadata{
			ClusterName: "demo-managed",
			Region:      "fake-region",
			ClusterType: consts.ClusterTypeMang,
			Provider:    consts.CloudAws,
		},
		&statefile.StorageDocument{},
		storeManaged,
		ProvideClient,
	)

	fakeClientManaged.ManagedK8sVersion("1.30")
	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, fakeClientManaged.ClusterType, consts.ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-data-will-not-be-used").NewNetwork(), nil, "resource grp should be created")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(fakeClientManaged.state.CloudInfra.Aws.VpcId) > 0)
		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {
		_ = fakeClientManaged.ManagedAddons(nil)

		assert.Equal(t, fakeClientManaged.Name("fake-managed").VMType("fake").NewManagedCluster(5), nil, "managed cluster should be created")
		assert.Equal(t, fakeClientManaged.state.CloudInfra.Aws.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, fakeClientManaged.state.CloudInfra.Aws.NoManagedNodes, 5)
		assert.Equal(t, *fakeClientManaged.state.Versions.Eks, fakeClientManaged.K8sVersion)
		assert.Assert(t, len(fakeClientManaged.state.CloudInfra.Aws.ManagedClusterName) > 0, "Managed cluster Name not saved")

		_, err := storeManaged.Read()
		if err != nil {
			t.Fatalf("kubeconfig should be present: %v", err)
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []provider.ClusterData{
			{
				Name:          fakeClientManaged.ClusterName,
				CloudProvider: consts.CloudAws,
				ClusterType:   consts.ClusterTypeMang,
				Team:          "47f9a67b-2499-4e96-9576-ddc703d839f0",
				Owner:         "dipankar.das@ksctl.com",
				State:         statefile.Creating, // As the controller is not here where it actually sets the state so it is creating
				NetworkName:   "demo-managed-vpc",
				NetworkID:     "3456d25f36g474g546",
				LB: provider.VMData{
					SubnetID:   "3456d25f36g474g546",
					SubnetName: "demo-managed-subnet0",
				},
				Region:     fakeClientManaged.Region,
				NoMgt:      fakeClientManaged.state.CloudInfra.Aws.NoManagedNodes,
				Mgt:        provider.VMData{VMSize: "fake"},
				K8sDistro:  consts.K8sEks,
				K8sVersion: *fakeClientManaged.state.Versions.Eks,
				Apps:       []string{"Name: eks-node-monitoring-agent, For: eks, Version: <nil>, KsctlSpecificComponents: map[]"},
				Cni:        "Name: aws, For: eks, Version: <nil>, KsctlSpecificComponents: map[]",
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(), nil, "managed cluster should be deleted")

		assert.Equal(t, len(fakeClientManaged.state.CloudInfra.Aws.ManagedClusterName), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(), nil, "Network should be deleted")

		assert.Equal(t, len(fakeClientManaged.state.CloudInfra.Aws.VpcId), 0, "resource grp still present")
		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}
