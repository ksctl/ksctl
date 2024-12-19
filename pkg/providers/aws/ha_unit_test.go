package aws

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

	if err := storeHA.Setup(consts.CloudAws, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeHa); err != nil {
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
		Region:      "fake-region",
		Provider:    consts.CloudAws,
		IsHA:        true,
		NoCP:        7,
		NoDS:        5,
		NoWP:        10,
		K8sDistro:   consts.K8sK3s,
	}, parentLogger, mainStateDocumentHa, ProvideClient)

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAws, "fake-region", "demo-ha", consts.ClusterTypeHa)
	_ = storeHA.Connect()

	fakeClientHA.metadata.noCP = 7
	fakeClientHA.metadata.noDS = 5
	fakeClientHA.metadata.noWP = 10

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-data-not-used").NewNetwork(storeHA), nil, "Network should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.VpcId, "3456d25f36g474g546", "want %s got %s", "3456d25f36g474g546", mainStateDocumentHa.CloudInfra.Aws.VpcId)
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.VpcName, fakeClientHA.clusterName+"-vpc", "virtual net should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.SubnetNames[0], fakeClientHA.clusterName+"-subnet0", "subnet should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.SubnetIDs[0], "3456d25f36g474g546", "subnet should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.RouteTableID, "3456d25f36g474g546", "route table should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.GatewayID, "3456d25f36g474g546", "gateway should be created")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.SSHKeyName, "fake-ssh", "sshid must be present")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.SSHUser, "ubuntu", "ssh user not set")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-fw-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-fw-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-fw-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-fw-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs) > 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")
			fakeClientHA.VMType("fake")

			assert.Equal(t, fakeClientHA.NewVM(storeHA, 0), nil, "new vm failed")

			assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.InstanceID, "test-instance-1234567890", "missmatch of Loadbalancer VM ID")
			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.PublicIP) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.PrivateIP, "192.168.1.2", "missmatch of Loadbalancer private ip NIC")

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

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.InstanceIds[i], "test-instance-1234567890", "missmatch of controlplane VM ID")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.HostNames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of controlplane private ip NIC")

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

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.InstanceIds[i], "test-instance-1234567890", "missmatch of datastore VM ID")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.HostNames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.PrivateIPs[i], "192.168.1.2", "missmatch of datastore private ip NIC")

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

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[i], "test-instance-1234567890", "missmatch of workerplane VM ID")
					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.HostNames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.IsCompleted, false, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.HostNames

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
			VMSize:     "fake",
			VMID:       "test-instance-1234567890",
			FirewallID: "test-security-group-1234567890",
			SubnetID:   "3456d25f36g474g546",
			SubnetName: "demo-ha-subnet0",
			PublicIP:   "A.B.C.D",
			PrivateIP:  "192.168.1.2",
		}
		expected := []cloud.AllClusterData{
			{
				Name:          fakeClientHA.clusterName,
				Region:        fakeClientHA.region,
				CloudProvider: consts.CloudAws,
				ClusterType:   consts.ClusterTypeHa,
				SSHKeyName:    "fake-ssh",
				NetworkName:   fakeClientHA.clusterName + "-vpc",
				NetworkID:     "3456d25f36g474g546",

				NoWP: fakeClientHA.metadata.noWP,
				NoCP: fakeClientHA.metadata.noCP,
				NoDS: fakeClientHA.metadata.noDS,

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
				LB: _e,

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
			Region:      "fake-region",
			Provider:    consts.CloudAws,
			IsHA:        true,
			NoCP:        7,
			NoDS:        5,
			NoWP:        10,
			K8sDistro:   consts.K8sK3s,
		}, parentLogger, mainStateDocumentHa, ProvideClient)
	}
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

			assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.InstanceID, "", "missmatch of Loadbalancer VM ID")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId) == 0, "missmatch of Loadbalancer nic must be created")
			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[i], "", "missmatch of workerplane VM ID")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.InstanceIds[i], "", "missmatch of controlplane VM ID")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.InstanceIds[i], "", "missmatch of datastore VM ID")

					assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "del firewall failed")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs) == 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.B.SSHUser, "", "ssh user not set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(storeHA), nil, "Network should be deleted")

		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.VpcId, "", "resource group not saved")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.VpcName, "", "virtual net should be created")
		assert.Equal(t, mainStateDocumentHa.CloudInfra.Aws.SubnetIDs[0], "", "subnet should be created")

		assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.RouteTableID) == 0, "route table should be created")
		assert.Assert(t, len(mainStateDocumentHa.CloudInfra.Aws.GatewayID) == 0, "gateway should be created")
	})
}
