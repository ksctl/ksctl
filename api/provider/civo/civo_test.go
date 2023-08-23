package civo

import (
	"errors"
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
	"os"
	"strings"
	"testing"
)

var (
	fakeClient *CivoProvider
	demoClient *resources.KsctlClient
	dir        = fmt.Sprintf("%s/ksctl-test", os.TempDir())
)

func TestMain(m *testing.M) {

	demoClient = &resources.KsctlClient{}

	demoClient.Metadata.ClusterName = "demo"
	demoClient.Metadata.Region = "demoRegion"
	demoClient.Metadata.Provider = "demoProvider"

	demoClient.Cloud, _ = ReturnCivoStruct(demoClient.Metadata, ProvideMockCivoClient)

	fakeClient, _ = ReturnCivoStruct(demoClient.Metadata, ProvideMockCivoClient)

	demoClient.Storage = localstate.InitStorage(false)

	// setup temporary folder
	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	civoHA := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, "ha")
	civoManaged := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, "managed")

	if err := os.MkdirAll(civoManaged, 0755); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(civoHA, 0755); err != nil {
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
		utils.GetPath(utils.CLUSTER_PATH, "civo", "abcd"),
		"genreatePath not compatable with utils.getpath()")
}

func TestIsValidK8sVersion(t *testing.T) {
	ver, _ := fakeClient.Client.ListAvailableKubernetesVersions()
	for _, vver := range ver {
		t.Log(vver)
	}
}

func TestCivoProvider_InitState(t *testing.T) {

	// get the data
	fakeClient.Region = "LON1"

	t.Run("Create state", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, utils.CLUSTER_TYPE_MANG, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeClient.ClusterName+" "+fakeClient.Region, "clusterdir not equal")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeClient.NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		civoCloudState.IsCompleted = true
		assert.Equal(t, civoCloudState.IsCompleted, true, "cluster should not be completed")

		if err := fakeClient.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to resume state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Get request", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, utils.OPERATION_STATE_GET); err != nil {
			t.Fatalf("Unable to get state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Delete request", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, utils.OPERATION_STATE_DELETE); err != nil {
			t.Fatalf("Unable to Delete state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Invalid request", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, "test"); err == nil {
			t.Fatalf("Expected error but not got: %v", err)
		}
	})
}

func TestFetchAPIKey(t *testing.T) {
	environmentTest := [][3]string{
		{"CIVO_TOKEN", "12", "12"},
		{"AZ_TOKEN", "234", ""},
		{"CIVO_TOKEN", "", ""},
	}
	for _, data := range environmentTest {
		if err := os.Setenv(data[0], data[1]); err != nil {
			t.Fatalf("unable to set env vars")
		}
		token := fetchAPIKey(demoClient.Storage)
		if strings.Compare(token, data[2]) != 0 {
			t.Fatalf("missmatch Key: `%s` -> `%s`\texpected `%s` but got `%s`", data[0], data[1], data[2], token)
		}
		if err := os.Unsetenv(data[0]); err != nil {
			t.Fatalf("unable to unset env vars")
		}
	}
}

func TestApplications(t *testing.T) {
	testPreInstalled := map[string]string{
		"":     "Traefik-v2-nodeport,metrics-server",
		"abcd": "abcd,Traefik-v2-nodeport,metrics-server",
	}

	for apps, setVal := range testPreInstalled {
		if retApps := fakeClient.Application(apps); retApps == nil {
			t.Fatalf("application returned nil for valid applications as input")
		} else {
			if fakeClient.Metadata.Apps != setVal {
				t.Fatalf("apps dont match `%s` Expected `%s` but got `%s`", apps, setVal, retApps)
			}
		}
	}
}

// Test for the Noof WP and setter and getter
func TestCivoProvider_NoOfControlPlane(t *testing.T) {
	var no int
	var err error
	no, err = demoClient.Cloud.NoOfControlPlane(-1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized controlplanes array got no: %d and err: %v", no, err)
	}

	_, err = demoClient.Cloud.NoOfControlPlane(1, true)
	// it should return error
	if err == nil {
		t.Fatalf("setter should fail on when no < 3 controlplanes provided_no: %d", 1)
	}

	_, err = demoClient.Cloud.NoOfControlPlane(5, true)
	// it should return error
	if err != nil {
		t.Fatalf("setter should not fail on when n >= 3 controlplanes err: %v", err)
	}

	no, err = demoClient.Cloud.NoOfControlPlane(-1, false)
	if no != 5 {
		t.Fatalf("Getter failed to get updated no of controlplanes array got no: %d and err: %v", no, err)
	}
}

func TestCivoProvider_NoOfDataStore(t *testing.T) {
	var no int
	var err error
	no, err = demoClient.Cloud.NoOfDataStore(-1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized datastore array got no: %d and err: %v", no, err)
	}

	_, err = demoClient.Cloud.NoOfDataStore(0, true)
	// it should return error
	if err == nil {
		t.Fatalf("setter should fail on when no < 1 datastore provided_no: %d", 1)
	}

	_, err = demoClient.Cloud.NoOfDataStore(5, true)
	// it should return error
	if err != nil {
		t.Fatalf("setter should not fail on when n >= 1 datastore err: %v", err)
	}

	no, err = demoClient.Cloud.NoOfDataStore(-1, false)
	if no != 5 {
		t.Fatalf("Getter failed to get updated no of datastore array got no: %d and err: %v", no, err)
	}
}

