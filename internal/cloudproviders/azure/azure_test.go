package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/ksctl/ksctl/pkg/logger"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"gotest.tools/v3/assert"
)

var (
	fakeClientHA *AzureProvider
	storeHA      types.StorageFactory

	fakeClientManaged *AzureProvider
	storeManaged      types.StorageFactory

	fakeClientVars *AzureProvider
	storeVars      types.StorageFactory
	parentCtx      context.Context
	parentLogger   types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)

	dir = path.Join(os.TempDir(), "ksctl-azure-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "fake",
		Provider:    consts.CloudAzure,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, ProvideMockClient)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAzure, "fake", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func TestInitState(t *testing.T) {

	t.Run("Create state", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeClientVars.Name("fake-net").NewNetwork(storeVars), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		mainStateDocument.CloudInfra.Azure.B.IsCompleted = true
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, true, "cluster should not be completed")

		if err := fakeClientVars.InitState(storeVars, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to resume state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Get request", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, consts.OperationGet); err != nil {
			t.Fatalf("Unable to get state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Delete request", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, consts.OperationDelete); err != nil {
			t.Fatalf("Unable to Delete state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Invalid request", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, "test"); err == nil {
			t.Fatalf("Expected error but not got: %v", err)
		}
	})
}

// Test for the Noof WP and setter and getter
func TestNoOfControlPlane(t *testing.T) {
	var no int
	var err error
	no, err = fakeClientVars.NoOfControlPlane(-1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized controlplanes array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfControlPlane(1, true)
	// it should return error
	if err == nil {
		t.Fatalf("setter should fail on when no < 3 controlplanes provided_no: %d", 1)
	}

	_, err = fakeClientVars.NoOfControlPlane(5, true)
	// it should return error
	if err != nil {
		t.Fatalf("setter should not fail on when n >= 3 controlplanes err: %v", err)
	}

	no, err = fakeClientVars.NoOfControlPlane(-1, false)
	if no != 5 {
		t.Fatalf("Getter failed to get updated no of controlplanes array got no: %d and err: %v", no, err)
	}
}

func TestNoOfDataStore(t *testing.T) {
	var no int
	var err error
	no, err = fakeClientVars.NoOfDataStore(-1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized datastore array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfDataStore(0, true)
	if err == nil {
		t.Fatalf("setter should fail on when no < 3 datastore provided_no: %d", 1)
	}

	_, err = fakeClientVars.NoOfDataStore(5, true)
	if err != nil {
		t.Fatalf("setter should not fail on when n >= 3 datastore err: %v", err)
	}

	no, err = fakeClientVars.NoOfDataStore(-1, false)
	if no != 5 {
		t.Fatalf("Getter failed to get updated no of datastore array got no: %d and err: %v", no, err)
	}
}

func TestNoOfWorkerPlane(t *testing.T) {
	var no int
	var err error
	no, err = fakeClientVars.NoOfWorkerPlane(storeVars, -1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized workerplane array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(storeVars, 2, true)
	// it shouldn't return err
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("setter should not fail on when no >= 0 workerplane provided_no: %d", 2)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(storeVars, 2, true)
	if err != nil {
		t.Fatalf("setter should return nil when no changes happen workerplane err: %v", err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(storeVars, 3, true)
	if err != nil {
		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(storeVars, 1, true)
	if err != nil {
		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
	}

	no, err = fakeClientVars.NoOfWorkerPlane(storeVars, -1, false)
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
		if aErr := isValidRegion(fakeClientVars, key); (aErr != nil && val == nil) || (aErr == nil && val != nil) {
			t.Fatalf("For Region `%s`. Expected `%v` but got `%v`", key, val, aErr)
		}
	}
}

func TestResName(t *testing.T) {

	if ret := fakeClientVars.Name("demo"); ret == nil {
		t.Fatalf("returned nil for valid res name")
	}

	name := <-fakeClientVars.chResName

	if name != "demo" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClientVars.Name("12demo"); ret != nil {
		t.Fatalf("returned interface for invalid res name")
	}
	//_ = <-fakeClientVars.chResName
}

func TestRole(t *testing.T) {
	validSet := []consts.KsctlRole{consts.RoleCp, consts.RoleLb, consts.RoleDs, consts.RoleWp}
	for _, val := range validSet {
		if ret := fakeClientVars.Role(val); ret == nil {
			t.Fatalf("returned nil for valid role")
		}
		role := <-fakeClientVars.chRole
		if role != val {
			t.Fatalf("Correct assignment missing")
		}
	}
	if ret := fakeClientVars.Role("fake"); ret != nil {
		t.Fatalf("returned interface for invalid role")
	}
	//_ = <-fakeClientVars.chRole
}

func TestVMType(t *testing.T) {
	if ret := fakeClientVars.VMType("fake"); ret == nil {
		t.Fatalf("returned nil for valid vm type")
	}
	vm := <-fakeClientVars.chVMType
	if vm != "fake" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClientVars.VMType(""); ret != nil {
		t.Fatalf("returned interface for invalid vm type")
	}
	//_ = <-fakeClientVars.chVMType
}

func TestVisibility(t *testing.T) {
	if fakeClientVars.Visibility(true); !fakeClientVars.metadata.public {
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
		if err := isValidRegion(fakeClientVars, key); (err == nil && val != nil) || (err != nil && val == nil) {
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
			if ret := fakeClientVars.ManagedK8sVersion(ver); ret == nil {
				t.Fatalf("returned nil for valid version")
			}
			if ver != fakeClientVars.metadata.k8sVersion {
				t.Fatalf("set value is not equal to input value")
			}
		} else {
			if ret := fakeClientVars.ManagedK8sVersion(ver); ret != nil {
				t.Fatalf("returned interface for invalid version")
			}
		}
	}

}

func TestCniAndApps(t *testing.T) {

	testCases := map[string]bool{
		string(consts.CNIAzure):   false,
		string(consts.CNIKubenet): false,
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

func TestFirewallRules(t *testing.T) {
	_rules := []helpers.FirewallRule{
		{
			Description: "nice",
			Name:        "hello",
			Protocol:    consts.FirewallActionUDP,
			Direction:   consts.FirewallActionEgress,
			Action:      consts.FirewallActionDeny,
			Cidr:        "1.1.1./0",
			StartPort:   "34",
			EndPort:     "34",
		},
		{
			Description: "324nice",
			Name:        "he23llo",
			Protocol:    consts.FirewallActionTCP,
			Direction:   consts.FirewallActionIngress,
			Cidr:        "1.1.12./0",
			StartPort:   "1",
			EndPort:     "65000",
		},
	}
	bkp := mainStateDocument.CloudInfra.Azure.NetCidr
	mainStateDocument.CloudInfra.Azure.NetCidr = "x.y.z.a/b"
	defer func() { mainStateDocument.CloudInfra.Azure.NetCidr = bkp }()
	_expected := []*armnetwork.SecurityRule{
		{
			Name: to.Ptr(_rules[0].Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      to.Ptr(mainStateDocument.CloudInfra.Azure.NetCidr),
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr(_rules[0].Cidr),
				DestinationPortRange:     to.Ptr(_rules[0].StartPort),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolUDP),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessDeny),
				Priority:                 to.Ptr[int32](101),
				Description:              to.Ptr(_rules[0].Description),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionOutbound),
			},
		},
		{
			Name: to.Ptr(_rules[1].Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      to.Ptr(_rules[1].Cidr),
				SourcePortRange:          to.Ptr("*"),
				DestinationAddressPrefix: to.Ptr(mainStateDocument.CloudInfra.Azure.NetCidr),
				DestinationPortRange:     to.Ptr(_rules[1].StartPort + "-" + _rules[1].EndPort),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Priority:                 to.Ptr[int32](102),
				Description:              to.Ptr(_rules[1].Description),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
			},
		},
	}
	assert.DeepEqual(t, _expected, convertToProviderSpecific(_rules))

}

func TestDeleteVarCluster(t *testing.T) {
	if err := storeVars.DeleteCluster(); err != nil {
		t.Fatal(err)
	}
}

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(consts.CloudAzure, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeMang); err != nil {
		t.Fatal(err)
	}
	read, err := storeManaged.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, mainStateDocument, read)
}

func checkCurrentStateFileHA(t *testing.T) {

	if err := storeHA.Setup(consts.CloudAzure, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeHa); err != nil {
		t.Fatal(err)
	}
	read, err := storeHA.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, mainStateDocument, read)
}

func TestManagedCluster(t *testing.T) {
	mainStateDocument = &storageTypes.StorageDocument{}
	fakeClientManaged, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo-managed",
		Region:      "fake",
		Provider:    consts.CloudAzure,
	}, parentLogger, mainStateDocument, ProvideMockClient)

	storeManaged = localstate.NewClient(parentCtx, parentLogger)
	_ = storeManaged.Setup(consts.CloudAzure, "fake", "demo-managed", consts.ClusterTypeMang)
	_ = storeManaged.Connect()

	fakeClientManaged.ManagedK8sVersion("1.27")
	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(storeManaged, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-data-will-not-be-used").NewNetwork(storeManaged), nil, "resource grp should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.ResourceGroupName) > 0)
		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		assert.Equal(t, fakeClientManaged.Name("fake-managed").VMType("fake").NewManagedCluster(storeManaged, 5), nil, "managed cluster should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.NoManagedNodes, 5)
		//assert.Equal(t, mainStateDocument.BootstrapProvider, "managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.KubernetesVer, fakeClientManaged.metadata.k8sVersion)
		assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.ManagedClusterName) > 0, "Managed cluster Name not saved")

		_, err := storeManaged.Read()
		if os.IsNotExist(err) {
			t.Fatalf("kubeconfig should not be absent")
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:          fakeClientManaged.clusterName,
				CloudProvider: consts.CloudAzure,
				ClusterType:   consts.ClusterTypeMang,
				Region:        fakeClientManaged.region,
				NoMgt:         mainStateDocument.CloudInfra.Azure.NoManagedNodes,
				Mgt:           cloud.VMData{VMSize: "fake"},
				K8sDistro:     "managed",
				K8sVersion:    mainStateDocument.CloudInfra.Azure.B.KubernetesVer,
			},
		}
		got, err := fakeClientManaged.GetRAWClusterInfos(storeManaged)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(storeManaged), nil, "managed cluster should be deleted")

		assert.Equal(t, len(mainStateDocument.CloudInfra.Azure.ManagedClusterName), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(storeManaged), nil, "Network should be deleted")

		assert.Equal(t, len(mainStateDocument.CloudInfra.Azure.ResourceGroupName), 0, "resource grp still present")
		// at this moment the file is not present
		_, err := storeManaged.Read()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory still present")
		}
	})
}

