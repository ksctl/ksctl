package aws

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"testing"

// 	localstate "github.com/kubesimplify/ksctl/internal/storage/local"
// 	"github.com/kubesimplify/ksctl/pkg/resources"
// 	"github.com/kubesimplify/ksctl/pkg/utils"
// 	"github.com/kubesimplify/ksctl/pkg/utils/consts"
// 	"gotest.tools/assert"
// )

// var (
// 	demoClient *resources.KsctlClient
// 	fakeaws    *AwsProvider
// 	dir        = fmt.Sprintf("%s/ksctl-aws-test", os.TempDir())
// )

// func TestMain(m *testing.M) {
// 	demoClient = &resources.KsctlClient{}
// 	demoClient.Metadata.ClusterName = "fake"
// 	demoClient.Metadata.Region = "fake"
// 	demoClient.Metadata.Provider = consts.CloudAws
// 	demoClient.Metadata.LogVerbosity = -1
// 	demoClient.Metadata.LogWritter = os.Stdout
// 	demoClient.Cloud, _ = ReturnAwsStruct(demoClient.Metadata, ProvideMockClient)

// 	fakeaws, _ = ReturnAwsStruct(demoClient.Metadata, ProvideMockClient)

// 	demoClient.Storage = localstate.InitStorage()
// 	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
// 	awsHa := utils.GetPath(consts.UtilClusterPath, consts.CloudAws, consts.ClusterTypeHa)

// 	if err := os.MkdirAll(awsHa, 0755); err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Created tmp directories")
// 	exitVal := m.Run()

// 	fmt.Println("Cleanup..")
// 	if err := os.RemoveAll(dir); err != nil {
// 		panic(err)
// 	}

// 	os.Exit(exitVal)
// }

// // func TestInitState(t *testing.T) {

// // 	//t.Run("Create state", func(t *testing.T) {
// // 	//
// // 	if err := fakeaws.InitState(demoClient.Storage, consts.OperationStateCreate); err != nil {
// // 		t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
// // 	}
// // 	fmt.Println("above sucess")
// // 	//
// // 	//	assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
// // 	//	assert.Equal(t, clusterDirName, fakeaws.clusterName+" "+fakeaws.vpc+" "+fakeaws.region, "clusterdir not equal")
// // 	//	assert.Equal(t, awsCloudState.IsCompleted, false, "cluster should not be completed")
// // 	//	assert.Equal(t, fakeaws.Name("fake-net").NewNetwork(demoClient.Storage), nil, "Network should be created")
// // 	//	assert.Equal(t, awsCloudState.IsCompleted, false, "cluster should not be completed")
// // 	//})

// // 	t.Run("Try to resume", func(t *testing.T) {
// // 		awsCloudState.IsCompleted = true
// // 		assert.Equal(t, awsCloudState.IsCompleted, true, "cluster should not be completed")

// // 		if err := fakeaws.InitState(demoClient.Storage, consts.OperationStateCreate); err != nil {
// // 			t.Fatalf("Unable to resume state, Reason: %v", err)
// // 		}
// // 	})

// // 	t.Run("try to Trigger Get request", func(t *testing.T) {

// // 		if err := fakeaws.InitState(demoClient.Storage, consts.OperationStateGet); err != nil {
// // 			t.Fatalf("Unable to get state, Reason: %v", err)
// // 		}
// // 	})

// // 	t.Run("try to Trigger Delete request", func(t *testing.T) {

// // 		if err := fakeaws.InitState(demoClient.Storage, consts.OperationStateDelete); err != nil {
// // 			t.Fatalf("Unable to Delete state, Reason: %v", err)
// // 		}
// // 	})

// // 	t.Run("try to Trigger Invalid request", func(t *testing.T) {

// // 		if err := fakeaws.InitState(demoClient.Storage, "test"); err == nil {
// // 			t.Fatalf("Expected error but not got: %v", err)
// // 		}
// // 	})
// // }

// // func TestConsts(t *testing.T) {
// // 	assert.Equal(t, KUBECONFIG_FILE_NAME, "kubeconfig", "kubeconfig file")
// // 	assert.Equal(t, STATE_FILE_NAME, "cloud-state.json", "cloud state file")

