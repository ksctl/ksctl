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

package civo

import (
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	localstate "github.com/ksctl/ksctl/pkg/storage/host"
	"gotest.tools/v3/assert"
	"testing"
)

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(consts.CloudCivo, fakeClientManaged.state.Region, fakeClientManaged.state.ClusterName, consts.ClusterTypeMang); err != nil {
		t.Fatal(err)
	}
	read, err := storeManaged.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, fakeClientManaged.state, read)
}

func TestManagedCluster(t *testing.T) {
	mainStateDocumentManaged := &statefile.StorageDocument{}
	storeManaged = localstate.NewClient(parentCtx, parentLogger)

	fakeClientManaged, _ = NewClient(parentCtx, parentLogger, controller.Metadata{
		ClusterName: "demo-managed",
		Region:      "LON1",
		Provider:    consts.CloudCivo,
	}, mainStateDocumentManaged, storeManaged, ProvideClient)

	_ = storeManaged.Setup(consts.CloudCivo, "LON1", "demo-managed", consts.ClusterTypeMang)
	_ = storeManaged.Connect()

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, fakeClientManaged.clusterType, consts.ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-net").NewNetwork(), nil, "Network should be created")
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocumentManaged.CloudInfra.Civo.NetworkID) > 0, "network id not saved")

		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		fakeClientManaged.CNI("cilium")
		fakeClientManaged.Application([]string{"abcd"})

		assert.Equal(t, fakeClientManaged.Name("fake").VMType("g4s.kube.small").NewManagedCluster(5), nil, "managed cluster should be created")

		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.NoManagedNodes, 5)
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.KubernetesVer, fakeClientManaged.K8sVersion)
		assert.Assert(t, len(mainStateDocumentManaged.CloudInfra.Civo.ManagedClusterID) > 0, "Managed clusterID not saved")

		_, err := storeManaged.Read()
		if err != nil {
			t.Fatalf("kubeconfig should not be absent")
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []logger.ClusterDataForLogging{
			{
				Name:          fakeClientManaged.ClusterName,
				CloudProvider: consts.CloudCivo,
				ClusterType:   consts.ClusterTypeMang,
				NetworkID:     "fake-net",
				ManagedK8sID:  "fake-k8s",
				Region:        fakeClientManaged.Region,
				NoMgt:         mainStateDocumentManaged.CloudInfra.Civo.NoManagedNodes,
				Mgt:           logger.VMData{VMSize: "g4s.kube.small"},

				K8sDistro:  "managed",
				K8sVersion: mainStateDocumentManaged.CloudInfra.Civo.B.KubernetesVer,
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(), nil, "managed cluster should be deleted")

		assert.Equal(t, len(mainStateDocumentManaged.CloudInfra.Civo.ManagedClusterID), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(), nil, "Network should be deleted")

		assert.Equal(t, len(mainStateDocumentManaged.CloudInfra.Civo.NetworkID), 0, "network id still present")
		// at this moment the file is not present
		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory still present")
		}
	})

}
