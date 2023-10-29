package civo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	localstate "github.com/kubesimplify/ksctl/internal/storage/local"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
	"gotest.tools/assert"
)

var (
	fakeClient *CivoProvider
	demoClient *resources.KsctlClient
	dir        = fmt.Sprintf("%s/ksctl-civo-test", os.TempDir())
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
	_ = os.Setenv(string(KsctlCustomDirEnabled), dir)
	civoHA := utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeHa)
	civoManaged := utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeMang)

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
		generatePath(UtilClusterPath, "abcd"),
		utils.GetPath(UtilClusterPath, "civo", "abcd"),
		"genreatePath not compatable with utils.getpath()")
}

func TestIsValidK8sVersion(t *testing.T) {
	ver, _ := fakeClient.client.ListAvailableKubernetesVersions()
	for _, vver := range ver {
		t.Log(vver)
	}
}

func TestCivoProvider_InitState(t *testing.T) {

	// get the data
	fakeClient.region = "LON1"

	t.Run("Create state", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, OperationStateCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeClient.clusterName+" "+fakeClient.region, "clusterdir not equal")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeClient.Name("fake").NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		civoCloudState.IsCompleted = true
		assert.Equal(t, civoCloudState.IsCompleted, true, "cluster should not be completed")

		if err := fakeClient.InitState(demoClient.Storage, OperationStateCreate); err != nil {
			t.Fatalf("Unable to resume state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Get request", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, OperationStateGet); err != nil {
			t.Fatalf("Unable to get state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Delete request", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, OperationStateDelete); err != nil {
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
		"":     "traefik2-nodeport,metrics-server",
		"abcd": "abcd,traefik2-nodeport,metrics-server",
	}

	for apps, setVal := range testPreInstalled {
		if retApps := fakeClient.Application(apps); retApps {
			t.Fatalf("application shouldn't be external flag")
		}
		if fakeClient.metadata.apps != setVal {
			t.Fatalf("apps dont match Expected `%s` but got `%s`", apps, setVal)
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
	fakeClient.mxName.Unlock() // to unlock the mutex lock
	if fakeClient.metadata.resName != "demo" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClient.Name("12demo"); ret != nil {
		t.Fatalf("returned interface for invalid res name")
	}
	fakeClient.mxName.Unlock() // to unlock the mutex lock
}

func TestRole(t *testing.T) {
	validSet := []KsctlRole{RoleCp, RoleLb, RoleDs, RoleWp}
	for _, val := range validSet {
		if ret := fakeClient.Role(val); ret == nil {
			t.Fatalf("returned nil for valid role")
		}
		fakeClient.mxRole.Unlock()
		if fakeClient.metadata.role != val {
			t.Fatalf("Correct assignment missing")
		}
	}
	if ret := fakeClient.Role("fake"); ret != nil {
		t.Fatalf("returned interface for invalid role")
	}
	fakeClient.mxRole.Unlock() // to unlock the mutex lock
}

func TestVMType(t *testing.T) {
	if ret := fakeClient.VMType("g4s.kube.small"); ret == nil {
		t.Fatalf("returned nil for valid vm type")
	}
	fakeClient.mxVMType.Unlock() // to unlock the mutex lock

	if fakeClient.metadata.vmType != "g4s.kube.small" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClient.VMType(""); ret != nil {
		t.Fatalf("returned interface for invalid vm type")
	}
	fakeClient.mxVMType.Unlock() // to unlock the mutex lock

}

func TestVisibility(t *testing.T) {
	if fakeClient.Visibility(true); !fakeClient.metadata.public {
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
			if ver+"-k3s1" != fakeClient.metadata.k8sVersion {
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
	if "1.26.4-k3s1" != fakeClient.metadata.k8sVersion {
		t.Fatalf("set value is not equal to input value")
	}
}

func TestCni(t *testing.T) {
	testCases := map[string]bool{
		string(CNICilium):  false,
		string(CNIFlannel): false,
		string(CNIKubenet): true,
		"abcd":             true,
	}

	for k, v := range testCases {
		got := fakeClient.CNI(k)
		assert.Equal(t, got, v, "missmatch")
	}
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

func checkCurrentStateFile(t *testing.T) {

	raw, err := demoClient.Storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeMang, clusterDirName, STATE_FILE_NAME)).Load()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("Reason: %v", err)
	}

	assert.DeepEqual(t, civoCloudState, data)
}

func checkCurrentStateFileHA(t *testing.T) {

	raw, err := demoClient.Storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeHa, clusterDirName, STATE_FILE_NAME)).Load()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("Reason: %v", err)
	}

	assert.DeepEqual(t, civoCloudState, data)
}

func TestManagedCluster(t *testing.T) {
	fakeClient.region = "LON1"
	t.Run("init state", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, OperationStateCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeClient.clusterName+" "+fakeClient.region, "clusterdir not equal")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")

		_, err := demoClient.Storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeMang, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClient.Name("fake-net").NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(civoCloudState.NetworkIDs.NetworkID) > 0, "network id not saved")

		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		fakeClient.CNI("cilium")
		fakeClient.Application("abcd")

		assert.Equal(t, fakeClient.Name("fake").VMType("g4s.kube.small").NewManagedCluster(demoClient.Storage, 5), nil, "managed cluster should be created")

		assert.Equal(t, civoCloudState.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, civoCloudState.NoManagedNodes, 5)
		assert.Equal(t, civoCloudState.KubernetesDistro, string(K8sK3s))
		assert.Equal(t, civoCloudState.KubernetesVer, fakeClient.metadata.k8sVersion)
		assert.Assert(t, len(civoCloudState.ManagedClusterID) > 0, "Managed clusterID not saved")

		_, err := demoClient.Storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeMang, clusterDirName, KUBECONFIG_FILE_NAME)).Load()
		if os.IsNotExist(err) {
			t.Fatalf("kubeconfig should not be absent")
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:     fakeClient.clusterName,
				Provider: CloudCivo,
				Type:     ClusterTypeMang,
				Region:   fakeClient.region,
				NoMgt:    civoCloudState.NoManagedNodes,

				K8sDistro:  K8sK3s,
				K8sVersion: civoCloudState.KubernetesVer,
			},
		}
		got, err := GetRAWClusterInfos(demoClient.Storage)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClient.DelManagedCluster(demoClient.Storage), nil, "managed cluster should be deleted")

		assert.Equal(t, len(civoCloudState.ManagedClusterID), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClient.DelNetwork(demoClient.Storage), nil, "Network should be deleted")

		assert.Equal(t, len(civoCloudState.NetworkIDs.NetworkID), 0, "network id still present")
		// at this moment the file is not present
		_, err := demoClient.Storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeMang, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}