// // 	assert.Equal(t, FILE_PERM_CLUSTER_STATE, os.FileMode(0640), "state file permission mismatch")
// // 	assert.Equal(t, FILE_PERM_CLUSTER_DIR, os.FileMode(0750), "cluster dir permission mismatch")
// // 	assert.Equal(t, FILE_PERM_CLUSTER_KUBECONFIG, os.FileMode(0755), "kubeconfig file permission mismatch")
// // }

// // func TestNoOfControlPlane(t *testing.T) {
// // 	var no int
// // 	var err error
// // 	no, err = demoClient.Cloud.NoOfControlPlane(-1, false)
// // 	if no != -1 || err != nil {
// // 		t.Fatalf("Getter failed on unintalized controlplanes array got no: %d and err: %v", no, err)
// // 	}

// // 	_, err = demoClient.Cloud.NoOfControlPlane(1, true)
// // 	// it should return error
// // 	if err == nil {
// // 		t.Fatalf("setter should fail on when no < 3 controlplanes provided_no: %d", 1)
// // 	}

// // 	_, err = demoClient.Cloud.NoOfControlPlane(5, true)
// // 	// it should return error
// // 	if err != nil {
// // 		t.Fatalf("setter should not fail on when n >= 3 controlplanes err: %v", err)
// // 	}

// // 	no, err = demoClient.Cloud.NoOfControlPlane(-1, false)
// // 	if no != 5 {
// // 		t.Fatalf("Getter failed to get updated no of controlplanes array got no: %d and err: %v", no, err)
// // 	}
// // }

// // func TestNoOfWorkerPlane(t *testing.T) {
// // 	var no int
// // 	var err error
// // 	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
// // 	if no != -1 || err != nil {
// // 		t.Fatalf("Getter failed on unintalized workerplane array got no: %d and err: %v", no, err)
// // 	}

// // 	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 2, true)
// // 	// it shouldn't return err
// // 	if err != nil && !os.IsNotExist(err) {
// // 		t.Fatalf("setter should not fail on when no >= 0 workerplane provided_no: %d", 2)
// // 	}

// // 	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 2, true)
// // 	if err != nil {
// // 		t.Fatalf("setter should return nil when no changes happen workerplane err: %v", err)
// // 	}

// // 	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 3, true)
// // 	if err != nil {
// // 		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
// // 	}

// // 	_, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, 1, true)
// // 	if err != nil {
// // 		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
// // 	}

// // 	no, err = demoClient.Cloud.NoOfWorkerPlane(demoClient.Storage, -1, false)
// // 	if no != 1 {
// // 		t.Fatalf("Getter failed to get updated no of workerplane array got no: %d and err: %v", no, err)
// // 	}
// // }

// // // func TestValidRegion(t *testing.T) {
// // // 	fortesting := map[string]error{
// // // 		"fake":    nil,
// // // 		"eastus":  errors.New("Error"),
// // // 		"eastus2": nil,
// // // 	}

// // // 	for key, val := range fortesting {
// // // 		if aErr := isValidRegion(fakeaws, key); (aErr != nil && val == nil) || (aErr == nil && val != nil) {
// // // 			t.Fatalf("For Region `%s`. Expected `%v` but got `%v`", key, val, aErr)
// // // 		}
// // // 	}
// // // }

// // func TestResName(t *testing.T) {

// // 	if ret := fakeaws.Name("demo"); ret == nil {
// // 		t.Fatalf("returned nil for valid res name")
// // 	}
// // 	fakeaws.mxName.Unlock()
// // 	if fakeaws.metadata.resName != "demo" {
// // 		t.Fatalf("Correct assignment missing")
// // 	}

// // 	if ret := fakeaws.Name("12demo"); ret != nil {
// // 		t.Fatalf("returned interface for invalid res name")
// // 	}
// // 	fakeaws.mxName.Unlock()
// // }