func TestHACluster(t *testing.T) {
	mainStateDocument = &storageTypes.StorageDocument{}
	fakeClientHA, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo-ha",
		Region:      "fake",
		Provider:    consts.CloudAzure,
		IsHA:        true,
		NoCP:        7,
		NoDS:        5,
		NoWP:        10,
		K8sDistro:   consts.K8sK3s,
	}, parentLogger, mainStateDocument, ProvideMockClient)

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "demo-ha", consts.ClusterTypeHa)
	_ = storeHA.Connect()

	fakeClientHA.metadata.noCP = 7
	fakeClientHA.metadata.noDS = 5
	fakeClientHA.metadata.noWP = 10

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-data-not-used").NewNetwork(storeHA), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.ResourceGroupName, fakeClientHA.resourceGroup, "resource group not saved")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.VirtualNetworkName, fakeClientHA.clusterName+"-vnet", "virtual net should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.NetCidr, "10.1.0.0/16", "network cidr should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.SubnetName, fakeClientHA.clusterName+"-subnet", "subnet should be created")

		assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.VirtualNetworkID) > 0, "virtual net should be created")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.SubnetID) > 0, "subnet should be created")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.SSHKeyName, "fake-ssh", "sshid must be present")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.SSHUser, "azureuser", "ssh user not set")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-fw-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName, "fake-fw-cp", "firewallID for controlplane absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-fw-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName, "fake-fw-wp", "firewallID for workerplane absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-fw-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName, "fake-fw-lb", "firewallID for loadbalacer absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-fw-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName, "fake-fw-ds", "firewallID for datastore absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID) > 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")
			fakeClientHA.VMType("fake")

			assert.Equal(t, fakeClientHA.NewVM(storeHA, 0), nil, "new vm failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name, "fake-lb", "missmatch of Loadbalancer VM name")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.DiskName, "fake-lb"+"-disk", "missmatch of Loadbalancer disk name")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPName, "fake-lb"+"-pub", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPID) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName, "fake-lb"+"-nic", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PrivateIP, "192.168.1.2", "missmatch of Loadbalancer private ip NIC")

			checkCurrentStateFileHA(t)
		})
		t.Run("Controlplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfControlPlane(fakeClientHA.metadata.noCP, true); err != nil {
				t.Fatalf("Failed to set the controlplane")
			}

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					fakeClientHA.Name(fmt.Sprintf("fake-cp-%d", i))
					fakeClientHA.Role(consts.RoleCp)
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[i], fmt.Sprintf("fake-cp-%d", i), "missmatch of controlplane VM name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Hostnames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.DiskNames[i], fmt.Sprintf("fake-cp-%d-disk", i), "missmatch of controlplane disk name")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[i], fmt.Sprintf("fake-cp-%d-pub", i), "missmatch of controlplane pub ip name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[i]) > 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[i], fmt.Sprintf("fake-cp-%d-nic", i), "missmatch of controlplane nic name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of controlplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})

		t.Run("Datastores", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfDataStore(fakeClientHA.metadata.noDS, true); err != nil {
				t.Fatalf("Failed to set the datastore")
			}

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					fakeClientHA.Role(consts.RoleDs)
					fakeClientHA.Name(fmt.Sprintf("fake-ds-%d", i))
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[i], fmt.Sprintf("fake-ds-%d", i), "missmatch of datastore VM name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.Hostnames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.DiskNames[i], fmt.Sprintf("fake-ds-%d", i)+"-disk", "missmatch of datastore disk name")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames[i], fmt.Sprintf("fake-ds-%d", i)+"-pub", "missmatch of datastore pub ip name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPIDs[i]) > 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[i], fmt.Sprintf("fake-ds-%d", i)+"-nic", "missmatch of datastore nic name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs[i], "192.168.1.2", "missmatch of datastore private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Workplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfWorkerPlane(storeHA, fakeClientHA.metadata.noWP, true); err != nil {
				t.Fatalf("Failed to set the workerplane")
			}

			for i := 0; i < fakeClientHA.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {

					fakeClientHA.Role(consts.RoleWp)
					fakeClientHA.Name(fmt.Sprintf("fake-wp-%d", i))
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[i], fmt.Sprintf("fake-wp-%d", i), "missmatch of workerplane VM name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[i], fmt.Sprintf("fake-wp-%d", i)+"-disk", "missmatch of workerplane disk name")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[i], fmt.Sprintf("fake-wp-%d", i)+"-pub", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[i]) > 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[i], fmt.Sprintf("fake-wp-%d", i)+"-nic", "missmatch of workerplane nic name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.IsCompleted, true, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames

		got := fakeClientHA.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientHA.GetStateFile(storeHA)
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(mainStateDocument)
		assert.DeepEqual(t, string(got), expected)
	})

	t.Run("Get cluster ha", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:          fakeClientHA.clusterName,
				Region:        fakeClientHA.region,
				CloudProvider: consts.CloudAzure,
				ClusterType:   consts.ClusterTypeHa,
				NoWP:          fakeClientHA.noWP,
				NoCP:          fakeClientHA.noCP,
				NoDS:          fakeClientHA.noDS,

				WP: []cloud.VMData{
					{VMSize: "fake-wp-0"}, {VMSize: "fake-wp-1"}, {VMSize: "fake-wp-2"},
					{VMSize: "fake-wp-3"}, {VMSize: "fake-wp-4"}, {VMSize: "fake-wp-5"},
					{VMSize: "fake-wp-6"}, {VMSize: "fake-wp-7"}, {VMSize: "fake-wp-8"},
					{VMSize: "fake-wp-9"},
				},
				CP: []cloud.VMData{
					{VMSize: "fake-cp-0"}, {VMSize: "fake-cp-1"}, {VMSize: "fake-cp-2"},
					{VMSize: "fake-cp-3"}, {VMSize: "fake-cp-4"}, {VMSize: "fake-cp-5"},
					{VMSize: "fake-cp-6"},
				},
				DS: []cloud.VMData{
					{VMSize: "fake-ds-0"}, {VMSize: "fake-ds-1"}, {VMSize: "fake-ds-2"},
					{VMSize: "fake-ds-3"}, {VMSize: "fake-ds-4"},
				},
				LB: cloud.VMData{VMSize: "fake-lb"},

				K8sDistro:  "",
				K8sVersion: mainStateDocument.CloudInfra.Azure.B.KubernetesVer,
			},
		}
		got, err := fakeClientHA.GetRAWClusterInfos(storeHA)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	// explicit clean
	mainStateDocument = &storageTypes.StorageDocument{}

	// use init state firest
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationDelete); err != nil {
			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
	})

	t.Run("Get all counters", func(t *testing.T) {
		var err error
		fakeClientHA.metadata.noCP, err = fakeClientHA.NoOfControlPlane(-1, false)
		assert.Assert(t, err == nil)

		fakeClientHA.metadata.noWP, err = fakeClientHA.NoOfWorkerPlane(storeHA, -1, false)
		assert.Assert(t, err == nil)

		fakeClientHA.metadata.noDS, err = fakeClientHA.NoOfDataStore(-1, false)
		assert.Assert(t, err == nil)
	})

	t.Run("Delete VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelVM(storeHA, 0), nil, "del vm failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name, "", "missmatch of Loadbalancer VM name")
			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.HostName, "", "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.DiskName, "", "missmatch of Loadbalancer disk name")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPName, "", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPID) == 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIP, "", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName, "", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID) == 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PrivateIP, "", "missmatch of Loadbalancer private ip NIC")
			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[i], "", "missmatch of workerplane VM name")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[i], "", "missmatch of workerplane vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[i], "", "missmatch of workerplane disk name")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[i], "", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[i]) == 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[i], "", "missmatch of workerplane pub ip")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[i], "", "missmatch of workerplane nic name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[i], "", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[i], "", "missmatch of controlplane VM name")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Hostnames[i], "", "missmatch of controlplane vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.DiskNames[i], "", "missmatch of controlplane disk name")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[i], "", "missmatch of controlplane pub ip name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[i]) == 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs[i], "", "missmatch of controlplane pub ip")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[i], "", "missmatch of controlplane nic name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[i], "", "missmatch of controlplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[i], "", "missmatch of datastore VM name")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.Hostnames[i], "", "missmatch of datastore vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.DiskNames[i], "", "missmatch of datastore disk name")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames[i], "", "missmatch of datastore pub ip name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPIDs[i]) == 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs[i], "", "missmatch of datastore pub ip")

					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[i], "", "missmatch of datastore nic name")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs[i], "", "missmatch of datastore private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "del firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName, "", "firewallID for controlplane absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName, "", "firewallID for workerplane absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName, "", "firewallID for loadbalacer absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName, "", "firewallID for datastore absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID) == 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.B.SSHUser, "", "ssh user not set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(storeHA), nil, "Network should be deleted")

		assert.Equal(t, mainStateDocument.CloudInfra.Azure.ResourceGroupName, "", "resource group not saved")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.VirtualNetworkName, "", "virtual net should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Azure.SubnetName, "", "subnet should be created")

		assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.VirtualNetworkID) == 0, "virtual net should be created")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Azure.SubnetID) == 0, "subnet should be created")
	})
}
