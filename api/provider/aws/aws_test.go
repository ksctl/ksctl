package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
)

var (
	demoClient *resources.KsctlClient
	fakeaws    *AwsProvider
	dir        = fmt.Sprintf("%s/ksctl-aws-test", os.TempDir())
)

func TestMain(m *testing.M) {
	demoClient = &resources.KsctlClient{}
	demoClient.Metadata.ClusterName = "fake-cluster"
	demoClient.Metadata.Region = "fake"
	demoClient.Metadata.Provider = "aws"
	demoClient.Cloud, _ = ReturnAwsStruct(demoClient.Metadata, ProvideMockClient)

	fakeaws, _ = ReturnAwsStruct(demoClient.Metadata, ProvideMockClient)

	demoClient.Storage = localstate.InitStorage(false)
	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)

	// awsHA := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AWS, utils.CLUSTER_TYPE_HA)

	// if err := os.MkdirAll(awsHA, 0755); err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Created tmp directories")
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestHACluster(t *testing.T) {
	fakeaws.region = "fake"
	fakeaws.clusterName = "fake"

	fakeaws.metadata.noCP = 7
	fakeaws.metadata.noDS = 5
	fakeaws.metadata.noWP = 10
	fakeaws.metadata.public = true
	fakeaws.metadata.k8sName = utils.K8S_K3S

	t.Run("init state", func(t *testing.T) {

		if err := fakeaws.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, utils.CLUSTER_TYPE_HA, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeaws.clusterName+" "+fakeaws.vpc+" "+fakeaws.region, "clusterdir not equal")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")

		_, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeaws.Name("fake-data-not-used").NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")

		checkCurrentStateFileHA(t)
	})
}

func checkCurrentStateFileHA(t *testing.T) {

	raw, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_HA, clusterDirName, STATE_FILE_NAME)).Load()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("Reason: %v", err)
	}

	assert.DeepEqual(t, azureCloudState, data)
}