func TestHACluster(t *testing.T) {

	fakeClient.region = "LON1"
	fakeClient.clusterName = "fakekeke"
	fakeClient.haCluster = true

	// size
	fakeClient.metadata.noCP = 7
	fakeClient.metadata.noDS = 5
	fakeClient.metadata.noWP = 10
	fakeClient.metadata.public = true
	fakeClient.metadata.vmType = "g4s.kube.small"
	fakeClient.metadata.k8sName = K8sK3s

	t.Run("init state", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, OperationStateCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeClient.clusterName+" "+fakeClient.region, "clusterdir not equal")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")

		_, err := demoClient.Storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, ClusterTypeHa, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClient.Name("fake-net").NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(civoCloudState.NetworkIDs.NetworkID) > 0, "network id not saved")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClient.Name("fake-ssh").CreateUploadSSHKeyPair(demoClient.Storage), nil, "ssh key failed")

		assert.Assert(t, len(civoCloudState.SSHID) > 0, "sshid must be present")
		assert.Equal(t, civoCloudState.SSHUser, "root", "ssh user not set")
		assert.Equal(t, civoCloudState.SSHPrivateKeyLoc, utils.GetPath(UtilSSHPath, CloudCivo, clusterType, clusterDirName), "ssh private key loc missing")

		assert.Equal(t, civoCloudState.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClient.Role(RoleCp)
			fakeClient.Name("fake-cp")

			assert.Equal(t, fakeClient.NewFirewall(demoClient.Storage), nil, "new firewall failed")

			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode) > 0, "firewallID for controlplane absent")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClient.Role(RoleWp)
			fakeClient.Name("fake-wp")

			assert.Equal(t, fakeClient.NewFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDWorkerNode) > 0, "firewallID for workerplane absent")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClient.Role(RoleLb)
			fakeClient.Name("fake-lb")

			assert.Equal(t, fakeClient.NewFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode) > 0, "firewallID for loadbalancer absent")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClient.Role(RoleDs)
			fakeClient.Name("fake-ds")

			assert.Equal(t, fakeClient.NewFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDDatabaseNode) > 0, "firewallID for datastore absent")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClient.Role(RoleLb)
			fakeClient.Name("fake-lb-vm")
			fakeClient.VMType("g4s.kube.small")

			assert.Equal(t, fakeClient.NewVM(demoClient.Storage, 0), nil, "new vm failed")
			assert.Assert(t, len(civoCloudState.InstanceIDs.LoadBalancerNode) > 0, "loadbalancer VM id absent")

			assert.Assert(t, len(civoCloudState.IPv4.IPLoadbalancer) > 0, "loadbalancer ipv4 absent")
			assert.Assert(t, len(civoCloudState.IPv4.PrivateIPLoadbalancer) > 0, "loadbalancer private ipv4 absent")
			assert.Assert(t, len(civoCloudState.HostNames.LoadBalancerNode) > 0, "loadbalancer hostname absent")

			checkCurrentStateFileHA(t)
		})
		t.Run("Controlplanes", func(t *testing.T) {

			if _, err := fakeClient.NoOfControlPlane(fakeClient.metadata.noCP, true); err != nil {
				t.Fatalf("Failed to set the controlplane")
			}

			for i := 0; i < fakeClient.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					fakeClient.Name(fmt.Sprintf("fake-cp-%d", i))
					fakeClient.VMType("g4s.kube.small")
					fakeClient.Role(RoleCp)

					assert.Equal(t, fakeClient.NewVM(demoClient.Storage, i), nil, "new vm failed")
					assert.Assert(t, len(civoCloudState.InstanceIDs.ControlNodes[i]) > 0, "controlplane VM id absent")

					assert.Assert(t, len(civoCloudState.IPv4.IPControlplane[i]) > 0, "controlplane ipv4 absent")
					assert.Assert(t, len(civoCloudState.IPv4.PrivateIPControlplane[i]) > 0, "controlplane private ipv4 absent")
					assert.Assert(t, len(civoCloudState.HostNames.ControlNodes[i]) > 0, "controlplane hostname absent")

					checkCurrentStateFileHA(t)
				})
			}
		})

		t.Run("Datastores", func(t *testing.T) {
			// NOTE: the noDS is set to 1 becuase current implementation is only for single datastore
			// TODO: use the 1 as limit

			fakeClient.metadata.noDS = 1

			if _, err := fakeClient.NoOfDataStore(fakeClient.metadata.noDS, true); err != nil {
				t.Fatalf("Failed to set the datastore")
			}

			for i := 0; i < fakeClient.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					fakeClient.Name(fmt.Sprintf("fake-ds-%d", i))
					fakeClient.VMType("g4s.kube.small")
					fakeClient.Role(RoleDs)

					assert.Equal(t, fakeClient.NewVM(demoClient.Storage, i), nil, "new vm failed")
					assert.Assert(t, len(civoCloudState.InstanceIDs.DatabaseNode[i]) > 0, "datastore VM id absent")

					assert.Assert(t, len(civoCloudState.IPv4.IPDataStore[i]) > 0, "datastore ipv4 absent")
					assert.Assert(t, len(civoCloudState.IPv4.PrivateIPDataStore[i]) > 0, "datastore private ipv4 absent")
					assert.Assert(t, len(civoCloudState.HostNames.DatabaseNode[i]) > 0, "datastore hostname absent")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Workplanes", func(t *testing.T) {

			if _, err := fakeClient.NoOfWorkerPlane(demoClient.Storage, fakeClient.metadata.noWP, true); err != nil {
				t.Fatalf("Failed to set the workerplane")
			}

			for i := 0; i < fakeClient.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {

					fakeClient.Name(fmt.Sprintf("fake-wp-%d", i))
					fakeClient.Role(RoleWp)
					fakeClient.VMType("g4s.kube.small")

					assert.Equal(t, fakeClient.NewVM(demoClient.Storage, i), nil, "new vm failed")
					assert.Assert(t, len(civoCloudState.InstanceIDs.WorkerNodes[i]) > 0, "workerplane VM id absent")

					assert.Assert(t, len(civoCloudState.IPv4.IPWorkerPlane[i]) > 0, "workerplane ipv4 absent")
					assert.Assert(t, len(civoCloudState.IPv4.PrivateIPWorkerPlane[i]) > 0, "workerplane private ipv4 absent")
					assert.Assert(t, len(civoCloudState.HostNames.WorkerNodes[i]) > 0, "workerplane hostname absent")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, civoCloudState.IsCompleted, true, "cluster should not be completed")
		})
	})

	fmt.Println(fakeClient.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := civoCloudState.HostNames.WorkerNodes

		got := fakeClient.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClient.GetStateFile(demoClient.Storage)
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(civoCloudState)
		assert.DeepEqual(t, string(got), expected)
	})

	// TODO: check for the Passing the state to the kubernetes distribution function GetStateForHACluster

	t.Run("Get cluster ha", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:       fakeClient.clusterName,
				Provider:   CloudCivo,
				Type:       ClusterTypeHa,
				Region:     fakeClient.region,
				NoWP:       fakeClient.noWP,
				NoCP:       fakeClient.noCP,
				NoDS:       fakeClient.noDS,
				K8sDistro:  K8sK3s,
				K8sVersion: civoCloudState.KubernetesVer,
			},
		}
		got, err := GetRAWClusterInfos(demoClient.Storage)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	// explicit clean
	civoCloudState = nil

	// use init state firest
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeClient.InitState(demoClient.Storage, OperationStateDelete); err != nil {
			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
		}

		assert.Equal(t, clusterType, ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeClient.clusterName+" "+fakeClient.region, "clusterdir not equal")
	})

	t.Run("Get all counters", func(t *testing.T) {
		var err error
		fakeClient.metadata.noCP, err = fakeClient.NoOfControlPlane(-1, false)
		assert.Assert(t, err == nil)

		fakeClient.metadata.noWP, err = fakeClient.NoOfWorkerPlane(demoClient.Storage, -1, false)
		assert.Assert(t, err == nil)

		fakeClient.metadata.noDS, err = fakeClient.NoOfDataStore(-1, false)
		assert.Assert(t, err == nil)
	})

	t.Run("Delete VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClient.Role(RoleLb)

			assert.Equal(t, fakeClient.DelVM(demoClient.Storage, 0), nil, "del vm failed")
			assert.Assert(t, len(civoCloudState.InstanceIDs.LoadBalancerNode) == 0, "loadbalancer VM id absent")

			assert.Assert(t, len(civoCloudState.IPv4.IPLoadbalancer) == 0, "loadbalancer ipv4 absent")
			assert.Assert(t, len(civoCloudState.IPv4.PrivateIPLoadbalancer) == 0, "loadbalancer private ipv4 present")
			assert.Assert(t, len(civoCloudState.HostNames.LoadBalancerNode) == 0, "loadbalancer hostname present")

			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClient.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClient.Role(RoleWp)

					assert.Equal(t, fakeClient.DelVM(demoClient.Storage, i), nil, "del vm failed")
					assert.Assert(t, len(civoCloudState.InstanceIDs.WorkerNodes[i]) == 0, "workerplane VM id present")

					assert.Assert(t, len(civoCloudState.IPv4.IPWorkerPlane[i]) == 0, "workerplane ipv4 present")
					assert.Assert(t, len(civoCloudState.IPv4.PrivateIPWorkerPlane[i]) == 0, "workerplane private ipv4 present")
					assert.Assert(t, len(civoCloudState.HostNames.WorkerNodes[i]) == 0, "workerplane hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClient.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClient.Role(RoleCp)

					assert.Equal(t, fakeClient.DelVM(demoClient.Storage, i), nil, "del vm failed")
					assert.Assert(t, len(civoCloudState.InstanceIDs.ControlNodes[i]) == 0, "controlplane VM id present")

					assert.Assert(t, len(civoCloudState.IPv4.IPControlplane[i]) == 0, "controlplane ipv4 present")
					assert.Assert(t, len(civoCloudState.IPv4.PrivateIPControlplane[i]) == 0, "controlplane private ipv4 present")
					assert.Assert(t, len(civoCloudState.HostNames.ControlNodes[i]) == 0, "controlplane hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClient.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClient.Role(RoleDs)

					assert.Equal(t, fakeClient.DelVM(demoClient.Storage, i), nil, "del vm failed")
					assert.Assert(t, len(civoCloudState.InstanceIDs.DatabaseNode[i]) == 0, "datastore VM id present")

					assert.Assert(t, len(civoCloudState.IPv4.IPDataStore[i]) == 0, "datastore ipv4 present")
					assert.Assert(t, len(civoCloudState.IPv4.PrivateIPDataStore[i]) == 0, "datastore private ipv4 present")
					assert.Assert(t, len(civoCloudState.HostNames.DatabaseNode[i]) == 0, "datastore hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClient.Role(RoleCp)

			assert.Equal(t, fakeClient.DelFirewall(demoClient.Storage), nil, "del firewall failed")

			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode) == 0, "firewallID for controlplane present")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClient.Role(RoleWp)

			assert.Equal(t, fakeClient.DelFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDWorkerNode) == 0, "firewallID for workerplane present")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClient.Role(RoleLb)

			assert.Equal(t, fakeClient.DelFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode) == 0, "firewallID for loadbalancer present")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClient.Role(RoleDs)

			assert.Equal(t, fakeClient.DelFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Assert(t, len(civoCloudState.NetworkIDs.FirewallIDDatabaseNode) == 0, "firewallID for datastore present")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClient.DelSSHKeyPair(demoClient.Storage), nil, "ssh key failed")

		assert.Assert(t, len(civoCloudState.SSHID) == 0, "sshid still present")
		assert.Equal(t, civoCloudState.SSHUser, "", "ssh user set")
		assert.Equal(t, civoCloudState.SSHPrivateKeyLoc, "", "ssh private key loc still present")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClient.DelNetwork(demoClient.Storage), nil, "Network should be deleted")
		assert.Assert(t, len(civoCloudState.NetworkIDs.NetworkID) == 0, "network id still present")
	})

}

func TestGetSecretTokens(t *testing.T) {
	t.Run("expect demo data", func(t *testing.T) {
		expected := map[string][]byte{
			"CIVO_TOKEN": []byte("demo-fake"),
		}

		for key, val := range expected {
			assert.NilError(t, os.Setenv(key, string(val)), "environment vars should be set")
		}
		actual, err := fakeClient.GetSecretTokens(demoClient.Storage)
		assert.NilError(t, err, "unable to get the secret token from the client")
		assert.DeepEqual(t, actual, expected)
	})
}
