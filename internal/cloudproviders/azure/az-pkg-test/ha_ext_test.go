package az_pkg_test

import (
	"fmt"
	"github.com/ksctl/ksctl/internal/cloudproviders/azure"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"gotest.tools/v3/assert"
	"testing"
)

func TestHACluster(t *testing.T) {
	var (
		cntCP       = 7
		cntWP       = 10
		cntDS       = 5
		clusterName = "demo-ha"
		regionCode  = "fake"
	)
	mainStateDocumentHA := &storageTypes.StorageDocument{}

	fakeClientHA, _ = azure.NewClient(parentCtx, types.Metadata{
		ClusterName: clusterName,
		Region:      regionCode,
		Provider:    consts.CloudAzure,
		IsHA:        true,
		NoCP:        cntCP,
		NoDS:        cntDS,
		NoWP:        cntWP,
		K8sDistro:   consts.K8sK3s,
	}, parentLogger, mainStateDocumentHA, azure.ProvideClient)

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "demo-ha", consts.ClusterTypeHa)
	_ = storeHA.Connect()

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-data-not-used").NewNetwork(storeHA), nil, "Network should be created")
		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.ResourceGroupName, genResourceGroup(clusterName, "ha"), "resource group not saved")
		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.VirtualNetworkName, clusterName+"-vnet", "virtual net should be created")
		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.NetCidr, "10.1.0.0/16", "network cidr should be created")
		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.SubnetName, clusterName+"-subnet", "subnet should be created")

		assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.VirtualNetworkID) > 0, "virtual net should be created")
		assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.SubnetID) > 0, "subnet should be created")

	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.SSHKeyName, "fake-ssh", "sshid must be present")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.SSHUser, "azureuser", "ssh user not set")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-fw-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName, "fake-fw-cp", "firewallID for controlplane absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-fw-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName, "fake-fw-wp", "firewallID for workerplane absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-fw-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName, "fake-fw-lb", "firewallID for loadbalacer absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-fw-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName, "fake-fw-ds", "firewallID for datastore absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID) > 0, "fw id for datastore missing")
		})

	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")
			fakeClientHA.VMType("fake")

			assert.Equal(t, fakeClientHA.NewVM(storeHA, 0), nil, "new vm failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.Name, "fake-lb", "missmatch of Loadbalancer VM name")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.DiskName, "fake-lb"+"-disk", "missmatch of Loadbalancer disk name")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PublicIPName, "fake-lb"+"-pub", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PublicIPID) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName, "fake-lb"+"-nic", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PrivateIP, "192.168.1.2", "missmatch of Loadbalancer private ip NIC")

		})
		t.Run("Controlplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfControlPlane(cntCP, true); err != nil {
				t.Fatalf("Failed to set the controlplane")
			}

			for i := 0; i < cntCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					fakeClientHA.Name(fmt.Sprintf("fake-cp-%d", i))
					fakeClientHA.Role(consts.RoleCp)
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.Names[i], fmt.Sprintf("fake-cp-%d", i), "missmatch of controlplane VM name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.Hostnames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.DiskNames[i], fmt.Sprintf("fake-cp-%d-disk", i), "missmatch of controlplane disk name")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[i], fmt.Sprintf("fake-cp-%d-pub", i), "missmatch of controlplane pub ip name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[i]) > 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[i], fmt.Sprintf("fake-cp-%d-nic", i), "missmatch of controlplane nic name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of controlplane private ip NIC")

				})
			}
		})

		t.Run("Datastores", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfDataStore(cntDS, true); err != nil {
				t.Fatalf("Failed to set the datastore")
			}

			for i := 0; i < cntDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					fakeClientHA.Role(consts.RoleDs)
					fakeClientHA.Name(fmt.Sprintf("fake-ds-%d", i))
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.Names[i], fmt.Sprintf("fake-ds-%d", i), "missmatch of datastore VM name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.Hostnames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.DiskNames[i], fmt.Sprintf("fake-ds-%d", i)+"-disk", "missmatch of datastore disk name")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PublicIPNames[i], fmt.Sprintf("fake-ds-%d", i)+"-pub", "missmatch of datastore pub ip name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PublicIPIDs[i]) > 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[i], fmt.Sprintf("fake-ds-%d", i)+"-nic", "missmatch of datastore nic name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PrivateIPs[i], "192.168.1.2", "missmatch of datastore private ip NIC")

				})
			}
		})
		t.Run("Workplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfWorkerPlane(storeHA, cntWP, true); err != nil {
				t.Fatalf("Failed to set the workerplane")
			}

			for i := 0; i < cntWP; i++ {
				t.Run("workerplane", func(t *testing.T) {

					fakeClientHA.Role(consts.RoleWp)
					fakeClientHA.Name(fmt.Sprintf("fake-wp-%d", i))
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(storeHA, i), nil, "new vm failed")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.Names[i], fmt.Sprintf("fake-wp-%d", i), "missmatch of workerplane VM name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[i], fmt.Sprintf("fake-wp-%d", i)+"-disk", "missmatch of workerplane disk name")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[i], fmt.Sprintf("fake-wp-%d", i)+"-pub", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[i]) > 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[i], fmt.Sprintf("fake-wp-%d", i)+"-nic", "missmatch of workerplane nic name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of workerplane private ip NIC")

				})
			}

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.IsCompleted, true, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.Hostnames

		got := fakeClientHA.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	t.Run("Get cluster ha", func(t *testing.T) {
		expected := []cloud.AllClusterData{
			{
				Name:            clusterName,
				Region:          regionCode,
				ResourceGrpName: genResourceGroup(clusterName, string(consts.ClusterTypeHa)),
				CloudProvider:   consts.CloudAzure,
				SSHKeyName:      "fake-ssh",
				NetworkName:     clusterName + "-vnet",
				NetworkID:       "XXYY",
				ClusterType:     consts.ClusterTypeHa,
				NoWP:            cntWP,
				NoCP:            cntCP,
				NoDS:            cntDS,

				WP: []cloud.VMData{
					{
						VMSize:       "fake-wp-0",
						VMName:       "fake-wp-0",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-1",
						VMName:       "fake-wp-1",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-2",
						VMName:       "fake-wp-2",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-3",
						VMName:       "fake-wp-3",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-4",
						VMName:       "fake-wp-4",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-5",
						VMName:       "fake-wp-5",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-6",
						VMName:       "fake-wp-6",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-7",
						VMName:       "fake-wp-7",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-8",
						VMName:       "fake-wp-8",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-wp-9",
						VMName:       "fake-wp-9",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-wp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
				},
				CP: []cloud.VMData{
					{
						VMSize:       "fake-cp-0",
						VMName:       "fake-cp-0",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-cp-1",
						VMName:       "fake-cp-1",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-cp-2",
						VMName:       "fake-cp-2",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-cp-3",
						VMName:       "fake-cp-3",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-cp-4",
						VMName:       "fake-cp-4",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-cp-5",
						VMName:       "fake-cp-5",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-cp-6",
						VMName:       "fake-cp-6",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-cp",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
				},
				DS: []cloud.VMData{
					{
						VMSize:       "fake-ds-0",
						VMName:       "fake-ds-0",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-ds",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-ds-1",
						VMName:       "fake-ds-1",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-ds",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-ds-2",
						VMName:       "fake-ds-2",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-ds",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-ds-3",
						VMName:       "fake-ds-3",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-ds",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
					{
						VMSize:       "fake-ds-4",
						VMName:       "fake-ds-4",
						FirewallID:   "XXYY",
						FirewallName: "fake-fw-ds",
						SubnetID:     "XXYY",
						SubnetName:   "demo-ha-subnet",
						PublicIP:     "A.B.C.D",
						PrivateIP:    "192.168.1.2",
					},
				},
				LB: cloud.VMData{
					VMSize:       "fake-lb",
					VMName:       "fake-lb",
					FirewallID:   "XXYY",
					FirewallName: "fake-fw-lb",
					SubnetID:     "XXYY",
					SubnetName:   "demo-ha-subnet",
					PublicIP:     "A.B.C.D",
					PrivateIP:    "192.168.1.2",
				},

				K8sDistro:      consts.K8sK3s,
				K8sVersion:     "fake",
				HAProxyVersion: "3.0",
				EtcdVersion:    "fake",
			},
		}

		{
			// simulate the distro did something
			mainStateDocumentHA.K8sBootstrap = &storageTypes.KubernetesBootstrapState{
				K3s: &storageTypes.StateConfigurationK3s{},
			}

			mainStateDocumentHA.K8sBootstrap.B.EtcdVersion = "fake"
			mainStateDocumentHA.K8sBootstrap.B.HAProxyVersion = "3.0"
			mainStateDocumentHA.K8sBootstrap.K3s.K3sVersion = "fake"
			mainStateDocumentHA.BootstrapProvider = consts.K8sK3s
			if err := storeHA.Write(mainStateDocumentHA); err != nil {
				t.Fatalf("Unable to write the state, Reason: %v", err)
			}
		}
		got, err := fakeClientHA.GetRAWClusterInfos(storeHA)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	{
		// explicit clean
		mainStateDocumentHA = &storageTypes.StorageDocument{}
		fakeClientHA, _ = azure.NewClient(parentCtx, types.Metadata{
			ClusterName: "demo-ha",
			Region:      "fake",
			Provider:    consts.CloudAzure,
			IsHA:        true,
			NoCP:        7,
			NoDS:        5,
			NoWP:        10,
			K8sDistro:   consts.K8sK3s,
		}, parentLogger, mainStateDocumentHA, azure.ProvideClient)
	}

	// use init state firest
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeClientHA.InitState(storeHA, consts.OperationDelete); err != nil {
			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
		}

	})

	t.Run("Get all counters", func(t *testing.T) {
		var err error
		cntCP, err = fakeClientHA.NoOfControlPlane(-1, false)
		assert.Assert(t, err == nil)

		cntWP, err = fakeClientHA.NoOfWorkerPlane(storeHA, -1, false)
		assert.Assert(t, err == nil)

		cntDS, err = fakeClientHA.NoOfDataStore(-1, false)
		assert.Assert(t, err == nil)
	})

	t.Run("Delete VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelVM(storeHA, 0), nil, "del vm failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.Name, "", "missmatch of Loadbalancer VM name")
			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.HostName, "", "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.DiskName, "", "missmatch of Loadbalancer disk name")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PublicIPName, "", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PublicIPID) == 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PublicIP, "", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName, "", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID) == 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.PrivateIP, "", "missmatch of Loadbalancer private ip NIC")
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < cntWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.Names[i], "", "missmatch of workerplane VM name")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[i], "", "missmatch of workerplane vm hostname")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[i], "", "missmatch of workerplane disk name")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[i], "", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[i]) == 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[i], "", "missmatch of workerplane pub ip")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[i], "", "missmatch of workerplane nic name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[i], "", "missmatch of workerplane private ip NIC")

				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < cntCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.Names[i], "", "missmatch of controlplane VM name")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.Hostnames[i], "", "missmatch of controlplane vm hostname")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.DiskNames[i], "", "missmatch of controlplane disk name")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[i], "", "missmatch of controlplane pub ip name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[i]) == 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PublicIPs[i], "", "missmatch of controlplane pub ip")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[i], "", "missmatch of controlplane nic name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[i], "", "missmatch of controlplane private ip NIC")

				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < cntDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.Names[i], "", "missmatch of datastore VM name")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.Hostnames[i], "", "missmatch of datastore vm hostname")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.DiskNames[i], "", "missmatch of datastore disk name")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PublicIPNames[i], "", "missmatch of datastore pub ip name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PublicIPIDs[i]) == 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PublicIPs[i], "", "missmatch of datastore pub ip")

					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[i], "", "missmatch of datastore nic name")
					assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")
					assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.PrivateIPs[i], "", "missmatch of datastore private ip NIC")

				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "del firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName, "", "firewallID for controlplane absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName, "", "firewallID for workerplane absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName, "", "firewallID for loadbalacer absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName, "", "firewallID for datastore absent")
			assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID) == 0, "fw id for datastore missing")
		})

	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.B.SSHUser, "", "ssh user not set")

	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(storeHA), nil, "Network should be deleted")

		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.ResourceGroupName, "", "resource group not saved")
		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.VirtualNetworkName, "", "virtual net should be created")
		assert.Equal(t, mainStateDocumentHA.CloudInfra.Azure.SubnetName, "", "subnet should be created")

		assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.VirtualNetworkID) == 0, "virtual net should be created")
		assert.Assert(t, len(mainStateDocumentHA.CloudInfra.Azure.SubnetID) == 0, "subnet should be created")
	})
}
