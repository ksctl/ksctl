package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"
	"gotest.tools/assert"
)

var (
	demoClient *resources.KsctlClient
	fakeAzure  *AzureProvider
	dir        = fmt.Sprintf("%s/ksctl-azure-test", os.TempDir())
)

func TestMain(m *testing.M) {

	demoClient = &resources.KsctlClient{}
	demoClient.Metadata.ClusterName = "fake"
	demoClient.Metadata.Region = "fake"
	demoClient.Metadata.Provider = utils.CLOUD_AZURE
	demoClient.Cloud, _ = ReturnAzureStruct(demoClient.Metadata, ProvideMockClient)

	fakeAzure, _ = ReturnAzureStruct(demoClient.Metadata, ProvideMockClient)

	demoClient.Storage = localstate.InitStorage(false)
	_ = os.Setenv(utils.KSCTL_TEST_DIR_ENABLED, dir)
	azHA := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_HA)
	azManaged := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_MANG)

	if err := os.MkdirAll(azManaged, 0755); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(azHA, 0755); err != nil {
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

func TestInitState(t *testing.T) {

	t.Run("Create state", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, utils.CLUSTER_TYPE_MANG, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeAzure.ClusterName+" "+fakeAzure.ResourceGroup+" "+fakeAzure.Region, "clusterdir not equal")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeAzure.NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		azureCloudState.IsCompleted = true
		assert.Equal(t, azureCloudState.IsCompleted, true, "cluster should not be completed")

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to resume state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Get request", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_GET); err != nil {
			t.Fatalf("Unable to get state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Delete request", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_DELETE); err != nil {
			t.Fatalf("Unable to Delete state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Invalid request", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, "test"); err == nil {
			t.Fatalf("Expected error but not got: %v", err)
		}
	})
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

// Test for the Noof WP and setter and getter
func TestNoOfControlPlane(t *testing.T) {
	var no int
	var err error
	no, err = demoClient.Cloud.NoOfControlPlane(-1, false)
	if no != -1 || err != nil {
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

func TestNoOfDataStore(t *testing.T) {
	var no int
	var err error
	no, err = demoClient.Cloud.NoOfDataStore(-1, false)
	if no != -1 || err != nil {
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

func TestNoOfWorkerPlane(t *testing.T) {
	var no int
	var err error
	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
	if no != -1 || err != nil {
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

func TestValidRegion(t *testing.T) {
	fortesting := map[string]error{
		"fake":    nil,
		"eastus":  errors.New("Error"),
		"eastus2": nil,
	}

	for key, val := range fortesting {
		if aErr := isValidRegion(fakeAzure, key); (aErr != nil && val == nil) || (aErr == nil && val != nil) {
			t.Fatalf("For Region `%s`. Expected `%v` but got `%v`", key, val, aErr)
		}
	}
}

func TestResName(t *testing.T) {

	if ret := fakeAzure.Name("demo"); ret == nil {
		t.Fatalf("returned nil for valid res name")
	}
	if fakeAzure.Metadata.ResName != "demo" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeAzure.Name("12demo"); ret != nil {
		t.Fatalf("returned interface for invalid res name")
	}
}

func TestRole(t *testing.T) {
	validSet := []string{utils.ROLE_CP, utils.ROLE_LB, utils.ROLE_DS, utils.ROLE_WP}
	for _, val := range validSet {
		if ret := fakeAzure.Role(val); ret == nil {
			t.Fatalf("returned nil for valid role")
		}
		if fakeAzure.Metadata.Role != val {
			t.Fatalf("Correct assignment missing")
		}
	}
	if ret := fakeAzure.Role("fake"); ret != nil {
		t.Fatalf("returned interface for invalid role")
	}
}

func TestVMType(t *testing.T) {
	if ret := fakeAzure.VMType("fake"); ret == nil {
		t.Fatalf("returned nil for valid vm type")
	}
	if fakeAzure.Metadata.VmType != "fake" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeAzure.VMType(""); ret != nil {
		t.Fatalf("returned interface for invalid vm type")
	}
}

func TestVisibility(t *testing.T) {
	if fakeAzure.Visibility(true); !fakeAzure.Metadata.Public {
		t.Fatalf("Visibility setting not working")
	}
}

// Mock the return of ValidListOfRegions
func TestRegion(t *testing.T) {

	forTesting := map[string]error{
		"Lon!": errors.New(""),
		"":     errors.New(""),
		"fake": nil,
	}

	for key, val := range forTesting {
		if err := isValidRegion(fakeAzure, key); (err == nil && val != nil) || (err != nil && val == nil) {
			t.Fatalf("Input region :`%s`. expected `%v` but got `%v`", key, val, err)
		}
	}
}

func TestK8sVersion(t *testing.T) {
	// these are invalid
	// input and output
	forTesting := []string{
		"1.27.1",
		"1.27",
		"1.28.1",
	}

	for i := 0; i < len(forTesting); i++ {
		var ver string = forTesting[i]
		if i < 2 {
			if ret := fakeAzure.Version(ver); ret == nil {
				t.Fatalf("returned nil for valid version")
			}
			if ver != fakeAzure.Metadata.K8sVersion {
				t.Fatalf("set value is not equal to input value")
			}
		} else {
			if ret := fakeAzure.Version(ver); ret != nil {
				t.Fatalf("returned interface for invalid version")
			}
		}
	}

}

func TestCniAndOthers(t *testing.T) {
	t.Run("CNI Support flag", func(t *testing.T) {
		if fakeAzure.SupportForCNI() {
			t.Fatal("Support for CNI must be false")
		}
	})

	t.Run("Application support flag", func(t *testing.T) {
		if fakeAzure.SupportForApplications() {
			t.Fatal("Support for Application must be false")
		}
	})

	t.Run("CNI set functionality", func(t *testing.T) {
		if ret := fakeAzure.CNI("cilium"); ret == nil {
			t.Fatalf("returned nil for valid CNI")
		}
	})
}

func TestFirewallRules(t *testing.T) {
	t.Run("Controlplane fw rules", func(t *testing.T) {
		assert.DeepEqual(t, []*armnetwork.SecurityRule{
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_6443"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](100),
					Description:              to.Ptr("sample network security group inbound port 6443"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				},
			},
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_30_to_35k"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](101),
					Description:              to.Ptr("sample network security group inbound port 30000-35000"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
				},
			},
		}, firewallRuleControlPlane())
	})

	t.Run("Workerplane fw rules", func(t *testing.T) {
		assert.DeepEqual(t, []*armnetwork.SecurityRule{
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_6443"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](100),
					Description:              to.Ptr("sample network security group inbound port 6443"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				},
			},
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_30_to_35k"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](101),
					Description:              to.Ptr("sample network security group inbound port 30000-35000"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
				},
			},
		}, firewallRuleWorkerPlane())
	})

	t.Run("Loadbalancer fw rules", func(t *testing.T) {

		assert.DeepEqual(t, []*armnetwork.SecurityRule{
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_6443"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](100),
					Description:              to.Ptr("sample network security group inbound port 6443"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				},
			}, &armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_30_to_35k"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](101),
					Description:              to.Ptr("sample network security group inbound port 30000-35000"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
				},
			},
		}, firewallRuleLoadBalancer())
	})

	t.Run("Datastore fw rules", func(t *testing.T) {
		assert.DeepEqual(t, []*armnetwork.SecurityRule{
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_6443"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](100),
					Description:              to.Ptr("sample network security group inbound port 6443"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				},
			},
			&armnetwork.SecurityRule{
				Name: to.Ptr("sample_inbound_30_to_35k"),
				Properties: &armnetwork.SecurityRulePropertiesFormat{
					SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
					SourcePortRange:          to.Ptr("*"),
					DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
					DestinationPortRange:     to.Ptr("*"),
					Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
					Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
					Priority:                 to.Ptr[int32](101),
					Description:              to.Ptr("sample network security group inbound port 30000-35000"),
					Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
				},
			},
		}, firewallRuleDataStore())

	})
}

