// Copyright 2024 ksctl
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
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"gotest.tools/v3/assert"
)

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
			Name: utilities.Ptr(_rules[0].Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      utilities.Ptr(mainStateDocument.CloudInfra.Azure.NetCidr),
				SourcePortRange:          utilities.Ptr("*"),
				DestinationAddressPrefix: utilities.Ptr(_rules[0].Cidr),
				DestinationPortRange:     utilities.Ptr(_rules[0].StartPort),
				Protocol:                 utilities.Ptr(armnetwork.SecurityRuleProtocolUDP),
				Access:                   utilities.Ptr(armnetwork.SecurityRuleAccessDeny),
				Priority:                 utilities.Ptr[int32](101),
				Description:              utilities.Ptr(_rules[0].Description),
				Direction:                utilities.Ptr(armnetwork.SecurityRuleDirectionOutbound),
			},
		},
		{
			Name: utilities.Ptr(_rules[1].Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      utilities.Ptr(_rules[1].Cidr),
				SourcePortRange:          utilities.Ptr("*"),
				DestinationAddressPrefix: utilities.Ptr(mainStateDocument.CloudInfra.Azure.NetCidr),
				DestinationPortRange:     utilities.Ptr(_rules[1].StartPort + "-" + _rules[1].EndPort),
				Protocol:                 utilities.Ptr(armnetwork.SecurityRuleProtocolTCP),
				Access:                   utilities.Ptr(armnetwork.SecurityRuleAccessAllow),
				Priority:                 utilities.Ptr[int32](102),
				Description:              utilities.Ptr(_rules[1].Description),
				Direction:                utilities.Ptr(armnetwork.SecurityRuleDirectionInbound),
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
