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
	"strconv"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/firewall"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
	"gotest.tools/v3/assert"
)

func TestInitState(t *testing.T) {

	t.Run("Create state", func(t *testing.T) {

		if err := fakeClientVars.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to init the state for fresh start, Reason: %v", err)
		}

		assert.Equal(t, fakeClientVars.ClusterType, consts.ClusterTypeSelfMang, "clustertype should be managed")
		assert.Equal(t, fakeClientVars.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
		assert.Equal(t, fakeClientVars.Name("fake-net").NewNetwork(), nil, "Network should be created")
		assert.Equal(t, fakeClientVars.state.CloudInfra.Aws.B.IsCompleted, false, "cluster should not be completed")
	})

	t.Run("Try to resume", func(t *testing.T) {
		fakeClientVars.state.CloudInfra.Aws.B.IsCompleted = true
		assert.Equal(t, fakeClientVars.state.CloudInfra.Aws.B.IsCompleted, true, "cluster should not be completed")

		if err := fakeClientVars.InitState(consts.OperationCreate); err != nil {
			t.Fatalf("Unable to resume state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Get request", func(t *testing.T) {

		if err := fakeClientVars.InitState(consts.OperationGet); err != nil {
			t.Fatalf("Unable to get state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Delete request", func(t *testing.T) {

		if err := fakeClientVars.InitState(consts.OperationDelete); err != nil {
			t.Fatalf("Unable to Delete state, Reason: %v", err)
		}
	})

	t.Run("try to Trigger Invalid request", func(t *testing.T) {

		if err := fakeClientVars.InitState("test"); err == nil {
			t.Fatalf("Expected error but not got: %v", err)
		}
	})
}

func TestNoOfControlPlane(t *testing.T) {
	var no int
	var err error
	no, err = fakeClientVars.NoOfControlPlane(-1, false)
	if no != -1 || err == nil || (err != nil && !ksctlErrors.IsInvalidNoOfControlplane(err)) {
		t.Fatalf("Getter failed on unintalized controlplanes array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfControlPlane(1, true)
	if err == nil || (err != nil && !ksctlErrors.IsInvalidNoOfControlplane(err)) {
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
	if no != -1 || err == nil || (err != nil && !ksctlErrors.IsInvalidNoOfDatastore(err)) {
		t.Fatalf("Getter failed on unintalized datastore array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfDataStore(0, true)
	if err == nil || (err != nil && !ksctlErrors.IsInvalidNoOfDatastore(err)) {
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
	no, err = fakeClientVars.NoOfWorkerPlane(-1, false)
	if no != -1 || err == nil || (err != nil && !ksctlErrors.IsInvalidNoOfWorkerplane(err)) {
		t.Fatalf("Getter failed on unintalized workerplane array got no: %d and err: %v", no, err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(2, true)
	if err != nil {
		t.Fatalf("setter should not fail on when no >= 0 workerplane provided_no: %d", 2)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(2, true)
	if err != nil {
		t.Fatalf("setter should return nil when no changes happen workerplane err: %v", err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(3, true)
	if err != nil {
		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
	}

	_, err = fakeClientVars.NoOfWorkerPlane(1, true)
	if err != nil {
		t.Fatalf("setter should return nil when upscaling changes happen workerplane err: %v", err)
	}

	no, err = fakeClientVars.NoOfWorkerPlane(-1, false)
	if no != 1 {
		t.Fatalf("Getter failed to get updated no of workerplane array got no: %d and err: %v", no, err)
	}
}

func TestValidRegion(t *testing.T) {
	fortesting := map[string]error{
		"ap-south-1":  nil,
		"fake-region": nil,
	}

	for key, val := range fortesting {
		fakeClientVars.Region = key
		if err := fakeClientVars.isValidRegion(); err != val {
			t.Fatalf("Input region :`%s`. expected `%v` but got `%v`", key, val, err)
		}
	}
}

func TestK8sVersion(t *testing.T) {
	forTesting := []string{
		"1.30",
		"1.29",
		"1.28",
	}

	for i := 0; i < len(forTesting); i++ {
		var ver string = forTesting[i]
		if i < 2 {
			if ret := fakeClientVars.ManagedK8sVersion(ver); ret == nil {
				t.Fatalf("returned nil for valid version")
			}
			if ver != fakeClientVars.K8sVersion {
				t.Fatalf("set value is not equal to input value")
			}
		} else {
			if ret := fakeClientVars.ManagedK8sVersion(ver); ret != nil {
				t.Fatalf("returned interface for invalid version")
			}
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
	if fakeClientVars.Visibility(true); !fakeClientVars.public {
		t.Fatalf("Visibility setting not working")
	}
}

func TestCniAndApps(t *testing.T) {
	testCases := []struct {
		Addon           addons.ClusterAddons
		Valid           bool
		managedAddonCNI string
		managedAddonApp []string
	}{
		{
			addons.ClusterAddons{
				{
					Label: "ksctl",
					Name:  "cilium",
					IsCNI: true,
				},
				{
					Label: "eks",
					Name:  "none",
					IsCNI: true,
				},
			}, true, "none", []string{"eks-node-monitoring-agent"},
		},
		{
			addons.ClusterAddons{
				{
					Label: "eks",
					Name:  "aws",
					IsCNI: true,
				},
			}, false, "aws", []string{"eks-node-monitoring-agent"},
		},
		{
			addons.ClusterAddons{
				{
					Label: "eks",
					Name:  "vpc-cni",
					IsCNI: true,
				},
			}, false, "vpc-cni", []string{"eks-node-monitoring-agent"},
		},
		{
			addons.ClusterAddons{}, false, "aws", []string{"eks-node-monitoring-agent"},
		},
		{
			nil, false, "aws", []string{"eks-node-monitoring-agent"},
		},
		{
			addons.ClusterAddons{
				{
					Label:  "eks",
					Name:   "heheheh",
					Config: utilities.Ptr(`{"key":"value"}`),
				},
			}, false, "aws", []string{"heheheh", "eks-node-monitoring-agent"},
		},
		{
			addons.ClusterAddons{
				{
					Label: "eks",
					Name:  "heheheh",
				},
			}, false, "aws", []string{"heheheh", "eks-node-monitoring-agent"},
		},
	}

	for _, v := range testCases {
		got := fakeClientVars.ManagedAddons(v.Addon)
		assert.Equal(t, got, v.Valid, "missmatch in return value")
		assert.Equal(t, fakeClientVars.managedAddonCNI, v.managedAddonCNI, "missmatch in managedAddonCNI")
		assert.DeepEqual(t, fakeClientVars.managedAddonApp, v.managedAddonApp)
	}
}

func TestDeleteVarCluster(t *testing.T) {
	if err := storeVars.DeleteCluster(); err != nil {
		t.Fatal(err)
	}
}

func TestFirewallRules(t *testing.T) {
	_rules := []firewall.FirewallRule{
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
