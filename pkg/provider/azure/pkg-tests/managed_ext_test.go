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

package pkg_tests_test

import (
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/provider/azure"
	"github.com/ksctl/ksctl/pkg/statefile"
	localstate "github.com/ksctl/ksctl/pkg/storage/host"
	"gotest.tools/v3/assert"
)

func TestManagedCluster(t *testing.T) {
	stateDocumentManaged := &statefile.StorageDocument{}
	var (
		clusterName = "demo-managed"
		regionCode  = "fake"
	)

	storeManaged = localstate.NewClient(parentCtx, parentLogger)
	_ = storeManaged.Setup(consts.CloudAzure, regionCode, clusterName, consts.ClusterTypeMang)
	_ = storeManaged.Connect()

	fakeClientManaged, _ = azure.NewClient(parentCtx, parentLogger, controller.Metadata{
		ClusterName: clusterName,
		Region:      regionCode,
		Provider:    consts.CloudAzure,
	}, stateDocumentManaged, storeManaged, azure.ProvideClient)

	fakeClientManaged.ManagedK8sVersion("1.27")
	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, stateDocumentManaged.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-data-will-not-be-used").NewNetwork(), nil, "resource grp should be created")
		assert.Equal(t, stateDocumentManaged.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(stateDocumentManaged.CloudInfra.Azure.ResourceGroupName) > 0)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		assert.Equal(t, fakeClientManaged.Name("fake-managed").VMType("fake").NewManagedCluster(5), nil, "managed cluster should be created")
		assert.Equal(t, stateDocumentManaged.CloudInfra.Azure.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, stateDocumentManaged.CloudInfra.Azure.NoManagedNodes, 5)
		assert.Equal(t, stateDocumentManaged.CloudInfra.Azure.B.KubernetesVer, "1.27")
		assert.Assert(t, len(stateDocumentManaged.CloudInfra.Azure.ManagedClusterName) > 0, "Managed cluster Name not saved")

		_, err := storeManaged.Read()
		if err != nil {
			t.Fatalf("kubeconfig should be present: %v", err)
		}
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []logger.ClusterDataForLogging{
			{
				Name:            clusterName,
				CloudProvider:   consts.CloudAzure,
				ClusterType:     consts.ClusterTypeMang,
				ResourceGrpName: genResourceGroup(clusterName, string(consts.ClusterTypeMang)),
				Region:          regionCode,
				ManagedK8sName:  "fake-managed",
				NoMgt:           stateDocumentManaged.CloudInfra.Azure.NoManagedNodes,
				Mgt:             logger.VMData{VMSize: "fake"},
				K8sDistro:       "managed",
				K8sVersion:      stateDocumentManaged.CloudInfra.Azure.B.KubernetesVer,
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(), nil, "managed cluster should be deleted")

		assert.Equal(t, len(stateDocumentManaged.CloudInfra.Azure.ManagedClusterName), 0, "managed cluster id still present")
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(), nil, "Network should be deleted")

		assert.Equal(t, len(stateDocumentManaged.CloudInfra.Azure.ResourceGroupName), 0, "resource grp still present")
		// at this moment the file is not present
		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}