// // func TestRole(t *testing.T) {
// // 	validSet := []consts.KsctlRole{consts.RoleCp, consts.RoleLb, consts.RoleDs, consts.RoleWp}
// // 	for _, val := range validSet {
// // 		if ret := fakeaws.Role(val); ret == nil {
// // 			t.Fatalf("returned nil for valid role")
// // 		}
// // 		fakeaws.mxRole.Unlock()
// // 		if fakeaws.metadata.role != val {
// // 			t.Fatalf("Correct assignment missing")
// // 		}
// // 	}
// // 	if ret := fakeaws.Role("fake"); ret != nil {
// // 		t.Fatalf("returned interface for invalid role")
// // 	}
// // 	fakeaws.mxRole.Unlock()
// // }

// // func TestVMType(t *testing.T) {
// // 	if ret := fakeaws.VMType("fake"); ret == nil {
// // 		t.Fatalf("returned nil for valid vm type")
// // 	}
// // 	fakeaws.mxVMType.Unlock()
// // 	if fakeaws.metadata.vmType != "fake" {
// // 		t.Fatalf("Correct assignment missing")
// // 	}

// // 	if ret := fakeaws.VMType(""); ret != nil {
// // 		t.Fatalf("returned interface for invalid vm type")
// // 	}
// // 	fakeaws.mxVMType.Unlock()
// // }

// // func TestVisibility(t *testing.T) {
// // 	if fakeaws.Visibility(true); !fakeaws.metadata.public {
// // 		t.Fatalf("Visibility setting not working")
// // 	}
// // }

// // func TestK8sVersion(t *testing.T) {
// // 	// these are invalid
// // 	// input and output
// // 	forTesting := []string{
// // 		"1.27.1",
// // 		"1.27",
// // 		"1.28.1",
// // 	}

// // 	for i := 0; i < len(forTesting); i++ {
// // 		var ver string = forTesting[i]
// // 		if i < 2 {
// // 			if ret := fakeaws.Version(ver); ret == nil {
// // 				t.Fatalf("returned nil for valid version")
// // 			}
// // 			if ver != fakeaws.metadata.k8sVersion {
// // 				t.Fatalf("set value is not equal to input value")
// // 			}
// // 		} else {
// // 			if ret := fakeaws.Version(ver); ret != nil {
// // 				t.Fatalf("returned interface for invalid version")
// // 			}
// // 		}
// // 	}

// // }

// // func TestCniAndApps(t *testing.T) {

// // 	testCases := map[string]bool{
// // 		string(consts.CNIAzure):   false,
// // 		string(consts.CNIKubenet): false,
// // 		string(consts.CNICilium):  true,
// // 	}

// // 	for k, v := range testCases {
// // 		got := fakeaws.CNI(k)
// // 		assert.Equal(t, got, v, "missmatch")
// // 	}

// // 	got := fakeaws.Application("abcd")
// // 	if !got {
// // 		t.Fatalf("application should be external")
// // 	}
// // }

// // ---------------------------------------------------------------------------------------------------

// func TestHACluster(t *testing.T) {

// 	fakeaws.region = "fake"
// 	fakeaws.clusterName = "fakeaws"
// 	fakeaws.haCluster = true
// 	fakeaws.region = "fake"

// 	// size
// 	fakeaws.metadata.noCP = 7
// 	fakeaws.metadata.noDS = 5
// 	fakeaws.metadata.noWP = 10
// 	fakeaws.metadata.public = true
// 	fakeaws.metadata.k8sName = consts.K8sK3s

// 	t.Run("init state", func(t *testing.T) {

// 		if err := fakeaws.InitState(demoClient.Storage, consts.OperationStateCreate); err != nil {
// 			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
// 		}

// 		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
// 		assert.Equal(t, clusterDirName, fakeaws.clusterName+" "+fakeaws.vpc+" "+fakeaws.region, "clusterdir not equal")
// 		assert.Equal(t, awsCloudState.IsCompleted, false, "cluster should not be completed")

// 		_, err := demoClient.Storage.Path(utils.GetPath(consts.UtilClusterPath, consts.CloudCivo, consts.ClusterTypeHa, clusterDirName, STATE_FILE_NAME)).Load()
// 		if os.IsExist(err) {
// 			t.Fatalf("State file and cluster directory present where it should not be")
// 		}
// 	})

// 	t.Run("Create network", func(t *testing.T) {
// 		fakeaws.Name("fake-data-not-used").NewNetwork(demoClient.Storage)
// 		assert.Equal(t, fakeaws.Name("fake-data-not-used").NewNetwork(demoClient.Storage), nil, "Network should be created")
// 		assert.Equal(t, awsCloudState.IsCompleted, false, "cluster should not be completed")

