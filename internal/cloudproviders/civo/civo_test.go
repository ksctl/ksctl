package civo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/pkg/resources/controllers/cloud"

	"github.com/ksctl/ksctl/internal/storage/types"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"gotest.tools/assert"
)

var (
	fakeClientHA *CivoProvider
	storeHA      resources.StorageFactory

	fakeClientManaged *CivoProvider
	storeManaged      resources.StorageFactory

	fakeClientVars *CivoProvider
	storeVars      resources.StorageFactory

	dir = fmt.Sprintf("%s ksctl-civo-test", os.TempDir())
)

func TestMain(m *testing.M) {

	func() {

		fakeClientVars, _ = ReturnCivoStruct(resources.Metadata{
			ClusterName:  "demo",
			Region:       "LON1",
			Provider:     consts.CloudCivo,
			IsHA:         true,
			LogVerbosity: -1,
			LogWritter:   os.Stdout,
		}, &types.StorageDocument{}, ProvideMockCivoClient)

		storeVars = localstate.InitStorage(-1, os.Stdout)
		_ = storeVars.Setup(consts.CloudCivo, "LON1", "demo", consts.ClusterTypeHa)
		_ = storeVars.Connect(context.TODO())
	}()

	// setup temporary folder
	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-civo-test"); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func TestIsValidK8sVersion(t *testing.T) {
	ver, _ := fakeClientVars.client.ListAvailableKubernetesVersions()
	for _, vver := range ver {
		t.Log(vver)
	}
}

