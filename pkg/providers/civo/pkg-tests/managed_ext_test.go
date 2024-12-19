package pkg_tests_test

import (
	"github.com/ksctl/ksctl/internal/cloudproviders/civo"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"gotest.tools/v3/assert"
	"testing"
)

func TestManagedCluster(t *testing.T) {
	mainStateDocumentManaged := &storageTypes.StorageDocument{}
	var (
		clusterName = "demo-managed"
		regionCode  = "LON1"
	)

	fakeClientManaged, _ = civo.NewClient(parentCtx, types.Metadata{
		ClusterName: clusterName,
		Region:      regionCode,
		Provider:    consts.CloudCivo,
	}, parentLogger, mainStateDocumentManaged, civo.ProvideClient)

	storeManaged = localstate.NewClient(parentCtx, parentLogger)
	_ = storeManaged.Setup(consts.CloudCivo, "LON1", "demo-managed", consts.ClusterTypeMang)
	_ = storeManaged.Connect()

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(storeManaged, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-net").NewNetwork(storeManaged), nil, "Network should be created")
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocumentManaged.CloudInfra.Civo.NetworkID) > 0, "network id not saved")

	})

	t.Run("Create managed cluster", func(t *testing.T) {

		fakeClientManaged.CNI("cilium")
		fakeClientManaged.Application([]string{"abcd"})
		assert.Assert(t, fakeClientManaged.ManagedK8sVersion("1.27.1") != nil)

		assert.Equal(t, fakeClientManaged.Name("fake").VMType("g4s.kube.small").NewManagedCluster(storeManaged, 5), nil, "managed cluster should be created")

		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.NoManagedNodes, 5)
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Civo.B.KubernetesVer, "1.27.1-k3s1")
		assert.Assert(t, len(mainStateDocumentManaged.CloudInfra.Civo.ManagedClusterID) > 0, "Managed clusterID not saved")

		_, err := storeManaged.Read()
		if err != nil {
			t.Fatalf("kubeconfig should not be absent")
		}
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			{
				Name:          clusterName,
				CloudProvider: consts.CloudCivo,
				ClusterType:   consts.ClusterTypeMang,
				NetworkID:     "fake-net",
				ManagedK8sID:  "fake-k8s",
				Region:        regionCode,
				NoMgt:         mainStateDocumentManaged.CloudInfra.Civo.NoManagedNodes,
				Mgt:           cloud.VMData{VMSize: "g4s.kube.small"},

				K8sDistro:  "managed",
				K8sVersion: mainStateDocumentManaged.CloudInfra.Civo.B.KubernetesVer,
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos(storeManaged)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(storeManaged), nil, "managed cluster should be deleted")

		assert.Equal(t, len(mainStateDocumentManaged.CloudInfra.Civo.ManagedClusterID), 0, "managed cluster id still present")

	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(storeManaged), nil, "Network should be deleted")

		assert.Equal(t, len(mainStateDocumentManaged.CloudInfra.Civo.NetworkID), 0, "network id still present")
		// at this moment the file is not present
		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory still present")
		}
	})

}
