package aws

import (
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"gotest.tools/v3/assert"
	"testing"
)

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(consts.CloudAws, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeMang); err != nil {
		t.Fatal(err)
	}
	read, err := storeManaged.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, mainStateDocument, read)
}

func TestManagedCluster(t *testing.T) {
	mainStateDocumentManaged := &storageTypes.StorageDocument{}
	fakeClientManaged, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo-managed",
		Region:      "fake-region",
		Provider:    consts.CloudAws,
	}, parentLogger, mainStateDocumentManaged, ProvideClient)

	storeManaged = localstate.NewClient(parentCtx, parentLogger)
	_ = storeManaged.Setup(consts.CloudAws, "fake-region", "demo-managed", consts.ClusterTypeMang)
	_ = storeManaged.Connect()

	fakeClientManaged.ManagedK8sVersion("1.30")
	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(storeManaged, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-data-will-not-be-used").NewNetwork(storeManaged), nil, "resource grp should be created")
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocumentManaged.CloudInfra.Aws.VpcId) > 0)
		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		assert.Equal(t, fakeClientManaged.Name("fake-managed").VMType("fake").NewManagedCluster(storeManaged, 5), nil, "managed cluster should be created")
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Aws.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Aws.NoManagedNodes, 5)
		assert.Equal(t, mainStateDocumentManaged.CloudInfra.Aws.B.KubernetesVer, fakeClientManaged.metadata.k8sVersion)
		assert.Assert(t, len(mainStateDocumentManaged.CloudInfra.Aws.ManagedClusterName) > 0, "Managed cluster Name not saved")

		_, err := storeManaged.Read()
		if err != nil {
			t.Fatalf("kubeconfig should be present: %v", err)
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:          fakeClientManaged.clusterName,
				CloudProvider: consts.CloudAws,
				ClusterType:   consts.ClusterTypeMang,
				NetworkName:   "demo-managed-vpc",
				NetworkID:     "3456d25f36g474g546",
				LB: cloud.VMData{
					SubnetID:   "3456d25f36g474g546",
					SubnetName: "demo-managed-subnet0",
				},
				Region:     fakeClientManaged.region,
				NoMgt:      mainStateDocumentManaged.CloudInfra.Aws.NoManagedNodes,
				Mgt:        cloud.VMData{VMSize: "fake"},
				K8sDistro:  "managed",
				K8sVersion: mainStateDocumentManaged.CloudInfra.Aws.B.KubernetesVer,
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos(storeManaged)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(storeManaged), nil, "managed cluster should be deleted")

		assert.Equal(t, len(mainStateDocumentManaged.CloudInfra.Aws.ManagedClusterName), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(storeManaged), nil, "Network should be deleted")

		assert.Equal(t, len(mainStateDocumentManaged.CloudInfra.Aws.VpcId), 0, "resource grp still present")
		_, err := storeManaged.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}
