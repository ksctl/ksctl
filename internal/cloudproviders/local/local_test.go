package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/pkg/logger"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"gotest.tools/v3/assert"
)

var (
	fakeClientManaged *LocalProvider
	storeManaged      types.StorageFactory

	fakeClientVars *LocalProvider
	//storeVars      storage.StorageFactory
	parentCtx    context.Context     = context.TODO()
	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)

	dir = fmt.Sprintf("%s ksctl-local-test", os.TempDir())
)

func TestMain(m *testing.M) {
	func() {

		fakeClientVars, _ = NewClient(parentCtx, types.Metadata{
			ClusterName: "demo",
			Region:      "LOCAL",
			Provider:    consts.CloudLocal,
		}, parentLogger, &storageTypes.StorageDocument{}, ProvideMockClient)

	}()

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)

	exitVal := m.Run()
	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-local-test"); err != nil {
		panic(err)
	}
	os.Exit(exitVal)
}

func TestRole(t *testing.T) {
	if factory := fakeClientVars.Role(""); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// it will contain which vmType to create
func TestVMType(t *testing.T) {
	if factory := fakeClientVars.VMType(""); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// whether to have the resource as public or private (i.e. VMs)
func TestVisibility(t *testing.T) {
	if factory := fakeClientVars.Visibility(false); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

func TestGetHostNameAllWorkerNode(t *testing.T) {
	if factory := fakeClientVars.GetHostNameAllWorkerNode(); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// CreateUploadSSHKeyPair implements types.CloudFactory.
func TestCreateUploadSSHKeyPair(t *testing.T) {
	if factory := fakeClientVars.CreateUploadSSHKeyPair(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelFirewall implements types.CloudFactory.
func TestDelFirewall(t *testing.T) {
	if factory := fakeClientVars.DelFirewall(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelNetwork implements types.CloudFactory.
func TestDelNetwork(t *testing.T) {
	if factory := fakeClientVars.DelNetwork(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelSSHKeyPair implements types.CloudFactory.
func TestDelSSHKeyPair(t *testing.T) {
	if factory := fakeClientVars.DelSSHKeyPair(nil); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// DelVM implements types.CloudFactory.
func TestDelVM(t *testing.T) {
	if factory := fakeClientVars.DelVM(nil, 0); factory != nil {
		t.Fatalf("it should not be implemented")
	}
}

// GetStateForHACluster implements types.CloudFactory.
func TestGetStateForHACluster(t *testing.T) {
	if _, err := fakeClientVars.GetStateForHACluster(nil); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

// NewFirewall implements types.CloudFactory.
func TestNewFirewall(t *testing.T) {
	if err := fakeClientVars.NewFirewall(nil); err != nil {
		t.Fatalf("it should not be implemented")
	}
}

// NewNetwork implements types.CloudFactory.
func TestNewNetwork(t *testing.T) {
	if err := fakeClientVars.NewNetwork(nil); err != nil {
		t.Fatalf("it should not be implemented")
	}
}

// NewVM implements types.CloudFactory.
func TestNewVM(t *testing.T) {
	if err := fakeClientVars.NewVM(nil, 0); err != nil {
		t.Fatalf("it should not be implemented")
	}
}

// NoOfControlPlane implements types.CloudFactory.
func TestNoOfControlPlane(t *testing.T) {
	if _, err := fakeClientVars.NoOfControlPlane(-1, false); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

// NoOfDataStore implements types.CloudFactory.
func TestNoOfDataStore(t *testing.T) {
	if _, err := fakeClientVars.NoOfDataStore(-1, false); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

// NoOfWorkerPlane implements types.CloudFactory.
func TestNoOfWorkerPlane(t *testing.T) {
	if _, err := fakeClientVars.NoOfWorkerPlane(nil, 0, false); err == nil {
		t.Fatalf("it should not be implemented")
	}
}

func TestCNIandApp(t *testing.T) {

	testCases := map[string]bool{
		string(consts.CNIKind):    false,
		string(consts.CNIKubenet): true,
		string(consts.CNICilium):  true,
	}

	for k, v := range testCases {
		got := fakeClientVars.CNI(k)
		assert.Equal(t, got, v, "missmatch")
	}

	got := fakeClientVars.Application([]string{"abcd"})
	if !got {
		t.Fatalf("application should be external")
	}
}

func TestGenerateConfig(t *testing.T) {
	if _, err := generateConfig(0, 0, false); err == nil {
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
		if raw, _ := generateConfig(noWP, noCP, false); !reflect.DeepEqual(raw, []byte(val)) {
			t.Fatalf("Data missmatch for noCP: %d, noWP: %d expected %s but got %s", noCP, noWP, val, string(raw))
		}
	}
}

func TestManagedCluster(t *testing.T) {
	mainStateDocument = &storageTypes.StorageDocument{}
	func() {
		fakeClientManaged, _ = NewClient(parentCtx, types.Metadata{
			ClusterName: "demo-managed",
			Region:      "LOCAL",
			Provider:    consts.CloudLocal,
		}, parentLogger, &storageTypes.StorageDocument{}, ProvideMockClient)

		storeManaged = localstate.NewClient(parentCtx, parentLogger)
		_ = storeManaged.Setup(consts.CloudLocal, "LOCAL", "demo-managed", consts.ClusterTypeMang)
		_ = storeManaged.Connect(context.TODO())

	}()

	assert.Equal(t, fakeClientManaged.InitState(storeManaged, consts.OperationCreate), nil, "Init must work before")
	fakeClientManaged.Version("1.27.1")
	fakeClientManaged.Name("fake")
	assert.Equal(t, fakeClientManaged.NewManagedCluster(storeManaged, 2), nil, "managed cluster should be created")
	assert.Equal(t, mainStateDocument.CloudInfra.Local.Nodes, 2, "missmatch of no of nodes")
	assert.Equal(t, mainStateDocument.CloudInfra.Local.B.KubernetesVer, fakeClientManaged.Metadata.Version, "k8s version does not match")
	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientManaged.GetStateFile(storeManaged)
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(mainStateDocument)
		assert.DeepEqual(t, string(got), expected)
	})

	assert.Equal(t, fakeClientManaged.DelManagedCluster(storeManaged), nil, "managed cluster should be deleted")
}