func TestCivoProvider_InitState(t *testing.T) {

	t.Run("Create state", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, consts.OperationStateCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeClientVars.Name("fake").NewNetwork(storeVars), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		mainStateDocument.CloudInfra.Civo.B.IsCompleted = true
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, true, "cluster should not be completed")

		if err := fakeClientVars.InitState(storeVars, consts.OperationStateCreate); err != nil {
			t.Fatalf("Unable to resume state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Get request", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, consts.OperationStateGet); err != nil {
			t.Fatalf("Unable to get state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Delete request", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, consts.OperationStateDelete); err != nil {
			t.Fatalf("Unable to Delete state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Invalid request", func(t *testing.T) {

		if err := fakeClientVars.InitState(storeVars, "test"); err == nil {
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
		token := fetchAPIKey(storeVars)
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
		if retApps := fakeClientVars.Application(apps); retApps {
			t.Fatalf("application shouldn't be external flag")
		}
		if fakeClientVars.metadata.apps != setVal {
			t.Fatalf("apps dont match Expected `%s` but got `%s`", apps, setVal)
		}
	}
}

// Test for the Noof WP and setter and getter
func TestCivoProvider_NoOfControlPlane(t *testing.T) {
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

func TestCivoProvider_NoOfDataStore(t *testing.T) {
	var no int
	var err error

	no, err = fakeClientVars.NoOfDataStore(-1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized datastore array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfDataStore(0, true)
	// it should return error
	if err == nil {
		t.Fatalf("setter should fail on when no < 1 datastore provided_no: %d", 1)
	}

	_, err = fakeClientVars.NoOfDataStore(5, true)
	// it should return error
	if err != nil {
		t.Fatalf("setter should not fail on when n >= 1 datastore err: %v", err)
	}

	no, err = fakeClientVars.NoOfDataStore(-1, false)
	if no != 5 {
		t.Fatalf("Getter failed to get updated no of datastore array got no: %d and err: %v", no, err)
	}
}

func TestCivoProvider_NoOfWorkerPlane(t *testing.T) {
	var no int
	var err error

	no, err = fakeClientVars.NoOfWorkerPlane(storeVars, -1, false)
	if no != -1 || err == nil {
		t.Fatalf("Getter failed on unintalized workerplane array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(storeVars, 2, true)
	// it shouldn't return err
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("setter should not fail on when no >= 0 workerplane provided_no: %d, err: %v", 2, err)
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
}

func TestVMType(t *testing.T) {
	if ret := fakeClientVars.VMType("g4s.kube.small"); ret == nil {
		t.Fatalf("returned nil for valid vm type")
	}
	vm := <-fakeClientVars.chVMType

	if vm != "g4s.kube.small" {
		t.Fatalf("Correct assignment missing")
	}

	if ret := fakeClientVars.VMType(""); ret != nil {
		t.Fatalf("returned interface for invalid vm type")
	}

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
		"NYC1": nil,
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
		"1.27.4",
		"1.27.1",
		"1.28",
	}

	for i := 0; i < len(forTesting); i++ {
		var ver string = forTesting[i]
		if i < 2 {
			if ret := fakeClientVars.Version(ver); ret == nil {
				t.Fatalf("returned nil for valid version")
			}
			if ver+"-k3s1" != fakeClientVars.metadata.k8sVersion {
				t.Fatalf("set value is not equal to input value")
			}
		} else {
			if ret := fakeClientVars.Version(ver); ret != nil {
				t.Fatalf("returned interface for invalid version")
			}
		}
	}

	if ret := fakeClientVars.Version(""); ret == nil {
		t.Fatalf("returned nil for valid version")
	}
	if "1.26.4-k3s1" != fakeClientVars.metadata.k8sVersion {
		t.Fatalf("set value is not equal to input value")
	}
}

func TestCni(t *testing.T) {
	testCases := map[string]bool{
		string(consts.CNICilium):  false,
		string(consts.CNIFlannel): false,
		string(consts.CNIKubenet): true,
		"abcd":                    true,
	}

	for k, v := range testCases {
		got := fakeClientVars.CNI(k)
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

func TestDeleteVarCluster(t *testing.T) {
	if err := storeVars.DeleteCluster(); err != nil {
		t.Fatal(err)
	}
}

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(consts.CloudCivo, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeMang); err != nil {
		t.Fatal(err)
	}
	read, err := storeManaged.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, mainStateDocument, read)
}

func checkCurrentStateFileHA(t *testing.T) {

	if err := storeHA.Setup(consts.CloudCivo, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeHa); err != nil {
		t.Fatal(err)
	}
	read, err := storeHA.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, mainStateDocument, read)
}

func TestManagedCluster(t *testing.T) {

	func() {
		fakeClientManaged, _ = ReturnCivoStruct(resources.Metadata{
			ClusterName:  "demo-managed",
			Region:       "LON1",
			Provider:     consts.CloudCivo,
			LogVerbosity: -1,
			LogWritter:   os.Stdout,
		}, &types.StorageDocument{}, ProvideMockCivoClient)

		storeManaged = localstate.InitStorage(-1, os.Stdout)
		_ = storeManaged.Setup(consts.CloudCivo, "LON1", "demo-managed", consts.ClusterTypeMang)
		_ = storeManaged.Connect(context.TODO())

	}()

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientManaged.InitState(storeManaged, consts.OperationStateCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeMang, "clustertype should be managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeManaged.Read()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.Name("fake-net").NewNetwork(storeManaged), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.NetworkID) > 0, "network id not saved")

		checkCurrentStateFile(t)
	})

	t.Run("Create managed cluster", func(t *testing.T) {

		fakeClientManaged.CNI("cilium")
		fakeClientManaged.Application("abcd")

		assert.Equal(t, fakeClientManaged.Name("fake").VMType("g4s.kube.small").NewManagedCluster(storeManaged, 5), nil, "managed cluster should be created")

		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, true, "cluster should not be completed")

		assert.Equal(t, mainStateDocument.CloudInfra.Civo.NoManagedNodes, 5)
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.KubernetesDistro, string(consts.K8sK3s))
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.KubernetesVer, fakeClientManaged.metadata.k8sVersion)
		assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.ManagedClusterID) > 0, "Managed clusterID not saved")

		_, err := storeManaged.Read()
		if os.IsNotExist(err) {
			t.Fatalf("kubeconfig should not be absent")
		}
		checkCurrentStateFile(t)
	})

	t.Run("Get cluster managed", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:     fakeClientManaged.clusterName,
				Provider: consts.CloudCivo,
				Type:     consts.ClusterTypeMang,
				Region:   fakeClientManaged.region,
				NoMgt:    mainStateDocument.CloudInfra.Civo.NoManagedNodes,

				K8sDistro:  consts.K8sK3s,
				K8sVersion: mainStateDocument.CloudInfra.Civo.B.KubernetesVer,
			},
		}
		got, err := GetRAWClusterInfos(storeManaged, resources.Metadata{LogWritter: os.Stdout, LogVerbosity: -1})
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Delete managed cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelManagedCluster(storeManaged), nil, "managed cluster should be deleted")

		assert.Equal(t, len(mainStateDocument.CloudInfra.Civo.ManagedClusterID), 0, "managed cluster id still present")

		checkCurrentStateFile(t)
	})

	t.Run("Delete Network cluster", func(t *testing.T) {
		assert.Equal(t, fakeClientManaged.DelNetwork(storeManaged), nil, "Network should be deleted")

		assert.Equal(t, len(mainStateDocument.CloudInfra.Civo.NetworkID), 0, "network id still present")
		// at this moment the file is not present
		_, err := storeManaged.Read()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory still present")
		}
	})

}