// 		//assert.Equal(t, awsCloudState.VPCNAME, fakeaws.vpc, "resource group not saved")
// 		//assert.Equal(t, awsCloudState.SubnetName, fakeaws.clusterName+"-subnet", "subnet should be created")
// 		//assert.Equal(t, awsCloudState.SubnetID, fakeaws.metadata.resName, "subnet should be created")
// 		//assert.Equal(t, awsCloudState.GatewayID, fakeaws.metadata.resName, "gateway should be created")
// 		//assert.Equal(t, awsCloudState.RouteTableID, fakeaws.metadata.resName, "route table should be created")
// 		//assert.Equal(t, awsCloudState.NetworkAclID, fakeaws.metadata.resName, "network acl should be created")

// 		assert.Assert(t, len(awsCloudState.VPCNAME) > 0, "resource group not saved")
// 		assert.Assert(t, len(awsCloudState.SubnetName) > 0, "subnet should be created")
// 		assert.Assert(t, len(awsCloudState.SubnetID) > 0, "subnet should be created")
// 		//assert.Assert(t, len(awsCloudState.GatewayID) > 0, "gateway should be created")
// 		//assert.Assert(t, len(awsCloudState.RouteTableID) > 0, "route table should be created")
// 		//assert.Assert(t, len(awsCloudState.NetworkAclID) > 0, "network acl should be created")

// 		log.Debug("Network created", "vpc", awsCloudState.VPCNAME, "subnet", awsCloudState.SubnetName, "subnetID", awsCloudState.SubnetID)

// 	})

// 	t.Run("Create ssh", func(t *testing.T) {

// 		assert.Equal(t, fakeaws.Name("fake-ssh").CreateUploadSSHKeyPair(demoClient.Storage), nil, "ssh key failed")

// 		assert.Equal(t, awsCloudState.SSHKeyName, fakeaws.metadata.resName, "sshid must be present")

// 		assert.Equal(t, awsCloudState.SSHUser, "ubuntu", "ssh user missing")
// 		assert.Equal(t, awsCloudState.SSHPrivateKeyLoc, utils.GetPath(consts.UtilSSHPath, consts.CloudAws, clusterType, clusterDirName), "ssh private key loc missing")

// 		assert.Equal(t, awsCloudState.IsCompleted, false, "cluster should not be completed")

// 	})

// 	t.Run("Create Firewalls", func(t *testing.T) {

// 		t.Run("Controlplane", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleCp)
// 			fakeaws.Name("fake-fw-cp")

// 			assert.Equal(t, fakeaws.NewFirewall(demoClient.Storage), nil, "new firewall failed")

// 			assert.Equal(t, awsCloudState.InfoControlPlanes.NetworkSecurityGroup, fakeaws.metadata.resName, "firewallID for controlplane absent")
// 			assert.Assert(t, len(awsCloudState.InfoControlPlanes.NetworkSecurityGroup) > 0, "fw id for controlplane missing")
// 		})
// 		t.Run("Workerplane", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleWp)
// 			fakeaws.Name("fake-fw-wp")

// 			assert.Equal(t, fakeaws.NewFirewall(demoClient.Storage), nil, "new firewall failed")
// 			assert.Equal(t, awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup, fakeaws.metadata.resName, "firewallID for workerplane absent")
// 			assert.Assert(t, len(awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup) > 0, "fw id for workerplane missing")
// 		})
// 		t.Run("Loadbalancer", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleLb)
// 			fakeaws.Name("fake-fw-lb")

// 			assert.Equal(t, fakeaws.NewFirewall(demoClient.Storage), nil, "new firewall failed")
// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.NetworkSecurityGroup, fakeaws.metadata.resName, "firewallID for loadbalacer absent")
// 			assert.Assert(t, len(awsCloudState.InfoLoadBalancer.NetworkSecurityGroup) > 0, "fw id for loadbalacer missing")
// 		})
// 		t.Run("Datastore", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleDs)
// 			fakeaws.Name("fake-fw-ds")

