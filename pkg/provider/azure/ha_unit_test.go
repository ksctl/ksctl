// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	localstate "github.com/ksctl/ksctl/v2/pkg/storage/host"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
	"gotest.tools/v3/assert"
)

func checkCurrentStateFileHA(t *testing.T) {

	if err := storeHA.Setup(consts.CloudAzure, fakeClientHA.state.Region, fakeClientHA.state.ClusterName, consts.ClusterTypeSelfMang); err != nil {
		t.Fatal(err)
	}
	read, err := storeHA.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, fakeClientHA.state, read)
}

func TestHACluster(t *testing.T) {
	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "demo-ha", consts.ClusterTypeSelfMang)
	_ = storeHA.Connect()

	fakeClientHA, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo-ha",
			Region:      "fake",
			Provider:    consts.CloudAzure,
			ClusterType: consts.ClusterTypeSelfMang,
			NoCP:        7,
			NoDS:        5,
			NoWP:        10,
			K8sDistro:   consts.K8sK3s,
		},
		&statefile.StorageDocument{},
		storeHA,
		ProvideClient,
	)

	fakeClientHA.NoCP = 7
	fakeClientHA.NoDS = 5
	fakeClientHA.NoWP = 10

	t.Run("init state", func(t *testing.T) {

		if err := fakeClientHA.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, fakeClientHA.ClusterType, consts.ClusterTypeSelfMang, "clustertype should be managed")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-data-not-used").NewNetwork(), nil, "Network should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.ResourceGroupName, fakeClientHA.resourceGroup, "resource group not saved")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.VirtualNetworkName, fakeClientHA.ClusterName+"-vnet", "virtual net should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.NetCidr, "10.1.0.0/16", "network cidr should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.SubnetName, fakeClientHA.ClusterName+"-subnet", "subnet should be created")

		assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.VirtualNetworkID) > 0, "virtual net should be created")
		assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.SubnetID) > 0, "subnet should be created")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(), nil, "ssh key failed")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.SSHKeyName, "fake-ssh", "sshid must be present")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.SSHUser, "azureuser", "ssh user not set")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-fw-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName, "fake-fw-cp", "firewallID for controlplane absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-fw-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName, "fake-fw-wp", "firewallID for workerplane absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-fw-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName, "fake-fw-lb", "firewallID for loadbalacer absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-fw-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName, "fake-fw-ds", "firewallID for datastore absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID) > 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")
			fakeClientHA.VMType("fake")

			assert.Equal(t, fakeClientHA.NewVM(0), nil, "new vm failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.Name, "fake-lb", "missmatch of Loadbalancer VM name")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.DiskName, "fake-lb"+"-disk", "missmatch of Loadbalancer disk name")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPName, "fake-lb"+"-pub", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPID) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName, "fake-lb"+"-nic", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PrivateIP, "192.168.1.2", "missmatch of Loadbalancer private ip NIC")

			checkCurrentStateFileHA(t)
		})
		t.Run("Controlplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfControlPlane(fakeClientHA.NoCP, true); err != nil {
				t.Fatalf("Failed to set the controlplane")
			}

			for i := 0; i < fakeClientHA.NoCP; i++ {
				t.Run("controlplane", func(t *testing.T) {

					fakeClientHA.Name(fmt.Sprintf("fake-cp-%d", i))
					fakeClientHA.Role(consts.RoleCp)
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(i), nil, "new vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.Names[i], fmt.Sprintf("fake-cp-%d", i), "missmatch of controlplane VM name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.Hostnames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.DiskNames[i], fmt.Sprintf("fake-cp-%d-disk", i), "missmatch of controlplane disk name")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[i], fmt.Sprintf("fake-cp-%d-pub", i), "missmatch of controlplane pub ip name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[i]) > 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[i], fmt.Sprintf("fake-cp-%d-nic", i), "missmatch of controlplane nic name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of controlplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})

		t.Run("Datastores", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfDataStore(fakeClientHA.NoDS, true); err != nil {
				t.Fatalf("Failed to set the datastore")
			}

			for i := 0; i < fakeClientHA.NoDS; i++ {
				t.Run("datastore", func(t *testing.T) {

					fakeClientHA.Role(consts.RoleDs)
					fakeClientHA.Name(fmt.Sprintf("fake-ds-%d", i))
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(i), nil, "new vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.Names[i], fmt.Sprintf("fake-ds-%d", i), "missmatch of datastore VM name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.Hostnames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.DiskNames[i], fmt.Sprintf("fake-ds-%d", i)+"-disk", "missmatch of datastore disk name")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PublicIPNames[i], fmt.Sprintf("fake-ds-%d", i)+"-pub", "missmatch of datastore pub ip name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PublicIPIDs[i]) > 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[i], fmt.Sprintf("fake-ds-%d", i)+"-nic", "missmatch of datastore nic name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PrivateIPs[i], "192.168.1.2", "missmatch of datastore private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Workplanes", func(t *testing.T) {

			if _, err := fakeClientHA.NoOfWorkerPlane(fakeClientHA.NoWP, true); err != nil {
				t.Fatalf("Failed to set the workerplane")
			}

			for i := 0; i < fakeClientHA.NoWP; i++ {
				t.Run("workerplane", func(t *testing.T) {

					fakeClientHA.Role(consts.RoleWp)
					fakeClientHA.Name(fmt.Sprintf("fake-wp-%d", i))
					fakeClientHA.VMType("fake")

					assert.Equal(t, fakeClientHA.NewVM(i), nil, "new vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.Names[i], fmt.Sprintf("fake-wp-%d", i), "missmatch of workerplane VM name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[i], fmt.Sprintf("fake-wp-%d", i)+"-disk", "missmatch of workerplane disk name")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[i], fmt.Sprintf("fake-wp-%d", i)+"-pub", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[i]) > 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[i], fmt.Sprintf("fake-wp-%d", i)+"-nic", "missmatch of workerplane nic name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.IsCompleted, true, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames

		got := fakeClientHA.GetHostNameAllWorkerNode()
		assert.DeepEqual(t, got, expected)
	})

	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClientHA.GetStateFile()
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(fakeClientHA.state)
		assert.DeepEqual(t, string(got), expected)
	})

	t.Run("Get cluster ha", func(t *testing.T) {
		expected := []logger.ClusterDataForLogging{
			{
				Name:            fakeClientHA.ClusterName,
				Region:          fakeClientHA.Region,
				ResourceGrpName: generateResourceGroupName(fakeClientHA.ClusterName, string(consts.ClusterTypeSelfMang)),
				CloudProvider:   consts.CloudAzure,
				SSHKeyName:      "fake-ssh",
				NetworkName:     fakeClientHA.ClusterName + "-vnet",
				NetworkID:       "XXYY",
				ClusterType:     consts.ClusterTypeSelfMang,
				NoWP:            fakeClientHA.NoWP,
				NoCP:            fakeClientHA.NoCP,
				NoDS:            fakeClientHA.NoDS,

				WP: []logger.VMData{
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
				CP: []logger.VMData{
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
				DS: []logger.VMData{
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
				LB: logger.VMData{
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
				Apps:           nil,
				Cni:            "Name: flannel, For: k3s, Version: <nil>, KsctlSpecificComponents: map[]",
			},
		}
		// ~adjustments
		fakeClientHA.state.ProvisionerAddons.Cni.Name = "flannel"
		fakeClientHA.state.ProvisionerAddons.Cni.For = consts.K8sK3s

		{
			// simulate the distro did something
			fakeClientHA.state.K8sBootstrap = &statefile.KubernetesBootstrapState{
				K3s: &statefile.StateConfigurationK3s{},
			}

			fakeClientHA.state.Versions.Etcd = utilities.Ptr("fake")
			fakeClientHA.state.Versions.HAProxy = utilities.Ptr("3.0")
			fakeClientHA.state.Versions.K3s = utilities.Ptr("fake")
			fakeClientHA.state.BootstrapProvider = consts.K8sK3s
			if err := storeHA.Write(fakeClientHA.state); err != nil {
				t.Fatalf("Unable to write the state, Reason: %v", err)
			}
		}
		got, err := fakeClientHA.GetRAWClusterInfos()
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	{
		// explicit clean
		fakeClientHA, _ = NewClient(
			parentCtx,
			parentLogger,
			controller.Metadata{
				ClusterName: "demo-ha",
				Region:      "fake",
				Provider:    consts.CloudAzure,
				ClusterType: consts.ClusterTypeSelfMang,
				NoCP:        7,
				NoDS:        5,
				NoWP:        10,
				K8sDistro:   consts.K8sK3s,
			},
			&statefile.StorageDocument{},
			storeHA,
			ProvideClient,
		)
	}

	// use init state firest
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeClientHA.InitState(consts.OperationDelete); err != nil {
			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
		}

		assert.Equal(t, fakeClientHA.ClusterType, consts.ClusterTypeSelfMang, "clustertype should be managed")
	})

	t.Run("Get all counters", func(t *testing.T) {
		var err error
		fakeClientHA.NoCP, err = fakeClientHA.NoOfControlPlane(-1, false)
		assert.Assert(t, err == nil)

		fakeClientHA.NoWP, err = fakeClientHA.NoOfWorkerPlane(-1, false)
		assert.Assert(t, err == nil)

		fakeClientHA.NoDS, err = fakeClientHA.NoOfDataStore(-1, false)
		assert.Assert(t, err == nil)
	})

	t.Run("Delete VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelVM(0), nil, "del vm failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.Name, "", "missmatch of Loadbalancer VM name")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.HostName, "", "missmatch of Loadbalancer vm hostname")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.DiskName, "", "missmatch of Loadbalancer disk name")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPName, "", "missmatch of Loadbalancer pub ip name")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPID) == 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PublicIP, "", "missmatch of Loadbalancer pub ip")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName, "", "missmatch of Loadbalancer nic name")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID) == 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.PrivateIP, "", "missmatch of Loadbalancer private ip NIC")
			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.NoWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(i), nil, "del vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.Names[i], "", "missmatch of workerplane VM name")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[i], "", "missmatch of workerplane vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[i], "", "missmatch of workerplane disk name")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[i], "", "missmatch of workerplane pub ip name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[i]) == 0, "missmatch of workerplane pub ip id must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[i], "", "missmatch of workerplane pub ip")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[i], "", "missmatch of workerplane nic name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[i], "", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.NoCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(i), nil, "del vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.Names[i], "", "missmatch of controlplane VM name")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.Hostnames[i], "", "missmatch of controlplane vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.DiskNames[i], "", "missmatch of controlplane disk name")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[i], "", "missmatch of controlplane pub ip name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[i]) == 0, "missmatch of controlplane pub ip id must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PublicIPs[i], "", "missmatch of controlplane pub ip")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[i], "", "missmatch of controlplane nic name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[i], "", "missmatch of controlplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.NoDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(i), nil, "del vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.Names[i], "", "missmatch of datastore VM name")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.Hostnames[i], "", "missmatch of datastore vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.DiskNames[i], "", "missmatch of datastore disk name")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PublicIPNames[i], "", "missmatch of datastore pub ip name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PublicIPIDs[i]) == 0, "missmatch of datastore pub ip id must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PublicIPs[i], "", "missmatch of datastore pub ip")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[i], "", "missmatch of datastore nic name")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.PrivateIPs[i], "", "missmatch of datastore private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "del firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName, "", "firewallID for controlplane absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName, "", "firewallID for workerplane absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName, "", "firewallID for loadbalacer absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "new firewall failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName, "", "firewallID for datastore absent")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID) == 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(), nil, "ssh key failed")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.B.SSHUser, "", "ssh user not set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(), nil, "Network should be deleted")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.ResourceGroupName, "", "resource group not saved")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.VirtualNetworkName, "", "virtual net should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Azure.SubnetName, "", "subnet should be created")

		assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.VirtualNetworkID) == 0, "virtual net should be created")
		assert.Assert(t, len(fakeClientHA.state.CloudInfra.Azure.SubnetID) == 0, "subnet should be created")
	})
}
