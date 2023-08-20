package civo

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
)

var (
	fakeClient *CivoProvider
	demoClient *resources.KsctlClient
)

func TestMain(m *testing.M) {

	demoClient = &resources.KsctlClient{}
	civoCloudState = &StateConfiguration{}
	demoClient.Cloud, _ = ReturnCivoStruct(demoClient.Metadata, ProvideClient)

	fakeClient, _ = ReturnCivoStruct(demoClient.Metadata, func() CivoGo {
		return &CivoGoMockClient{}
	})

	demoClient.ClusterName = "demo"
	demoClient.Region = "demoRegion"
	demoClient.Provider = "demoProvider"
	demoClient.Storage = localstate.InitStorage(false)

	exitVal := m.Run()

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

func TestConvertStateToBytes(t *testing.T) {
	civoCloudState.ClusterName = "demo"
	bytes, err := convertStateToBytes(*civoCloudState)
	if err != nil {
		t.Fatal("missmatch in conversion of state to bytes")
	}
	a, err := json.Marshal(civoCloudState)
	assert.DeepEqual(t, bytes, a, nil)
}

func TestCivoProvider_InitState(t *testing.T) {
	//TODO: add
}

func TestFetchAPIKey(t *testing.T) {
	t.Logf("try checking for env")

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
		if retApps := aggregratedApps(apps); strings.Compare(retApps, setVal) != 0 {
			t.Fatalf("apps dont match `%s` Expected `%s` but got `%s`", apps, setVal, retApps)
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

	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
	if no != 2 {
		t.Fatalf("Getter failed to get updated no of workerplane array got no: %d and err: %v", no, err)
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
	forTesting := map[string]error{
		"":            errors.New(""),
		"1.28":        errors.New(""),
		"1.27.4-k3s1": nil,
		"1.27-k3s1":   errors.New(""),
		"1.27.1-k3s1": nil,
	}

	for ver, Rver := range forTesting {
		if err := isValidK8sVersion(fakeClient, ver); (err == nil && Rver != nil) || (err != nil && Rver == nil) {
			t.Fatalf("version dont match we have `%s` Expected `%s` but got `%s`", ver, Rver, err)
		}
	}
}

func TestVMType(t *testing.T) {

	// input and output
	forTesting := map[string]error{
		"g3.dca":         errors.New(""),
		"":               errors.New(""),
		"dca":            errors.New(""),
		"g4s.kube.small": nil,
	}
	for key, val := range forTesting {
		if err := isValidVMSize(fakeClient, key); (err != nil && val == nil) || (err == nil && val != nil) {
			t.Fatalf("VM type: `%s` Expected `%v` got `%v`", key, val, err)
		}
	}
}