// 			assert.Equal(t, fakeaws.NewFirewall(demoClient.Storage), nil, "new firewall failed")
// 			assert.Equal(t, awsCloudState.InfoDatabase.NetworkSecurityGroup, fakeaws.metadata.resName, "firewallID for datastore absent")
// 			assert.Assert(t, len(awsCloudState.InfoDatabase.NetworkSecurityGroup) > 0, "fw id for datastore missing")
// 		})

// 	})

// 	t.Run("Create VMs", func(t *testing.T) {
// 		t.Run("Loadbalancer", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleLb)
// 			fakeaws.Name("fake-lb")
// 			fakeaws.VMType("fake")

// 			assert.Equal(t, fakeaws.NewVM(demoClient.Storage, 0), nil, "new vm failed")
// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.Name, fakeaws.metadata.resName, "missmatch of Loadbalancer VM name")
// 			// assert.Assert(t, len(awsCloudState.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

// 			// assert.Equal(t, awsCloudState.InfoLoadBalancer.DiskName, fakeaws.metadata.resName+"-disk", "missmatch of Loadbalancer disk name")

// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.NetworkInterfaceName, fakeaws.metadata.resName+"-nic", "missmatch of Loadbalancer nic name")
// 			assert.Assert(t, len(awsCloudState.InfoLoadBalancer.NetworkInterfaceName) > 0, "missmatch of Loadbalancer nic must be created")
// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.PrivateIP, "192.168.X.Y", "missmatch of Loadbalancer private ip NIC")

// 		})
// 		t.Run("Controlplanes", func(t *testing.T) {

// 			if _, err := fakeaws.NoOfControlPlane(fakeaws.metadata.noCP, true); err != nil {
// 				t.Fatalf("Failed to set the controlplane")
// 			}

// 			for i := 0; i < fakeaws.metadata.noCP; i++ {
// 				t.Run("controlplane", func(t *testing.T) {

// 					fakeaws.Name(fmt.Sprintf("fake-cp-%d", i))
// 					fakeaws.Role(consts.RoleCp)
// 					fakeaws.VMType("fake")

// 					assert.Equal(t, fakeaws.NewVM(demoClient.Storage, i), nil, "new vm failed")
// 					assert.Equal(t, awsCloudState.InfoControlPlanes.Names[i], fakeaws.metadata.resName, "missmatch of controlplane VM name")
// 					// assert.Assert(t, len(awsCloudState.InfoControlPlanes.Hostnames[i]) > 0, "missmatch of controlplane vm hostname")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.DiskNames[i], fakeaws.metadata.resName+"-disk", "missmatch of controlplane disk name")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.NetworkInterfaceNames[i], fakeaws.metadata.resName+"-nic", "missmatch of controlplane nic name")
// 					assert.Assert(t, len(awsCloudState.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
// 					assert.Equal(t, awsCloudState.InfoControlPlanes.PrivateIPs[i], "192.168.X.Y", "missmatch of controlplane private ip NIC")

// 				})
// 			}
// 		})

// 		t.Run("Datastores", func(t *testing.T) {
// 			// NOTE: the noDS is set to 1 becuase current implementation is only for single datastore
// 			// TODO: use the 1 as limit

// 			fakeaws.metadata.noDS = 1

// 			if _, err := fakeaws.NoOfDataStore(fakeaws.metadata.noDS, true); err != nil {
// 				t.Fatalf("Failed to set the datastore")
// 			}

// 			for i := 0; i < fakeaws.metadata.noDS; i++ {
// 				t.Run("datastore", func(t *testing.T) {

// 					fakeaws.Role(consts.RoleDs)
// 					fakeaws.Name(fmt.Sprintf("fake-ds-%d", i))
// 					fakeaws.VMType("fake")

// 					assert.Equal(t, fakeaws.NewVM(demoClient.Storage, i), nil, "new vm failed")

// 					assert.Equal(t, awsCloudState.InfoDatabase.Names[i], fakeaws.metadata.resName, "missmatch of datastore VM name")
// 					// assert.Assert(t, len(awsCloudState.InfoDatabase.Hostnames[i]) > 0, "missmatch of datastore vm hostname")

