package local

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
)

var (
	demoClient *resources.KsctlClient
	testClient *LocalProvider
	dir        = fmt.Sprintf("%s/ksctl-local-test", os.TempDir())
)

func TestMain(m *testing.M) {
	demoClient = &resources.KsctlClient{}
	demoClient.Metadata.ClusterName = "demo"
	demoClient.Metadata.Region = "demoRegion"
	demoClient.Metadata.Provider = "demoProvider"
	demoClient.Storage = localstate.InitStorage(false)
	localState = &StateConfiguration{}
	demoClient.Cloud, _ = ReturnLocalStruct(demoClient.Metadata)

	testClient, _ = ReturnLocalStruct(demoClient.Metadata)

	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	civoManaged := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, "managed")

	if err := os.MkdirAll(civoManaged, 0755); err != nil {
		panic(err)
	}

	fmt.Println("Created tmp directories")

	exitVal := m.Run()
	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
	os.Exit(exitVal)
}

func TestRole(t *testing.T) {
	if factory := demoClient.Cloud.Role(""); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// it will contain which vmType to create
func TestVMType(t *testing.T) {
	if factory := demoClient.Cloud.VMType(""); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// whether to have the resource as public or private (i.e. VMs)
func TestVisibility(t *testing.T) {
	if factory := demoClient.Cloud.Visibility(false); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

func TestGetHostNameAllWorkerNode(t *testing.T) {
	if factory := demoClient.Cloud.GetHostNameAllWorkerNode(); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func TestCreateUploadSSHKeyPair(t *testing.T) {
	if factory := demoClient.Cloud.CreateUploadSSHKeyPair(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelFirewall implements resources.CloudFactory.
func TestDelFirewall(t *testing.T) {
	if factory := demoClient.Cloud.DelFirewall(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelNetwork implements resources.CloudFactory.
func TestDelNetwork(t *testing.T) {
	if factory := demoClient.Cloud.DelNetwork(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelSSHKeyPair implements resources.CloudFactory.
func TestDelSSHKeyPair(t *testing.T) {
	if factory := demoClient.Cloud.DelSSHKeyPair(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelVM implements resources.CloudFactory.
func TestDelVM(t *testing.T) {
	if factory := demoClient.Cloud.DelVM(nil, 0); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// GetManagedKubernetes implements resources.CloudFactory.
func TestGetManagedKubernetes(t *testing.T) {
	demoClient.Cloud.GetManagedKubernetes(nil)

}

// GetStateForHACluster implements resources.CloudFactory.
func TestGetStateForHACluster(t *testing.T) {
	if _, err := demoClient.Cloud.GetStateForHACluster(nil); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

// NewFirewall implements resources.CloudFactory.
func TestNewFirewall(t *testing.T) {
	if err := demoClient.Cloud.NewFirewall(nil); err != nil {
		t.Fatalf("it should not be implemented")
	}
}

// NewNetwork implements resources.CloudFactory.
func TestNewNetwork(t *testing.T) {
	if err := demoClient.Cloud.NewNetwork(nil); err != nil {
		t.Fatalf("it should not be implemented")
	}
}

// NewVM implements resources.CloudFactory.
func TestNewVM(t *testing.T) {
	if err := demoClient.Cloud.NewVM(nil, 0); err != nil {
		t.Fatalf("it should not be implemented")
	}
}

// NoOfControlPlane implements resources.CloudFactory.
func TestNoOfControlPlane(t *testing.T) {
	if _, err := demoClient.Cloud.NoOfControlPlane(-1, false); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

// NoOfDataStore implements resources.CloudFactory.
func TestNoOfDataStore(t *testing.T) {
	if _, err := demoClient.Cloud.NoOfDataStore(-1, false); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

// NoOfWorkerPlane implements resources.CloudFactory.
func TestNoOfWorkerPlane(t *testing.T) {
	if _, err := demoClient.Cloud.NoOfWorkerPlane(nil, 0, false); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

func TestGenerateConfig(t *testing.T) {
	if _, err := generateConfig(0, 0); err == nil {
		t.Fatalf("It should throw err as no of controlplane is 0")
	}

	valid := map[string]string{
		strconv.Itoa(1) + " " + strconv.Itoa(1): `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
...`,
		strconv.Itoa(0) + " " + strconv.Itoa(1): `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
...`,
	}
	for key, val := range valid {
		inp := strings.Split(key, " ")
		noWP, _ := strconv.Atoi(inp[0])
		noCP, _ := strconv.Atoi(inp[1])
		if raw, _ := generateConfig(noWP, noCP); !reflect.DeepEqual(raw, []byte(val)) {
			t.Fatalf("Data missmatch for noCP: %d, noWP: %d expected %s but got %s", noCP, noWP, val, string(raw))
		}
	}
}

func TestManagedCluster(t *testing.T) {

	if runtime.GOOS == "linux" {
		testClient.Version("1.27.1")
		testClient.Name("fake")
		assert.Equal(t, testClient.NewManagedCluster(demoClient.Storage, 2), nil, "managed cluster should be created")
		assert.Equal(t, localState.Nodes, 2, "missmatch of no of nodes")
		assert.Equal(t, localState.Version, testClient.Metadata.Version, "k8s version does not match")

		assert.Equal(t, testClient.DelManagedCluster(demoClient.Storage), nil, "managed cluster should be deleted")
	}
}
