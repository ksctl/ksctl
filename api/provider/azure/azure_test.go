package azure

import (
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
	"os"
	"testing"
)

// TODO: seperate the api calls so that we can add mocks

var (
	demoClient *resources.KsctlClient
)

func TestMain(m *testing.M) {
	var err error
	if err != nil {
		panic("unable to start fake client")
	}
	demoClient = &resources.KsctlClient{}
	azureCloudState = &StateConfiguration{}
	demoClient.Cloud, _ = ReturnAzureStruct(demoClient.Metadata)

	demoClient.ClusterName = "demo"
	demoClient.Region = "demoRegion"
	demoClient.Provider = "demoProvider"
	demoClient.Storage = localstate.InitStorage(false)

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestInit(t *testing.T) {
	// NOTE: Refer for fake client https://gist.github.com/jhendrixMSFT/efd5bcc94f2d545b5cb3a4f5e66446f4
}

func TestConsts(t *testing.T) {
	assert.Equal(t, KUBECONFIG_FILE_NAME, "kubeconfig", "kubeconfig file")
	assert.Equal(t, STATE_FILE_NAME, "cloud-state.json", "cloud state file")

	assert.Equal(t, FILE_PERM_CLUSTER_STATE, os.FileMode(0640), "state file permission mismatch")
	assert.Equal(t, FILE_PERM_CLUSTER_DIR, os.FileMode(0750), "cluster dir permission mismatch")
	assert.Equal(t, FILE_PERM_CLUSTER_KUBECONFIG, os.FileMode(0755), "kubeconfig file permission mismatch")
}

func TestGenPath(t *testing.T) {
	assert.Equal(t,
		generatePath(utils.CLUSTER_PATH, "abcd"),
		utils.GetPath(utils.CLUSTER_PATH, "azure", "abcd"),
		"genreatePath not compatable with utils.getpath()")
}