func checkCurrentStateFile(t *testing.T) {

	raw, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_MANG, clusterDirName, STATE_FILE_NAME)).Load()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("Reason: %v", err)
	}

	assert.DeepEqual(t, azureCloudState, data)
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

func TestManagedCluster(t *testing.T) {
	fakeAzure.Version("1.27")
	fakeAzure.VMType("fake")
	t.Run("init state", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, utils.CLUSTER_TYPE_MANG, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeAzure.ClusterName+" "+fakeAzure.ResourceGroup+" "+fakeAzure.Region, "clusterdir not equal")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")

		_, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_MANG, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeAzure.NewNetwork(demoClient.Storage), nil, "resource grp should be created")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(azureCloudState.ResourceGroupName) > 0)
		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		assert.Equal(t, fakeAzure.NewManagedCluster(demoClient.Storage, 5), nil, "managed cluster should be created")
		assert.Equal(t, azureCloudState.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, azureCloudState.NoManagedNodes, 5)
		//assert.Equal(t, azureCloudState.KubernetesDistro, utils.K8S_K3S)
		assert.Equal(t, azureCloudState.KubernetesVer, fakeAzure.Metadata.K8sVersion)
		assert.Assert(t, len(azureCloudState.ManagedClusterName) > 0, "Managed cluster Name not saved")

		_, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_MANG, clusterDirName, KUBECONFIG_FILE_NAME)).Load()
		if os.IsNotExist(err) {
			t.Fatalf("kubeconfig should not be absent")
		}
		checkCurrentStateFile(t)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeAzure.DelManagedCluster(demoClient.Storage), nil, "managed cluster should be deleted")

		assert.Equal(t, len(azureCloudState.ManagedClusterName), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeAzure.DelNetwork(demoClient.Storage), nil, "Network should be deleted")

		assert.Equal(t, len(azureCloudState.ResourceGroupName), 0, "resource grp still present")
		// at this moment the file is not present
		_, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, utils.CLUSTER_TYPE_MANG, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}