// 					assert.Equal(t, awsCloudState.InfoDatabase.DiskNames[i], fakeaws.metadata.resName+"-disk", "missmatch of datastore disk name")

// 					assert.Equal(t, awsCloudState.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

// 					assert.Equal(t, awsCloudState.InfoDatabase.NetworkInterfaceNames[i], fakeaws.metadata.resName+"-nic", "missmatch of datastore nic name")
// 					assert.Assert(t, len(awsCloudState.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
// 					assert.Equal(t, awsCloudState.InfoDatabase.PrivateIPs[i], "192.168.X.Y", "missmatch of datastore private ip NIC")

// 				})
// 			}
// 		})
// 		t.Run("Workplanes", func(t *testing.T) {

// 			if _, err := fakeaws.NoOfWorkerPlane(demoClient.Storage, fakeaws.metadata.noWP, true); err != nil {
// 				t.Fatalf("Failed to set the workerplane")
// 			}

// 			for i := 0; i < fakeaws.metadata.noWP; i++ {
// 				t.Run("workerplane", func(t *testing.T) {

// 					fakeaws.Role(consts.RoleWp)
// 					fakeaws.Name(fmt.Sprintf("fake-wp-%d", i))
// 					fakeaws.VMType("fake")

// 					assert.Equal(t, fakeaws.NewVM(demoClient.Storage, i), nil, "new vm failed")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.Names[i], fakeaws.metadata.resName, "missmatch of workerplane VM name")
// 					// assert.Assert(t, len(awsCloudState.InfoWorkerPlanes.Hostnames[i]) > 0, "missmatch of workerplane vm hostname")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.DiskNames[i], fakeaws.metadata.resName+"-disk", "missmatch of workerplane disk name")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames[i], fakeaws.metadata.resName+"-nic", "missmatch of workerplane nic name")
// 					assert.Assert(t, len(awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.PrivateIPs[i], "192.168.X.Y", "missmatch of workerplane private ip NIC")

// 				})
// 			}

// 			assert.Equal(t, awsCloudState.IsCompleted, true, "cluster should be completed")
// 		})
// 	})

// 	fmt.Println(fakeaws.GetHostNameAllWorkerNode())
// 	// t.Run("get hostname of workerplanes", func(t *testing.T) {
// 	// 	expected := awsCloudState.InfoWorkerPlanes.Hostnames

// 	// 	got := fakeaws.GetHostNameAllWorkerNode()
// 	// 	assert.DeepEqual(t, got, expected)
// 	// })

// 	t.Run("check getState()", func(t *testing.T) {
// 		expected, err := fakeaws.GetStateFile(demoClient.Storage)
// 		assert.NilError(t, err, "no error should be there for getstate")

// 		got, _ := json.Marshal(awsCloudState)
// 		assert.DeepEqual(t, string(got), expected)
// 	})

// 	// t.Run("Get cluster ha", func(t *testing.T) {
// 	// 	expected := []cloud.AllClusterData{
// 	// 		cloud.AllClusterData{
// 	// 			Name:       fakeaws.clusterName,
// 	// 			Region:     fakeaws.region,
// 	// 			Provider:   consts.CloudAzure,
// 	// 			Type:       consts.ClusterTypeHa,
// 	// 			NoWP:       fakeaws.noWP,
// 	// 			NoCP:       fakeaws.noCP,
// 	// 			NoDS:       fakeaws.noDS,
// 	// 			K8sDistro:  consts.K8sK3s,
// 	// 			K8sVersion: awsCloudState.KubernetesVer,
// 	// 		},
// 	// 	}
// 	// 	// got, err := GetRAWClusterInfos(demoClient.Storage, demoClient.Metadata)
// 	// 	// assert.NilError(t, err, "no error should be there")
// 	// 	// assert.DeepEqual(t, got, expected)
// 	// })

// 	// explicit clean
// 	awsCloudState = nil

// 	// TODO: check for the Passing the state to the kubernetes distribution function GetStateForHACluster

// 	// use init state firest
// 	t.Run("init state deletion", func(t *testing.T) {

// 		if err := fakeaws.InitState(demoClient.Storage, consts.OperationStateDelete); err != nil {
// 			t.Fatalf("Unable to init the state for delete, Reason: %v", err)
// 		}