func TestCivoProvider_NoOfWorkerPlane(t *testing.T) {
	var no int
	var err error
	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized workerplane array got no: %d and err: %v", no, err)
	}

	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 2, true)
	// it shouldn't return err
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("setter should not fail on when no >= 0 workerplane provided_no: %d", 2)
	}

	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 2, true)
	if err != nil {
		t.Fatalf("setter should return nil when no changes happen workerplane err: %v", err)
	}

	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 3, true)
	if err != nil {
		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
	}

	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 1, true)
	if err != nil {
		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
	}

	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
	if no != 1 {
		t.Fatalf("Getter failed to get updated no of workerplane array got no: %d and err: %v", no, err)
	}
}

func TestResName(t *testing.T) {

	if ret := fakeClient.Name("demo"); ret == nil {
		t.Fatalf("returned nil for valid res name")
	}
	if fakeClient.Metadata.ResName != "demo" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClient.Name("12demo"); ret != nil {
		t.Fatalf("returned interface for invalid res name")
	}
}

func TestRole(t *testing.T) {
	validSet := []string{utils.ROLE_CP, utils.ROLE_LB, utils.ROLE_DS, utils.ROLE_WP}
	for _, val := range validSet {
		if ret := fakeClient.Role(val); ret == nil {
			t.Fatalf("returned nil for valid role")
		}
		if fakeClient.Metadata.Role != val {
			t.Fatalf("Correct assignment missing")
		}
	}
	if ret := fakeClient.Role("fake"); ret != nil {
		t.Fatalf("returned interface for invalid role")
	}
}

func TestVMType(t *testing.T) {
	if ret := fakeClient.VMType("g4s.kube.small"); ret == nil {
		t.Fatalf("returned nil for valid vm type")
	}
	if fakeClient.Metadata.VmType != "g4s.kube.small" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClient.VMType(""); ret != nil {
		t.Fatalf("returned interface for invalid vm type")
	}
}

func TestVisibility(t *testing.T) {
	if fakeClient.Visibility(true); !fakeClient.Metadata.Public {
		t.Fatalf("Visibility setting not working")
	}
}

// Mock the return of ValidListOfRegions
func TestRegion(t *testing.T) {

	forTesting := map[string]error{
		"Lon!": errors.New(""),
		"":     errors.New(""),
		"NYC1": nil,
	}

	for key, val := range forTesting {
		if err := isValidRegion(fakeClient, key); (err == nil && val != nil) || (err != nil && val == nil) {
			t.Fatalf("Input region :`%s`. expected `%v` but got `%v`", key, val, err)
		}
	}
}

func TestK8sVersion(t *testing.T) {
	// these are invalid
	// input and output
	forTesting := []string{
		"1.27.4",
		"1.27.1",
		"1.28",
	}

	for i := 0; i < len(forTesting); i++ {
		var ver string = forTesting[i]
		if i < 2 {
			if ret := fakeClient.Version(ver); ret == nil {
				t.Fatalf("returned nil for valid version")
			}
			if ver+"-k3s1" != fakeClient.Metadata.K8sVersion {
				t.Fatalf("set value is not equal to input value")
			}
		} else {
			if ret := fakeClient.Version(ver); ret != nil {
				t.Fatalf("returned interface for invalid version")
			}
		}
	}

	if ret := fakeClient.Version(""); ret == nil {
		t.Fatalf("returned nil for valid version")
	}
	if "1.26.4-k3s1" != fakeClient.Metadata.K8sVersion {
		t.Fatalf("set value is not equal to input value")
	}
}

func TestCniAndOthers(t *testing.T) {
	t.Run("CNI Support flag", func(t *testing.T) {
		if !fakeClient.SupportForCNI() {
			t.Fatal("Support for CNI must be true")
		}
	})

	t.Run("Application support flag", func(t *testing.T) {
		if !fakeClient.SupportForApplications() {
			t.Fatal("Support for Application must be true")
		}
	})

	t.Run("CNI set functionality", func(t *testing.T) {
		if ret := fakeClient.CNI("cilium"); ret == nil {
			t.Fatalf("returned nil for valid CNI")
		}
		if ret := fakeClient.CNI(""); ret == nil {
			t.Fatalf("returned nil for valid CNI")
		}

		if ret := fakeClient.CNI("abcd"); ret != nil {
			t.Fatalf("returned interface for invalid CNI")
		}
	})
}

func TestFirewallRules(t *testing.T) {
	t.Run("Controlplane fw rules", func(t *testing.T) {
		if firewallRuleControlPlane() != nil {
			t.Fatalf("missmatch firewall rule")
		}
	})

	t.Run("Workerplane fw rules", func(t *testing.T) {
		if firewallRuleWorkerPlane() != nil {
			t.Fatalf("missmatch firewall rule")
		}
	})

	t.Run("Loadbalancer fw rules", func(t *testing.T) {
		if firewallRuleLoadBalancer() != nil {
			t.Fatalf("missmatch firewall rule")
		}
	})

	t.Run("Datastore fw rules", func(t *testing.T) {
		if firewallRuleDataStore() != nil {
			t.Fatalf("missmatch firewall rule")
		}
	})
}
