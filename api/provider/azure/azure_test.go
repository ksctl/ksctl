package azure

import (
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"os"
	"testing"

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

	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
	if no != 2 {
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

func TestManagedCluster(t *testing.T) {
}

func TestHACluster(t *testing.T) {
}