func TestHACluster(t *testing.T) {
	func() {
		fakeClientHA, _ = ReturnCivoStruct(resources.Metadata{
			ClusterName:  "demo-ha",
			Region:       "LON1",
			Provider:     consts.CloudCivo,
			IsHA:         true,
			LogVerbosity: -1,
			LogWritter:   os.Stdout,
			NoCP:         7,
			NoDS:         5,
			NoWP:         10,
			K8sDistro:    consts.K8sK3s,
		}, &types.StorageDocument{}, ProvideMockCivoClient)

		storeHA = localstate.InitStorage(-1, os.Stdout)
		_ = storeHA.Setup(consts.CloudCivo, "LON1", "demo-ha", consts.ClusterTypeHa)
		_ = storeHA.Connect(context.TODO())

	}()
	fakeClientHA.metadata.noCP = 7
	fakeClientHA.metadata.noDS = 5
	fakeClientHA.metadata.noWP = 10

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationStateCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if os.IsExist(err) {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-net").NewNetwork(storeHA), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.NetworkID) > 0, "network id not saved")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.B.SSHID) > 0, "sshid must be present")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.SSHUser, "root", "ssh user not set")

		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes) > 0, "firewallID for controlplane absent")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes) > 0, "firewallID for workerplane absent")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer) > 0, "firewallID for loadbalancer absent")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes) > 0, "firewallID for datastore absent")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb-vm")
			fakeClientHA.VMType("g4s.kube.small")

			assert.Equal(t, fakeClientHA.NewVM(storeHA, 0), nil, "new vm failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.VMID) > 0, "loadbalancer VM id absent")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PublicIP) > 0, "loadbalancer ipv4 absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PrivateIP) > 0, "loadbalancer private ipv4 absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.HostName) > 0, "loadbalancer hostname absent")

			checkCurrentStateFileHA(t)
		})
		t.Run("Controlplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfControlPlane(fakeClientHA.metadata.noCP, true); err != nil {
				t.Fatalf("Failed to set the controlplane, err: %v", err)
			}

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					fakeClientHA.Name(fmt.Sprintf("fake-cp-%d", i))
					fakeClientHA.VMType("g4s.kube.small")
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs[i]) > 0, "controlplane VM id absent")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs[i]) > 0, "controlplane ipv4 absent")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[i]) > 0, "controlplane private ipv4 absent")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames[i]) > 0, "controlplane hostname absent")

					checkCurrentStateFileHA(t)
				})
			}
		})

		t.Run("Datastores", func(t *testing.T) {
			// NOTE: the noDS is set to 1 becuase current implementation is only for single datastore
			// TODO: use the 1 as limit

			fakeClientHA.metadata.noDS = 1

			if _, err := fakeClientHA.NoOfDataStore(fakeClientHA.metadata.noDS, true); err != nil {
				t.Fatalf("Failed to set the datastore")
			}

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					fakeClientHA.Name(fmt.Sprintf("fake-ds-%d", i))
					fakeClientHA.VMType("g4s.kube.small")
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs[i]) > 0, "datastore VM id absent")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs[i]) > 0, "datastore ipv4 absent")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs[i]) > 0, "datastore private ipv4 absent")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames[i]) > 0, "datastore hostname absent")

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

					fakeClientHA.Name(fmt.Sprintf("fake-wp-%d", i))
					fakeClientHA.Role(consts.RoleWp)
					fakeClientHA.VMType("g4s.kube.small")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[i]) > 0, "workerplane VM id absent")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[i]) > 0, "workerplane ipv4 absent")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[i]) > 0, "workerplane private ipv4 absent")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[i]) > 0, "workerplane hostname absent")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.IsCompleted, true, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames

		got := fakeClientHA.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientHA.GetStateFile(storeHA)
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(mainStateDocument)
		assert.DeepEqual(t, string(got), expected)
	})

	// TODO: check for the Passing the state to the kubernetes distribution function GetStateForHACluster

	t.Run("Get cluster ha", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			cloud.AllClusterData{
				Name:       fakeClientHA.clusterName,
				Provider:   consts.CloudCivo,
				Type:       consts.ClusterTypeHa,
				Region:     fakeClientHA.region,
				NoWP:       fakeClientHA.noWP,
				NoCP:       fakeClientHA.noCP,
				NoDS:       fakeClientHA.noDS,
				K8sDistro:  consts.K8sK3s,
				K8sVersion: mainStateDocument.CloudInfra.Civo.B.KubernetesVer,
			},
		}
		got, err := GetRAWClusterInfos(storeHA, resources.Metadata{LogWritter: os.Stdout, LogVerbosity: -1})
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	// explicit clean
	mainStateDocument = &types.StorageDocument{}

	// use init state firest
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationStateDelete); err != nil {
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
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.VMID) == 0, "loadbalancer VM id absent")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PublicIP) == 0, "loadbalancer ipv4 absent")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PrivateIP) == 0, "loadbalancer private ipv4 present")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.HostName) == 0, "loadbalancer hostname present")

			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[i]) == 0, "workerplane VM id present")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[i]) == 0, "workerplane ipv4 present")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[i]) == 0, "workerplane private ipv4 present")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[i]) == 0, "workerplane hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs[i]) == 0, "controlplane VM id present")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs[i]) == 0, "controlplane ipv4 present")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[i]) == 0, "controlplane private ipv4 present")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames[i]) == 0, "controlplane hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs[i]) == 0, "datastore VM id present")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs[i]) == 0, "datastore ipv4 present")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs[i]) == 0, "datastore private ipv4 present")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames[i]) == 0, "datastore hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "del firewall failed")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes) == 0, "firewallID for controlplane present")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes) == 0, "firewallID for workerplane present")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer) == 0, "firewallID for loadbalancer present")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes) == 0, "firewallID for datastore present")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.B.SSHID) == 0, "sshid still present")
		assert.Equal(t, mainStateDocument.CloudInfra.Civo.B.SSHUser, "", "ssh user set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(storeHA), nil, "Network should be deleted")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Civo.NetworkID) == 0, "network id still present")
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
		actual, err := fakeClientVars.GetSecretTokens(nil) // non need to pass as we are providing the env var
		assert.NilError(t, err, "unable to get the secret token from the client")
		assert.DeepEqual(t, actual, expected)
	})
}
