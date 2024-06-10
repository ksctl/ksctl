package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	awsTypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"gotest.tools/v3/assert"
)

var (
	fakeClientHA *AwsProvider
	storeHA      types.StorageFactory

	fakeClientManaged *AwsProvider
	storeManaged      types.StorageFactory
	parentCtx         context.Context
	fakeClientVars    *AwsProvider
	storeVars         types.StorageFactory

	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)

	dir = path.Join(os.TempDir(), "ksctl-aws-test")
)

func TestMain(m *testing.M) {

	parentCtx = context.WithValue(
		context.TODO(),
		consts.KsctlCustomDirLoc,
		dir)

	fakeClientVars, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "fake-region",
		Provider:    consts.CloudAws,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, ProvideClient)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAws, "fake-region", "demo", consts.ClusterTypeHa)
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
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeClientVars.Name("fake-net").NewNetwork(storeVars), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		mainStateDocument.CloudInfra.Aws.B.IsCompleted = true
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, true, "cluster should not be completed")

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

func TestNoOfControlPlane(t *testing.T) {
	var no int
	var err error
	no, err = fakeClientVars.NoOfControlPlane(-1, false)
	if no != -1 || err == nil || (err != nil && !ksctlErrors.ErrInvalidNoOfControlplane.Is(err)) {
		t.Fatalf("Getter failed on unintalized controlplanes array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfControlPlane(1, true)
	if err == nil || (err != nil && !ksctlErrors.ErrInvalidNoOfControlplane.Is(err)) {
		t.Fatalf("setter should fail on when no < 3 controlplanes provided_no: %d", 1)
	}

	_, err = fakeClientVars.NoOfControlPlane(5, true)
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
	if no != -1 || err == nil || (err != nil && !ksctlErrors.ErrInvalidNoOfDatastore.Is(err)) {
		t.Fatalf("Getter failed on unintalized datastore array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfDataStore(0, true)
	if err == nil || (err != nil && !ksctlErrors.ErrInvalidNoOfDatastore.Is(err)) {
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
	if no != -1 || err == nil || (err != nil && !ksctlErrors.ErrInvalidNoOfWorkerplane.Is(err)) {
		t.Fatalf("Getter failed on unintalized workerplane array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(storeVars, 2, true)
	if err != nil {
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
		"ap-south-1": nil,
		"ap-south-2": nil,
	}

	for key, val := range fortesting {
		fakeClientVars.client.SetRegion(key)
		if err := isValidRegion(fakeClientVars); err != val {
			t.Fatalf("Input region :`%s`. expected `%v` but got `%v`", key, val, err)
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
}

func TestVisibility(t *testing.T) {
	if fakeClientVars.Visibility(true); !fakeClientVars.metadata.public {
		t.Fatalf("Visibility setting not working")
	}
}

func TestCniAndApps(t *testing.T) {

	testCases := map[string]bool{
		string(consts.CNIFlannel): false,
		string(consts.CNICilium):  false,
		string(consts.CNINone):    true,
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

func TestDeleteVarCluster(t *testing.T) {
	if err := storeVars.DeleteCluster(); err != nil {
		t.Fatal(err)
	}
}

func checkCurrentStateFile(t *testing.T) {

	if err := storeManaged.Setup(consts.CloudAws, mainStateDocument.Region, mainStateDocument.ClusterName, consts.ClusterTypeMang); err != nil {
		t.Fatal(err)
	}
	read, err := storeManaged.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, mainStateDocument, read)
}

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
	expectIng := ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: utilities.Ptr("143e124"),
		IpPermissions: []awsTypes.IpPermission{
			{
				FromPort: utilities.Ptr[int32](func() int32 {
					_p, _ := strconv.Atoi(_rules[1].StartPort)
					return int32(_p)
				}()),
				ToPort: utilities.Ptr[int32](func() int32 {
					_p, _ := strconv.Atoi(_rules[1].EndPort)
					return int32(_p)
				}()),
				IpProtocol: utilities.Ptr[string]("tcp"),
				IpRanges: []awsTypes.IpRange{
					{
						CidrIp:      utilities.Ptr[string](_rules[1].Cidr),
						Description: utilities.Ptr[string](_rules[1].Description),
					},
				},
			},
		},
	}
	expectEgr := ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: utilities.Ptr("143e124"),
		IpPermissions: []awsTypes.IpPermission{
			{
				FromPort: utilities.Ptr[int32](func() int32 {
					_p, _ := strconv.Atoi(_rules[0].StartPort)
					return int32(_p)
				}()),
				ToPort: utilities.Ptr[int32](func() int32 {
					_p, _ := strconv.Atoi(_rules[0].EndPort)
					return int32(_p)
				}()),
				IpProtocol: utilities.Ptr[string]("udp"),
				IpRanges: []awsTypes.IpRange{
					{
						CidrIp:      utilities.Ptr[string](_rules[0].Cidr),
						Description: utilities.Ptr[string](_rules[0].Description),
					},
				},
			},
		},
	}
	gotIng, gotEgr := convertToProviderSpecific(_rules, utilities.Ptr("143e124"))

	// Compare the expected and actual values
	assert.DeepEqual(t, expectIng.GroupId, gotIng.GroupId)
	assert.DeepEqual(t, expectIng.IpPermissions[0].FromPort, gotIng.IpPermissions[0].FromPort)
	assert.DeepEqual(t, expectIng.IpPermissions[0].ToPort, gotIng.IpPermissions[0].ToPort)
	assert.DeepEqual(t, expectIng.IpPermissions[0].IpProtocol, gotIng.IpPermissions[0].IpProtocol)
	assert.DeepEqual(t, expectIng.IpPermissions[0].IpRanges[0].CidrIp, gotIng.IpPermissions[0].IpRanges[0].CidrIp)
	assert.DeepEqual(t, expectIng.IpPermissions[0].IpRanges[0].Description, gotIng.IpPermissions[0].IpRanges[0].Description)

	assert.DeepEqual(t, expectEgr.GroupId, gotEgr.GroupId)
	assert.DeepEqual(t, expectEgr.IpPermissions[0].FromPort, gotEgr.IpPermissions[0].FromPort)
	assert.DeepEqual(t, expectEgr.IpPermissions[0].ToPort, gotEgr.IpPermissions[0].ToPort)
	assert.DeepEqual(t, expectEgr.IpPermissions[0].IpProtocol, gotEgr.IpPermissions[0].IpProtocol)
	assert.DeepEqual(t, expectEgr.IpPermissions[0].IpRanges[0].CidrIp, gotEgr.IpPermissions[0].IpRanges[0].CidrIp)
	assert.DeepEqual(t, expectEgr.IpPermissions[0].IpRanges[0].Description, gotEgr.IpPermissions[0].IpRanges[0].Description)

}

func TestHACluster(t *testing.T) {

	mainStateDocument = &storageTypes.StorageDocument{}
	fakeClientHA, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo-ha",
		Region:      "fake-region",
		Provider:    consts.CloudAws,
		IsHA:        true,
		NoCP:        7,
		NoDS:        5,
		NoWP:        10,
		K8sDistro:   consts.K8sK3s,
	}, parentLogger, mainStateDocument, ProvideClient)

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
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		_, err := storeHA.Read()
		if err == nil {
			t.Fatalf("State file and cluster directory present where it should not be")
		}
	})

	t.Run("Create network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.Name("fake-data-not-used").NewNetwork(storeHA), nil, "Network should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.VpcId, "3456d25f36g474g546", "want %s got %s", "3456d25f36g474g546", mainStateDocument.CloudInfra.Aws.VpcId)
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.VpcName, fakeClientHA.clusterName+"-vpc", "virtual net should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.SubnetName, fakeClientHA.clusterName+"-subnet", "subnet should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.SubnetID, "3456d25f36g474g546", "subnet should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.RouteTableID, "3456d25f36g474g546", "route table should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.GatewayID, "3456d25f36g474g546", "gateway should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.SubnetName, fakeClientHA.clusterName+"-subnet", "subnet should be created")

		checkCurrentStateFileHA(t)
	})

	t.Run("Create ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.Name("fake-ssh").CreateUploadSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.SSHKeyName, "fake-ssh", "sshid must be present")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.SSHUser, "ubuntu", "ssh user not set")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		checkCurrentStateFileHA(t)
	})

	t.Run("Create Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)
			fakeClientHA.Name("fake-fw-cp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup) > 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)
			fakeClientHA.Name("fake-fw-wp")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup) > 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-fw-lb")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup) > 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)
			fakeClientHA.Name("fake-fw-ds")

			assert.Equal(t, fakeClientHA.NewFirewall(storeHA), nil, "new firewall failed")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup) > 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Create VMs", func(t *testing.T) {
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)
			fakeClientHA.Name("fake-lb")
			fakeClientHA.VMType("fake")

			assert.Equal(t, fakeClientHA.NewVM(storeHA, 0), nil, "new vm failed")

			assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID, "test-instance-1234567890", "missmatch of Loadbalancer VM ID")
			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.HostName) > 0, "missmatch of Loadbalancer vm hostname")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP) > 0, "missmatch of Loadbalancer pub ip id must be created")
			assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP, "A.B.C.D", "missmatch of Loadbalancer pub ip")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId) > 0, "missmatch of Loadbalancer nic must be created")
			assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PrivateIP, "192.169.1.2", "missmatch of Loadbalancer private ip NIC")

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

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[i], "test-instance-1234567890", "missmatch of controlplane VM ID")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames[i]) > 0, "missmatch of controlplane vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs[i], "A.B.C.D", "missmatch of controlplane pub ip")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of controlplane nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[i], "192.169.1.2", "missmatch of controlplane private ip NIC")

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

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[i], "test-instance-1234567890", "missmatch of datastore VM ID")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames[i]) > 0, "missmatch of datastore vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs[i], "A.B.C.D", "missmatch of datastore pub ip")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[i]) > 0, "missmatch of datastore nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs[i], "192.169.1.2", "missmatch of datastore private ip NIC")

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

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[i], "test-instance-1234567890", "missmatch of workerplane VM ID")
					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames[i]) > 0, "missmatch of workerplane vm hostname")

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[i], "A.B.C.D", "missmatch of workerplane pub ip")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[i]) > 0, "missmatch of workerplane nic must be created")
					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[i], "192.169.1.2", "missmatch of workerplane private ip NIC")

					checkCurrentStateFileHA(t)
				})
			}

			assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.IsCompleted, false, "cluster should be completed")
		})
	})

	fmt.Println(fakeClientHA.GetHostNameAllWorkerNode())
	t.Run("get hostname of workerplanes", func(t *testing.T) {
		expected := mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames

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
				CloudProvider: consts.CloudAws,
				ClusterType:   consts.ClusterTypeHa,
				NoWP:          fakeClientHA.metadata.noWP,
				NoCP:          fakeClientHA.metadata.noCP,
				NoDS:          fakeClientHA.metadata.noDS,

				WP: []cloud.VMData{
					{VMSize: "fake"}, {VMSize: "fake"}, {VMSize: "fake"},
					{VMSize: "fake"}, {VMSize: "fake"}, {VMSize: "fake"},
					{VMSize: "fake"}, {VMSize: "fake"}, {VMSize: "fake"},
					{VMSize: "fake"},
				},
				CP: []cloud.VMData{
					{VMSize: "fake"}, {VMSize: "fake"}, {VMSize: "fake"},
					{VMSize: "fake"}, {VMSize: "fake"}, {VMSize: "fake"},
					{VMSize: "fake"},
				},
				DS: []cloud.VMData{
					{VMSize: "fake"}, {VMSize: "fake"}, {VMSize: "fake"},
					{VMSize: "fake"}, {VMSize: "fake"},
				},
				LB: cloud.VMData{VMSize: "fake"},

				K8sDistro:  "",
				K8sVersion: mainStateDocument.CloudInfra.Aws.B.KubernetesVer,
			},
		}
		got, err := fakeClientHA.GetRAWClusterInfos(storeHA)
		assert.NilError(t, err, "no error should be there")
		assert.DeepEqual(t, got, expected)
	})

	mainStateDocument = &storageTypes.StorageDocument{}
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

			assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID, "", "missmatch of Loadbalancer VM ID")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId) == 0, "missmatch of Loadbalancer nic must be created")
			checkCurrentStateFileHA(t)
		})

		t.Run("Workerplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noWP; i++ {
				t.Run("workerplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleWp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[i], "", "missmatch of workerplane VM ID")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of workerplane nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("Controlplane", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noCP; i++ {
				t.Run("controlplane", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleCp)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[i], "", "missmatch of controlplane VM ID")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[i]) == 0, "missmatch of controlplane nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
		t.Run("DataStore", func(t *testing.T) {

			for i := 0; i < fakeClientHA.metadata.noDS; i++ {
				t.Run("datastore", func(t *testing.T) {
					fakeClientHA.Role(consts.RoleDs)

					assert.Equal(t, fakeClientHA.DelVM(storeHA, i), nil, "del vm failed")

					assert.Equal(t, mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[i], "", "missmatch of datastore VM ID")

					assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[i]) == 0, "missmatch of datastore nic must be created")

					checkCurrentStateFileHA(t)
				})
			}
		})
	})

	t.Run("Delete Firewalls", func(t *testing.T) {

		t.Run("Controlplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleCp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "del firewall failed")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup) == 0, "fw id for controlplane missing")
		})
		t.Run("Workerplane", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleWp)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup) == 0, "fw id for workerplane missing")
		})
		t.Run("Loadbalancer", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleLb)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup) == 0, "fw id for loadbalacer missing")
		})
		t.Run("Datastore", func(t *testing.T) {
			fakeClientHA.Role(consts.RoleDs)

			assert.Equal(t, fakeClientHA.DelFirewall(storeHA), nil, "new firewall failed")

			assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup) == 0, "fw id for datastore missing")
		})

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete ssh", func(t *testing.T) {

		assert.Equal(t, fakeClientHA.DelSSHKeyPair(storeHA), nil, "ssh key failed")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.SSHKeyName, "", "sshid must be present")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.B.SSHUser, "", "ssh user not set")

		checkCurrentStateFileHA(t)
	})

	t.Run("Delete network", func(t *testing.T) {
		assert.Equal(t, fakeClientHA.DelNetwork(storeHA), nil, "Network should be deleted")

		assert.Equal(t, mainStateDocument.CloudInfra.Aws.VpcId, "", "resource group not saved")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.VpcName, "", "virtual net should be created")
		assert.Equal(t, mainStateDocument.CloudInfra.Aws.SubnetID, "", "subnet should be created")

		assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.SubnetID) == 0, "subnet should be created")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.RouteTableID) == 0, "route table should be created")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.GatewayID) == 0, "gateway should be created")
		assert.Assert(t, len(mainStateDocument.CloudInfra.Aws.SubnetName) == 0, "subnet should be created")
	})
}

// func TestGetSecretTokens(t *testing.T) {
// 	t.Run("expect demo data", func(t *testing.T) {
// 		expected := map[string][]byte{
// 			"aws_access_key_id":     []byte("fake"),
// 			"aws_secret_access_key": []byte("fake"),
// 		}

// 		for key, val := range expected {
// 			assert.NilError(t, os.Setenv(key, string(val)), "environment vars should be set")
// 		}
// 		actual, err := fakeClientVars.GetSecretTokens(storeVars)
// 		assert.NilError(t, err, "unable to get the secret token from the client")
// 		assert.DeepEqual(t, actual, expected)
// 	})
// }