func TestHACluster(t *testing.T) {

	fakeAzure.Region = "fake"
	fakeAzure.ClusterName = "fakeazure"
	fakeAzure.HACluster = true

	// size
	fakeAzure.Metadata.NoCP = 7
	fakeAzure.Metadata.NoDS = 5
	fakeAzure.Metadata.NoWP = 10
	fakeAzure.Metadata.Public = true
	fakeAzure.Metadata.VmType = "fake"
	fakeAzure.Metadata.K8sName = utils.K8S_K3S
	fakeAzure.VMType("fake")

	t.Run("init state", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_CREATE); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, utils.CLUSTER_TYPE_HA, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeAzure.ClusterName+" "+fakeAzure.ResourceGroup+" "+fakeAzure.Region, "clusterdir not equal")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")

		_, err := demoClient.Storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA, clusterDirName, STATE_FILE_NAME)).Load()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeAzure.NewNetwork(demoClient.Storage), nil, "Network should be created")
		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, azureCloudState.ResourceGroupName, fakeAzure.ResourceGroup, "resource group not saved")
		assert.Equal(t, azureCloudState.VirtualNetworkName, fakeAzure.ClusterName+"-vnet", "virtual net should be created")
		assert.Equal(t, azureCloudState.SubnetName, fakeAzure.ClusterName+"-subnet", "subnet should be created")

		assert.Assert(t, len(azureCloudState.VirtualNetworkID) > 0, "virtual net should be created")
		assert.Assert(t, len(azureCloudState.SubnetID) > 0, "subnet should be created")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		fakeAzure.Name("fake-ssh-key") // only to test the name working
		assert.Equal(t, fakeAzure.CreateUploadSSHKeyPair(demoClient.Storage), nil, "ssh key failed")

		assert.Equal(t, azureCloudState.SSHKeyName, fakeAzure.Metadata.ResName, "sshid must be present")

		assert.Equal(t, azureCloudState.SSHUser, "azureuser", "ssh user not set")
		assert.Equal(t, azureCloudState.SSHPrivateKeyLoc, utils.GetPath(utils.SSH_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName), "ssh private key loc missing")

		assert.Equal(t, azureCloudState.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_CP)
			fakeAzure.Name("fake-fw-cp")

			assert.Equal(t, fakeAzure.NewFirewall(demoClient.Storage), nil, "new firewall failed")

			assert.Equal(t, azureCloudState.InfoControlPlanes.NetworkSecurityGroupName, fakeAzure.Metadata.ResName, "firewallID for controlplane absent")
			assert.Assert(t, len(azureCloudState.InfoControlPlanes.NetworkSecurityGroupID) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_WP)
			fakeAzure.Name("fake-fw-wp")

			assert.Equal(t, fakeAzure.NewFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Equal(t, azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName, fakeAzure.Metadata.ResName, "firewallID for workerplane absent")
			assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_LB)
			fakeAzure.Name("fake-fw-lb")

			assert.Equal(t, fakeAzure.NewFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName, fakeAzure.Metadata.ResName, "firewallID for loadbalacer absent")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_DS)
			fakeAzure.Name("fake-fw-ds")

			assert.Equal(t, fakeAzure.NewFirewall(demoClient.Storage), nil, "new firewall failed")
			assert.Equal(t, azureCloudState.InfoDatabase.NetworkSecurityGroupName, fakeAzure.Metadata.ResName, "firewallID for datastore absent")
			assert.Assert(t, len(azureCloudState.InfoDatabase.NetworkSecurityGroupID) > 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_LB)
			fakeAzure.Name("fake-lb")

			assert.Equal(t, fakeAzure.NewVM(demoClient.Storage, 0), nil, "new vm failed")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.Name, fakeAzure.Metadata.ResName, "missmatch of Loadbalancer VM name")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.DiskName, fakeAzure.Metadata.ResName+"-disk", "missmatch of Loadbalancer disk name")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.PublicIPName, fakeAzure.Metadata.ResName+"-pub", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.PublicIPID) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.NetworkInterfaceName, fakeAzure.Metadata.ResName+"-nic", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.NetworkInterfaceID) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.PrivateIP, "192.168.X.Y", "missmatch of Loadbalancer private ip NIC")

			checkCurrentStateFileHA(t)
		})
		t.Run("Controlplanes", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_CP)

			if _, err := fakeAzure.NoOfControlPlane(fakeAzure.Metadata.NoCP, true); err != nil {
				t.Fatalf("Failed to set the controlplane")
			}

			for i := 0; i < fakeAzure.Metadata.NoCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					fakeAzure.Name(fmt.Sprintf("fake-cp-%d", i))

					assert.Equal(t, fakeAzure.NewVM(demoClient.Storage, i), nil, "new vm failed")
					assert.Equal(t, azureCloudState.InfoControlPlanes.Names[i], fakeAzure.Metadata.ResName, "missmatch of controlplane VM name")
					assert.Assert(t, len(azureCloudState.InfoControlPlanes.Hostnames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, azureCloudState.InfoControlPlanes.DiskNames[i], fakeAzure.Metadata.ResName+"-disk", "missmatch of controlplane disk name")

					assert.Equal(t, azureCloudState.InfoControlPlanes.PublicIPNames[i], fakeAzure.Metadata.ResName+"-pub", "missmatch of controlplane pub ip name")
					assert.Assert(t, len(azureCloudState.InfoControlPlanes.PublicIPIDs[i]) > 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, azureCloudState.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Equal(t, azureCloudState.InfoControlPlanes.NetworkInterfaceNames[i], fakeAzure.Metadata.ResName+"-nic", "missmatch of controlplane nic name")
					assert.Assert(t, len(azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, azureCloudState.InfoControlPlanes.PrivateIPs[i], "192.168.X.Y", "missmatch of controlplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})

		t.Run("Datastores", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_DS)
			// NOTE: the noDS is set to 1 becuase current implementation is only for single datastore
			// TODO: use the 1 as limit

			fakeAzure.Metadata.NoDS = 1

			if _, err := fakeAzure.NoOfDataStore(fakeAzure.Metadata.NoDS, true); err != nil {
				t.Fatalf("Failed to set the datastore")
			}

			for i := 0; i < fakeAzure.Metadata.NoDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					fakeAzure.Name(fmt.Sprintf("fake-ds-%d", i))

					assert.Equal(t, fakeAzure.NewVM(demoClient.Storage, i), nil, "new vm failed")

					assert.Equal(t, azureCloudState.InfoDatabase.Names[i], fakeAzure.Metadata.ResName, "missmatch of datastore VM name")
					assert.Assert(t, len(azureCloudState.InfoDatabase.Hostnames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, azureCloudState.InfoDatabase.DiskNames[i], fakeAzure.Metadata.ResName+"-disk", "missmatch of datastore disk name")

					assert.Equal(t, azureCloudState.InfoDatabase.PublicIPNames[i], fakeAzure.Metadata.ResName+"-pub", "missmatch of datastore pub ip name")
					assert.Assert(t, len(azureCloudState.InfoDatabase.PublicIPIDs[i]) > 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, azureCloudState.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Equal(t, azureCloudState.InfoDatabase.NetworkInterfaceNames[i], fakeAzure.Metadata.ResName+"-nic", "missmatch of datastore nic name")
					assert.Assert(t, len(azureCloudState.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, azureCloudState.InfoDatabase.PrivateIPs[i], "192.168.X.Y", "missmatch of datastore private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Workplanes", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_WP)

			if _, err := fakeAzure.NoOfWorkerPlane(demoClient.Storage, fakeAzure.Metadata.NoWP, true); err != nil {
				t.Fatalf("Failed to set the workerplane")
			}

			for i := 0; i < fakeAzure.Metadata.NoWP; i++ {
				t.Run("workerplane", func(t *testing.T) {

					fakeAzure.Name(fmt.Sprintf("fake-wp-%d", i))

					assert.Equal(t, fakeAzure.NewVM(demoClient.Storage, i), nil, "new vm failed")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.Names[i], fakeAzure.Metadata.ResName, "missmatch of workerplane VM name")
					assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.Hostnames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.DiskNames[i], fakeAzure.Metadata.ResName+"-disk", "missmatch of workerplane disk name")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.PublicIPNames[i], fakeAzure.Metadata.ResName+"-pub", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.PublicIPIDs[i]) > 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, azureCloudState.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[i], fakeAzure.Metadata.ResName+"-nic", "missmatch of workerplane nic name")
					assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, azureCloudState.InfoWorkerPlanes.PrivateIPs[i], "192.168.X.Y", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, azureCloudState.IsCompleted, true, "cluster should be completed")
		})
	})

	fmt.Println(fakeAzure.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := azureCloudState.InfoWorkerPlanes.Hostnames

		got := fakeAzure.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	// explicit clean
	azureCloudState = nil

	// TODO: check for the Passing the state to the kubernetes distribution function GetStateForHACluster

	// use init state firest
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeAzure.InitState(demoClient.Storage, utils.OPERATION_STATE_DELETE); err != nil {
			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
		}

		assert.Equal(t, clusterType, utils.CLUSTER_TYPE_HA, "clustertype should be managed")
		assert.Equal(t, clusterDirName, fakeAzure.ClusterName+" "+fakeAzure.ResourceGroup+" "+fakeAzure.Region, "clusterdir not equal")
	})

	t.Run("Get all counters", func(t *testing.T) {
		var err error
		fakeAzure.Metadata.NoCP, err = fakeAzure.NoOfControlPlane(-1, false)
		assert.Assert(t, err == nil)

		fakeAzure.Metadata.NoWP, err = fakeAzure.NoOfWorkerPlane(demoClient.Storage, -1, false)
		assert.Assert(t, err == nil)

		fakeAzure.Metadata.NoDS, err = fakeAzure.NoOfDataStore(-1, false)
		assert.Assert(t, err == nil)
	})

	t.Run("Delete VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_LB)

			assert.Equal(t, fakeAzure.DelVM(demoClient.Storage, 0), nil, "del vm failed")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.Name, "", "missmatch of Loadbalancer VM name")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.HostName, "", "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.DiskName, "", "missmatch of Loadbalancer disk name")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.PublicIPName, "", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.PublicIPID) == 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.PublicIP, "", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.NetworkInterfaceName, "", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.NetworkInterfaceID) == 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, azureCloudState.InfoLoadBalancer.PrivateIP, "", "missmatch of Loadbalancer private ip NIC")
			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_WP)

			for i := 0; i < fakeAzure.Metadata.NoWP; i++ {
				t.Run("workerplane", func(t *testing.T) {

					assert.Equal(t, fakeAzure.DelVM(demoClient.Storage, i), nil, "del vm failed")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.Names[i], "", "missmatch of workerplane VM name")
					assert.Equal(t, azureCloudState.InfoWorkerPlanes.Hostnames[i], "", "missmatch of workerplane vm hostname")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.DiskNames[i], "", "missmatch of workerplane disk name")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.PublicIPNames[i], "", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.PublicIPIDs[i]) == 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, azureCloudState.InfoWorkerPlanes.PublicIPs[i], "", "missmatch of workerplane pub ip")

					assert.Equal(t, azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[i], "", "missmatch of workerplane nic name")
					assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, azureCloudState.InfoWorkerPlanes.PrivateIPs[i], "", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_CP)

			for i := 0; i < fakeAzure.Metadata.NoCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					assert.Equal(t, fakeAzure.DelVM(demoClient.Storage, i), nil, "del vm failed")

					assert.Equal(t, azureCloudState.InfoControlPlanes.Names[i], "", "missmatch of controlplane VM name")
					assert.Equal(t, azureCloudState.InfoControlPlanes.Hostnames[i], "", "missmatch of controlplane vm hostname")

					assert.Equal(t, azureCloudState.InfoControlPlanes.DiskNames[i], "", "missmatch of controlplane disk name")

					assert.Equal(t, azureCloudState.InfoControlPlanes.PublicIPNames[i], "", "missmatch of controlplane pub ip name")
					assert.Assert(t, len(azureCloudState.InfoControlPlanes.PublicIPIDs[i]) == 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, azureCloudState.InfoControlPlanes.PublicIPs[i], "", "missmatch of controlplane pub ip")

					assert.Equal(t, azureCloudState.InfoControlPlanes.NetworkInterfaceNames[i], "", "missmatch of controlplane nic name")
					assert.Assert(t, len(azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, azureCloudState.InfoControlPlanes.PrivateIPs[i], "", "missmatch of controlplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_DS)

			for i := 0; i < fakeAzure.Metadata.NoDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					assert.Equal(t, fakeAzure.DelVM(demoClient.Storage, i), nil, "del vm failed")

					assert.Equal(t, azureCloudState.InfoDatabase.Names[i], "", "missmatch of datastore VM name")
					assert.Equal(t, azureCloudState.InfoDatabase.Hostnames[i], "", "missmatch of datastore vm hostname")

					assert.Equal(t, azureCloudState.InfoDatabase.DiskNames[i], "", "missmatch of datastore disk name")

					assert.Equal(t, azureCloudState.InfoDatabase.PublicIPNames[i], "", "missmatch of datastore pub ip name")
					assert.Assert(t, len(azureCloudState.InfoDatabase.PublicIPIDs[i]) == 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, azureCloudState.InfoDatabase.PublicIPs[i], "", "missmatch of datastore pub ip")

					assert.Equal(t, azureCloudState.InfoDatabase.NetworkInterfaceNames[i], "", "missmatch of datastore nic name")
					assert.Assert(t, len(azureCloudState.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")
					assert.Equal(t, azureCloudState.InfoDatabase.PrivateIPs[i], "", "missmatch of datastore private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_CP)

			assert.Equal(t, fakeAzure.DelFirewall(demoClient.Storage), nil, "del firewall failed")

			assert.Equal(t, azureCloudState.InfoControlPlanes.NetworkSecurityGroupName, "", "firewallID for controlplane absent")
			assert.Assert(t, len(azureCloudState.InfoControlPlanes.NetworkSecurityGroupID) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_WP)

			assert.Equal(t, fakeAzure.DelFirewall(demoClient.Storage), nil, "new firewall failed")

			assert.Equal(t, azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupName, "", "firewallID for workerplane absent")
			assert.Assert(t, len(azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_LB)

			assert.Equal(t, fakeAzure.DelFirewall(demoClient.Storage), nil, "new firewall failed")

			assert.Equal(t, azureCloudState.InfoLoadBalancer.NetworkSecurityGroupName, "", "firewallID for loadbalacer absent")
			assert.Assert(t, len(azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeAzure.Role(utils.ROLE_DS)

			assert.Equal(t, fakeAzure.DelFirewall(demoClient.Storage), nil, "new firewall failed")

			assert.Equal(t, azureCloudState.InfoDatabase.NetworkSecurityGroupName, "", "firewallID for datastore absent")
			assert.Assert(t, len(azureCloudState.InfoDatabase.NetworkSecurityGroupID) == 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeAzure.DelSSHKeyPair(demoClient.Storage), nil, "ssh key failed")

		assert.Equal(t, azureCloudState.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, azureCloudState.SSHUser, "", "ssh user not set")
		assert.Equal(t, azureCloudState.SSHPrivateKeyLoc, "", "ssh private key loc missing")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeAzure.DelNetwork(demoClient.Storage), nil, "Network should be deleted")

		assert.Equal(t, azureCloudState.ResourceGroupName, "", "resource group not saved")
		assert.Equal(t, azureCloudState.VirtualNetworkName, "", "virtual net should be created")
		assert.Equal(t, azureCloudState.SubnetName, "", "subnet should be created")

		assert.Assert(t, len(azureCloudState.VirtualNetworkID) == 0, "virtual net should be created")
		assert.Assert(t, len(azureCloudState.SubnetID) == 0, "subnet should be created")
	})
}
