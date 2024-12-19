package civo

import (
	"encoding/json"
	"fmt"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"gotest.tools/v3/assert"
	"testing"
)

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

func TestHACluster(t *testing.T) {
	mainStateDocumentHa := &storageTypes.StorageDocument{}

	fakeClientHA, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo-ha",
		Region:      "LON1",
		Provider:    consts.CloudCivo,
		IsHA:        true,
		NoCP:        7,
		NoDS:        5,
		NoWP:        10,
		K8sDistro:   consts.K8sK3s,
	}, parentLogger, mainStateDocumentHa, ProvideClient)

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudCivo, "LON1", "demo-ha", consts.ClusterTypeHa)
	_ = storeHA.Connect()

	fakeClientHA.metadata.noCP = 7
	fakeClientHA.metadata.noDS = 5
	fakeClientHA.metadata.noWP = 10

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-net").NewNetwork(storeHA), nil, "Network should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.NetworkID) > 0, "network id not saved")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.B.SSHID) > 0, "sshid must be present")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Civo.B.SSHUser, "root", "ssh user not set")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Civo.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDControlPlanes) > 0, "firewallID for controlplane absent")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDWorkerNodes) > 0, "firewallID for workerplane absent")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDLoadBalancer) > 0, "firewallID for loadbalancer absent")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDDatabaseNodes) > 0, "firewallID for datastore absent")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb-vm")
			fakeClientHA.VMType("g4s.kube.small")

			assert.Equal(t, fakeClientHA.NewVM(storeHA, 0), nil, "new vm failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.VMID) > 0, "loadbalancer VM id absent")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.PublicIP) > 0, "loadbalancer ipv4 absent")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.PrivateIP) > 0, "loadbalancer private ipv4 absent")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.HostName) > 0, "loadbalancer hostname absent")

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
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.VMIDs[i]) > 0, "controlplane VM id absent")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.PublicIPs[i]) > 0, "controlplane ipv4 absent")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[i]) > 0, "controlplane private ipv4 absent")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.Hostnames[i]) > 0, "controlplane hostname absent")

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

					fakeClientHA.Name(fmt.Sprintf("fake-ds-%d", i))
					fakeClientHA.VMType("g4s.kube.small")
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.VMIDs[i]) > 0, "datastore VM id absent")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.PublicIPs[i]) > 0, "datastore ipv4 absent")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.PrivateIPs[i]) > 0, "datastore private ipv4 absent")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.Hostnames[i]) > 0, "datastore hostname absent")

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
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[i]) > 0, "workerplane VM id absent")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[i]) > 0, "workerplane ipv4 absent")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[i]) > 0, "workerplane private ipv4 absent")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[i]) > 0, "workerplane hostname absent")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, mainStateDocumentHa.CloudInfra.Civo.B.IsCompleted, true, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.Hostnames

		got := fakeClientHA.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientHA.GetStateFile(storeHA)
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(mainStateDocumentHa)
		assert.DeepEqual(t, string(got), expected)
	})

	t.Run("Get cluster ha", func(t *testing.T) {
		_e := cloud.VMData{
			VMSize:     "g4s.kube.small",
			VMID:       "vm-fake",
			FirewallID: "fake-fw",
			PublicIP:   "A.B.C.D",
			PrivateIP:  "192.168.1.2",
		}
		expected := []cloud.AllClusterData{
			{
				Name:          fakeClientHA.clusterName,
				CloudProvider: consts.CloudCivo,
				ClusterType:   consts.ClusterTypeHa,
				Region:        fakeClientHA.region,
				NoWP:          fakeClientHA.noWP,
				NoCP:          fakeClientHA.noCP,
				NoDS:          fakeClientHA.noDS,
				SSHKeyID:      "fake-ssh",
				NetworkID:     "fake-net",
				LB:            _e,
				WP: []cloud.VMData{
					_e, _e, _e, _e,
					_e, _e, _e, _e,
					_e, _e,
				},
				CP: []cloud.VMData{
					_e, _e, _e, _e,
					_e, _e, _e,
				},
				DS: []cloud.VMData{
					_e, _e, _e, _e,
					_e,
				},
				K8sDistro:      consts.K8sK3s,
				K8sVersion:     "fake",
				HAProxyVersion: "3.0",
				EtcdVersion:    "fake",
			},
		}

		{
			// simulate the distro did something
			mainStateDocumentHa.K8sBootstrap = &storageTypes.KubernetesBootstrapState{
				K3s: &storageTypes.StateConfigurationK3s{},
			}

			mainStateDocumentHa.K8sBootstrap.B.EtcdVersion = "fake"
			mainStateDocumentHa.K8sBootstrap.B.HAProxyVersion = "3.0"
			mainStateDocumentHa.K8sBootstrap.K3s.K3sVersion = "fake"
			mainStateDocumentHa.BootstrapProvider = consts.K8sK3s
			if err := storeHA.Write(mainStateDocumentHa); err != nil {
				t.Fatalf("Unable to write the state, Reason: %v", err)
			}
		}
		got, err := fakeClientHA.GetRAWClusterInfos(storeHA)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	{
		// explicit clean
		mainStateDocumentHa = &storageTypes.StorageDocument{}
		fakeClientHA, _ = NewClient(parentCtx, types.Metadata{
			ClusterName: "demo-ha",
			Region:      "LON1",
			Provider:    consts.CloudCivo,
			IsHA:        true,
			NoCP:        7,
			NoDS:        5,
			NoWP:        10,
			K8sDistro:   consts.K8sK3s,
		}, parentLogger, mainStateDocumentHa, ProvideClient)
	}

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
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.VMID) == 0, "loadbalancer VM id absent")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.PublicIP) == 0, "loadbalancer ipv4 absent")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.PrivateIP) == 0, "loadbalancer private ipv4 present")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoLoadBalancer.HostName) == 0, "loadbalancer hostname present")

			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[i]) == 0, "workerplane VM id present")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[i]) == 0, "workerplane ipv4 present")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[i]) == 0, "workerplane private ipv4 present")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[i]) == 0, "workerplane hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.VMIDs[i]) == 0, "controlplane VM id present")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.PublicIPs[i]) == 0, "controlplane ipv4 present")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[i]) == 0, "controlplane private ipv4 present")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoControlPlanes.Hostnames[i]) == 0, "controlplane hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.VMIDs[i]) == 0, "datastore VM id present")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.PublicIPs[i]) == 0, "datastore ipv4 present")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.PrivateIPs[i]) == 0, "datastore private ipv4 present")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.InfoDatabase.Hostnames[i]) == 0, "datastore hostname present")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "del firewall failed")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDControlPlanes) == 0, "firewallID for controlplane present")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDWorkerNodes) == 0, "firewallID for workerplane present")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDLoadBalancer) == 0, "firewallID for loadbalancer present")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.FirewallIDDatabaseNodes) == 0, "firewallID for datastore present")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.B.SSHID) == 0, "sshid still present")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Civo.B.SSHUser, "", "ssh user set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(storeHA), nil, "Network should be deleted")
		assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Civo.NetworkID) == 0, "network id still present")
	})

}
