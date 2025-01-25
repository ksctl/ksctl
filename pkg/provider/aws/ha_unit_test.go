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

package aws

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

	if err := storeHA.Setup(consts.CloudAws, fakeClientHA.state.Region, fakeClientHA.state.ClusterName, consts.ClusterTypeSelfMang); err != nil {
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
	_ = storeHA.Setup(consts.CloudAws, "fake-region", "demo-ha", consts.ClusterTypeSelfMang)
	_ = storeHA.Connect()

	fakeClientHA, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo-ha",
			Region:      "fake-region",
			Provider:    consts.CloudAws,
			SelfManaged: true,
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

		assert.Equal(t, fakeClientHA.clusterType, consts.ClusterTypeSelfMang, "clustertype should be managed")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-data-not-used").NewNetwork(), nil, "Network should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.VpcId, "3456d25f36g474g546", "want %s got %s", "3456d25f36g474g546", fakeClientHA.state.CloudInfra.Aws.VpcId)
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.VpcName, fakeClientHA.ClusterName+"-vpc", "virtual net should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.SubnetNames[0], fakeClientHA.ClusterName+"-subnet0", "subnet should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.SubnetIDs[0], "3456d25f36g474g546", "subnet should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.RouteTableID, "3456d25f36g474g546", "route table should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.GatewayID, "3456d25f36g474g546", "gateway should be created")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(), nil, "ssh key failed")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.SSHKeyName, "fake-ssh", "sshid must be present")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.SSHUser, "ubuntu", "ssh user not set")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-fw-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-fw-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-fw-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-fw-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(), nil, "new firewall failed")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs) > 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")
			fakeClientHA.VMType("fake")

			assert.Equal(t, fakeClientHA.NewVM(0), nil, "new vm failed")

			assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.InstanceID, "test-instance-1234567890", "missmatch of Loadbalancer VM ID")
			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.PublicIP) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.PrivateIP, "192.168.1.2", "missmatch of Loadbalancer private ip NIC")

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

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds[i], "test-instance-1234567890", "missmatch of controlplane VM ID")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.HostNames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of controlplane private ip NIC")

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

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoDatabase.InstanceIds[i], "test-instance-1234567890", "missmatch of datastore VM ID")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoDatabase.HostNames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoDatabase.PrivateIPs[i], "192.168.1.2", "missmatch of datastore private ip NIC")

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

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[i], "test-instance-1234567890", "missmatch of workerplane VM ID")
					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[i], "192.168.1.2", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames

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
		_e := logger.VMData{
			VMSize:     "fake",
			VMID:       "test-instance-1234567890",
			FirewallID: "test-security-group-1234567890",
			SubnetID:   "3456d25f36g474g546",
			SubnetName: "demo-ha-subnet0",
			PublicIP:   "A.B.C.D",
			PrivateIP:  "192.168.1.2",
		}

		// ~adjustments
		fakeClientHA.state.ProvisionerAddons.Cni.Name = "flannel"
		fakeClientHA.state.ProvisionerAddons.Cni.For = consts.K8sK3s

		expected := []logger.ClusterDataForLogging{
			{
				Name:          fakeClientHA.ClusterName,
				Region:        fakeClientHA.Region,
				CloudProvider: consts.CloudAws,
				ClusterType:   consts.ClusterTypeSelfMang,
				SSHKeyName:    "fake-ssh",
				NetworkName:   fakeClientHA.ClusterName + "-vpc",
				NetworkID:     "3456d25f36g474g546",

				NoWP: fakeClientHA.NoWP,
				NoCP: fakeClientHA.NoCP,
				NoDS: fakeClientHA.NoDS,

				WP: []logger.VMData{
					_e, _e, _e, _e,
					_e, _e, _e, _e,
					_e, _e,
				},
				CP: []logger.VMData{
					_e, _e, _e, _e,
					_e, _e, _e,
				},
				DS: []logger.VMData{
					_e, _e, _e, _e,
					_e,
				},
				LB: _e,

				K8sDistro:      consts.K8sK3s,
				K8sVersion:     "fake",
				HAProxyVersion: "3.0",
				EtcdVersion:    "fake",
				Apps:           nil,
				Cni:            "Name: flannel, For: k3s, Version: <nil>, KsctlSpecificComponents: map[]",
			},
		}

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
				Region:      "fake-region",
				Provider:    consts.CloudAws,
				SelfManaged: true,
				NoCP:        7,
				NoDS:        5,
				NoWP:        10,
				K8sDistro:   consts.K8sK3s,
			},
			&statefile.StorageDocument{},
			storeHA,
			ProvideClient)
	}
	t.Run("init state deletion", func(t *testing.T) {

		if err := fakeClientHA.InitState(consts.OperationDelete); err != nil {
			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
		}

		assert.Equal(t, fakeClientHA.clusterType, consts.ClusterTypeSelfMang, "clustertype should be managed")
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

			assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.InstanceID, "", "missmatch of Loadbalancer VM ID")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId) == 0, "missmatch of Loadbalancer nic must be created")
			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.NoWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(i), nil, "del vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[i], "", "missmatch of workerplane VM ID")

					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.NoCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(i), nil, "del vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds[i], "", "missmatch of controlplane VM ID")

					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.NoDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(i), nil, "del vm failed")

					assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.InfoDatabase.InstanceIds[i], "", "missmatch of datastore VM ID")

					assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "del firewall failed")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "new firewall failed")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "new firewall failed")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(), nil, "new firewall failed")

			assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs) == 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(), nil, "ssh key failed")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.B.SSHUser, "", "ssh user not set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(), nil, "Network should be deleted")

		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.VpcId, "", "resource group not saved")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.VpcName, "", "virtual net should be created")
		assert.Equal(t, fakeClientHA.state.CloudInfra.Aws.SubnetIDs[0], "", "subnet should be created")

		assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.RouteTableID) == 0, "route table should be created")
		assert.Assert(t, len(fakeClientHA.state.CloudInfra.Aws.GatewayID) == 0, "gateway should be created")
	})
}