// 		assert.Equal(t, clusterType, consts.ClusterTypeHa, "clustertype should be managed")
// 		assert.Equal(t, clusterDirName, fakeaws.clusterName+" "+fakeaws.vpc+" "+fakeaws.region, "clusterdir not equal")
// 	})

// 	t.Run("Get all counters", func(t *testing.T) {
// 		var err error
// 		fakeaws.metadata.noCP, err = fakeaws.NoOfControlPlane(-1, false)
// 		assert.Assert(t, err == nil)

// 		fakeaws.metadata.noWP, err = fakeaws.NoOfWorkerPlane(demoClient.Storage, -1, false)
// 		assert.Assert(t, err == nil)

// 		fakeaws.metadata.noDS, err = fakeaws.NoOfDataStore(-1, false)
// 		assert.Assert(t, err == nil)
// 	})

// 	t.Run("Delete VMs", func(t *testing.T) {
// 		t.Run("Loadbalancer", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleLb)

// 			assert.Equal(t, fakeaws.DelVM(demoClient.Storage, 0), nil, "del vm failed")

// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.Name, "", "missmatch of Loadbalancer VM name")
// 			// assert.Equal(t, awsCloudState.InfoLoadBalancer.HostName, "", "missmatch of Loadbalancer vm hostname")

// 			// assert.Equal(t, awsCloudState.InfoLoadBalancer.DiskName, "", "missmatch of Loadbalancer disk name")

// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.PublicIP, "", "missmatch of Loadbalancer pub ip")

// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.NetworkInterfaceName, "", "missmatch of Loadbalancer nic name")
// 			assert.Assert(t, len(awsCloudState.InfoLoadBalancer.NetworkInterfaceName) == 0, "missmatch of Loadbalancer nic must be created")
// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.PrivateIP, "", "missmatch of Loadbalancer private ip NIC")

// 		})

// 		t.Run("Workerplane", func(t *testing.T) {

// 			for i := 0; i < fakeaws.metadata.noWP; i++ {
// 				t.Run("workerplane", func(t *testing.T) {
// 					fakeaws.Role(consts.RoleWp)

// 					assert.Equal(t, fakeaws.DelVM(demoClient.Storage, i), nil, "del vm failed")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.Names[i], "", "missmatch of workerplane VM name")
// 					// assert.Equal(t, awsCloudState.InfoWorkerPlanes.Hostnames[i], "", "missmatch of workerplane vm hostname")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.DiskNames[i], "", "missmatch of workerplane disk name")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.PublicIPs[i], "", "missmatch of workerplane pub ip")

// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames[i], "", "missmatch of workerplane nic name")
// 					assert.Assert(t, len(awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")
// 					assert.Equal(t, awsCloudState.InfoWorkerPlanes.PrivateIPs[i], "", "missmatch of workerplane private ip NIC")

// 				})
// 			}
// 		})
// 		t.Run("Controlplane", func(t *testing.T) {

// 			for i := 0; i < fakeaws.metadata.noCP; i++ {
// 				t.Run("controlplane", func(t *testing.T) {
// 					fakeaws.Role(consts.RoleCp)

// 					assert.Equal(t, fakeaws.DelVM(demoClient.Storage, i), nil, "del vm failed")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.Names[i], "", "missmatch of controlplane VM name")
// 					// assert.Equal(t, awsCloudState.InfoControlPlanes.Hostnames[i], "", "missmatch of controlplane vm hostname")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.DiskNames[i], "", "missmatch of controlplane disk name")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.PublicIPs[i], "", "missmatch of controlplane pub ip")

// 					assert.Equal(t, awsCloudState.InfoControlPlanes.NetworkInterfaceNames[i], "", "missmatch of controlplane nic name")
// 					assert.Assert(t, len(awsCloudState.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")
// 					assert.Equal(t, awsCloudState.InfoControlPlanes.PrivateIPs[i], "", "missmatch of controlplane private ip NIC")

// 				})
// 			}
// 		})
// 		t.Run("DataStore", func(t *testing.T) {

// 			for i := 0; i < fakeaws.metadata.noDS; i++ {
// 				t.Run("datastore", func(t *testing.T) {
// 					fakeaws.Role(consts.RoleDs)

