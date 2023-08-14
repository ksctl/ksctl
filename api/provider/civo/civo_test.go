package civo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
)

var (
	fakeClient *civogo.FakeClient
	demoClient *resources.KsctlClient
)

// TODO: seperate the api calls so that we can add mocks

func TestMain(m *testing.M) {
	var err error
	fakeClient, err = civogo.NewFakeClient()
	if err != nil {
		panic("unable to start fake client")
	}
	demoClient = &resources.KsctlClient{}
	civoCloudState = &StateConfiguration{}
	demoClient.Cloud, _ = ReturnCivoStruct(demoClient.Metadata)

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
	ver, _ := fakeClient.ListAvailableKubernetesVersions()
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

func TestCivoProvider_VMType(t *testing.T) {
	VMType := make(map[string]error)
	var availableVMTypes []string

	VMType["g4s.kube.small"] = nil // making it valid

	fakeInstances, _ := fakeClient.ListInstanceSizes()
	for _, r := range fakeInstances {
		VMType[r.ID] = nil
	}
	for k, _ := range VMType {
		availableVMTypes = append(availableVMTypes, k)
	}

	// input and output
	forTesting := map[string]error{
		"g3.dca":         errors.New("invalid"),
		"":               errors.New("invalid"),
		"dca":            errors.New("invalid"),
		"g4s.kube.small": nil,
	}
	for key, val := range forTesting {
		if err := isValidRegion(availableVMTypes, key); (err != nil && val == nil) || (err == nil && val != nil) {
			t.Fatalf("VM ID: `%s` Expected `%v` got `%v`", key, val, err)
		}
	}
}

// Mock the return of ValidListOfRegions
func TestIsValidRegion(t *testing.T) {
	valRegions := make(map[string]error)
	var availableRegions []string

	valRegions = map[string]error{
		"LON1": nil,
		"FRA1": nil,
		"NYC1": nil,
	}
	reg, _ := fakeClient.ListRegions()
	for _, r := range reg {
		valRegions[r.Code] = nil
	}
	for k, _ := range valRegions {
		availableRegions = append(availableRegions, k)
	}

	forTesting := map[string]error{
		"Lon!": fmt.Errorf("invalid"),
		"":     fmt.Errorf("invalid"),
		"NYC1": nil,
	}
	for key, val := range forTesting {
		if err := isValidRegion(availableRegions, key); (err != nil && val == nil) || (err == nil && val != nil) {
			t.Fatalf("Region Code: `%s` Expected `%v` got `%v`", key, val, err)
		}
	}
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

func TestK8sVersion(t *testing.T) {
	testK8sVer := map[string]string{
		"1.26.4": "1.26.4-k3s1",
		"1.27.4": "1.27.4-k3s1",
	}
	var validVersion = func() []string {
		var ret []string
		for _, val := range testK8sVer {
			ret = append(ret, val)
		}
		return ret
	}

	// these are invalid
	// input and output
	forTesting := map[string]string{
		"":       "1.26.4-k3s1",
		"1.28":   "",
		"1.27.4": "1.27.4-k3s1",
	}

	for ver, Rver := range forTesting {
		if version := k8sVersion(ver, validVersion); version != Rver {
			t.Fatalf("version dont match `%s` Expected `%s` but got `%s`", ver, Rver, version)
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