// 					assert.Equal(t, fakeaws.DelVM(demoClient.Storage, i), nil, "del vm failed")

// 					assert.Equal(t, awsCloudState.InfoDatabase.Names[i], "", "missmatch of datastore VM name")
// 					// assert.Equal(t, awsCloudState.InfoDatabase.Hostnames[i], "", "missmatch of datastore vm hostname")

// 					assert.Equal(t, awsCloudState.InfoDatabase.DiskNames[i], "", "missmatch of datastore disk name")

// 					assert.Equal(t, awsCloudState.InfoDatabase.PublicIPs[i], "", "missmatch of datastore pub ip")

// 					assert.Equal(t, awsCloudState.InfoDatabase.NetworkInterfaceNames[i], "", "missmatch of datastore nic name")
// 					assert.Assert(t, len(awsCloudState.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")
// 					assert.Equal(t, awsCloudState.InfoDatabase.PrivateIPs[i], "", "missmatch of datastore private ip NIC")

// 				})
// 			}
// 		})
// 	})

// 	t.Run("Delete Firewalls", func(t *testing.T) {

// 		t.Run("Controlplane", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleCp)

// 			assert.Equal(t, fakeaws.DelFirewall(demoClient.Storage), nil, "del firewall failed")

// 			assert.Equal(t, awsCloudState.InfoControlPlanes.NetworkSecurityGroup, "", "firewallID for controlplane absent")
// 			assert.Assert(t, len(awsCloudState.InfoControlPlanes.NetworkSecurityGroup) == 0, "fw id for controlplane missing")
// 		})
// 		t.Run("Workerplane", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleWp)

// 			assert.Equal(t, fakeaws.DelFirewall(demoClient.Storage), nil, "new firewall failed")

// 			assert.Equal(t, awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup, "", "firewallID for workerplane absent")
// 			assert.Assert(t, len(awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup) == 0, "fw id for workerplane missing")
// 		})
// 		t.Run("Loadbalancer", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleLb)

// 			assert.Equal(t, fakeaws.DelFirewall(demoClient.Storage), nil, "new firewall failed")

// 			assert.Equal(t, awsCloudState.InfoLoadBalancer.NetworkInterfaceName, "", "firewallID for loadbalacer absent")
// 			assert.Assert(t, len(awsCloudState.InfoLoadBalancer.NetworkSecurityGroup) == 0, "fw id for loadbalacer missing")
// 		})
// 		t.Run("Datastore", func(t *testing.T) {
// 			fakeaws.Role(consts.RoleDs)

// 			assert.Equal(t, fakeaws.DelFirewall(demoClient.Storage), nil, "new firewall failed")

// 			assert.Equal(t, awsCloudState.InfoDatabase.NetworkSecurityGroup, "", "firewallID for datastore absent")
// 			assert.Assert(t, len(awsCloudState.InfoDatabase.NetworkSecurityGroup) == 0, "fw id for datastore missing")
// 		})

// 	})

// 	t.Run("Delete ssh", func(t *testing.T) {

// 		assert.Equal(t, fakeaws.DelSSHKeyPair(demoClient.Storage), nil, "ssh key failed")

// 		assert.Equal(t, awsCloudState.SSHKeyName, "", "sshid must be present")

// 		assert.Equal(t, awsCloudState.SSHUser, "", "ssh user not set")
// 		assert.Equal(t, awsCloudState.SSHPrivateKeyLoc, "", "ssh private key loc missing")

// 	})

// 	t.Run("Delete network", func(t *testing.T) {
// 		assert.Equal(t, fakeaws.DelNetwork(demoClient.Storage), nil, "Network should be deleted")

// 		assert.Equal(t, awsCloudState.VPCID, "", "resource group not saved")
// 		// assert.Equal(t, awsCloudState.VirtualNetworkName, "", "virtual net should be created")
// 		assert.Equal(t, awsCloudState.SubnetID, "", "subnet should be created")

// 		// assert.Assert(t, len(awsCloudState.VirtualNetworkID) == 0, "virtual net should be created")
// 		assert.Assert(t, len(awsCloudState.SubnetID) == 0, "subnet should be created")
// 	})
// }
