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
	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *AzureProvider) DelFirewall(storage types.StorageFactory) error {
	role := <-obj.chRole

	log.Debug(azureCtx, "Printing", "role", role)

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName
	case consts.RoleWp:
		nsg = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName
	case consts.RoleLb:
		nsg = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName
	case consts.RoleDs:
		nsg = mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName
	}

	if len(nsg) == 0 {
		log.Print(azureCtx, "skipped firewall already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteSecurityGrp(nsg, nil)
	if err != nil {
		return err
	}
	log.Print(azureCtx, "firewall deleting...", "name", nsg)

	_, err = obj.client.PollUntilDoneDelNSG(azureCtx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch role {
	case consts.RoleCp:
		mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName = ""
		mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID = ""
	case consts.RoleWp:
		mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID = ""
		mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName = ""
	case consts.RoleLb:
		mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID = ""
		mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName = ""
	case consts.RoleDs:
		mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID = ""
		mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName = ""
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "Deleted network security group", "name", nsg)

	return nil
}

func (obj *AzureProvider) NewFirewall(storage types.StorageFactory) error {
	name := <-obj.chResName
	role := <-obj.chRole

	log.Debug(azureCtx, "Printing", "name", name, "role", role)

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName
	case consts.RoleWp:
		nsg = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName
	case consts.RoleLb:
		nsg = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName
	case consts.RoleDs:
		nsg = mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName
	}
	if len(nsg) != 0 {
		log.Success(azureCtx, "skipped firewall already created", "name", nsg)
		return nil
	}
	netCidr := mainStateDocument.CloudInfra.Azure.NetCidr
	kubernetesDistro := consts.KsctlKubernetes(mainStateDocument.BootstrapProvider)

	var securityRules []*armnetwork.SecurityRule
	switch role {
	case consts.RoleCp:
		securityRules = firewallRuleControlPlane(netCidr, kubernetesDistro)
	case consts.RoleWp:
		securityRules = firewallRuleWorkerPlane(netCidr, kubernetesDistro)
	case consts.RoleLb:
		securityRules = firewallRuleLoadBalancer()
	case consts.RoleDs:
		securityRules = firewallRuleDataStore(netCidr)
	}

	log.Debug(azureCtx, "Printing", "firewallrule", securityRules)

	parameters := armnetwork.SecurityGroup{
		Location: utilities.Ptr(obj.region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}

	pollerResponse, err := obj.client.BeginCreateSecurityGrp(name, parameters, nil)
	if err != nil {
		return err
	}

	switch role {
	case consts.RoleCp:
		mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName = name
	case consts.RoleWp:
		mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName = name
	case consts.RoleLb:
		mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName = name
	case consts.RoleDs:
		mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName = name
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Print(azureCtx, "creating firewall...", "name", name)

	resp, err := obj.client.PollUntilDoneCreateNSG(azureCtx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch role {
	case consts.RoleCp:
		mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID = *resp.ID
	case consts.RoleWp:
		mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID = *resp.ID
	case consts.RoleLb:
		mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID = *resp.ID
	case consts.RoleDs:
		mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID = *resp.ID
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "Created network security group", "name", *resp.Name)

	return nil
}

func convertToProviderSpecific(_rules []helpers.FirewallRule) []*armnetwork.SecurityRule {
	rules := []*armnetwork.SecurityRule{}
	priority := int32(100)
	for _, _r := range _rules {
		priority++

		var protocol armnetwork.SecurityRuleProtocol
		var action armnetwork.SecurityRuleAccess
		var direction armnetwork.SecurityRuleDirection
		var portRange string
		var srcCidr, destCidr string

		switch _r.Action {
		case consts.FirewallActionAllow:
			action = armnetwork.SecurityRuleAccessAllow
		case consts.FirewallActionDeny:
			action = armnetwork.SecurityRuleAccessDeny
		default:
			action = armnetwork.SecurityRuleAccessAllow
		}

		switch _r.Protocol {
		case consts.FirewallActionTCP:
			protocol = armnetwork.SecurityRuleProtocolTCP
		case consts.FirewallActionUDP:
			protocol = armnetwork.SecurityRuleProtocolUDP
		default:
			protocol = armnetwork.SecurityRuleProtocolTCP
		}

		switch _r.Direction {
		case consts.FirewallActionIngress:
			direction = armnetwork.SecurityRuleDirectionInbound
			srcCidr = _r.Cidr
			destCidr = mainStateDocument.CloudInfra.Azure.NetCidr
		case consts.FirewallActionEgress:
			direction = armnetwork.SecurityRuleDirectionOutbound
			destCidr = _r.Cidr
			srcCidr = mainStateDocument.CloudInfra.Azure.NetCidr
		default:
			direction = armnetwork.SecurityRuleDirectionInbound
		}

		if _r.StartPort == _r.EndPort {
			portRange = _r.StartPort
		} else {
			portRange = _r.StartPort + "-" + _r.EndPort
			if portRange == "1-65535" {
				portRange = "*"
			}
		}

		rules = append(rules, &armnetwork.SecurityRule{
			Name: utilities.Ptr(_r.Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				SourceAddressPrefix:      utilities.Ptr(srcCidr),
				SourcePortRange:          utilities.Ptr("*"),
				DestinationAddressPrefix: utilities.Ptr(destCidr),
				DestinationPortRange:     utilities.Ptr(portRange),
				Protocol:                 utilities.Ptr(protocol),
				Access:                   utilities.Ptr(action),
				Priority:                 utilities.Ptr[int32](priority),
				Description:              utilities.Ptr(_r.Description),
				Direction:                utilities.Ptr(direction),
			},
		})
	}

	return rules

}

func firewallRuleControlPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) (securityRules []*armnetwork.SecurityRule) {
	return convertToProviderSpecific(
		helpers.FirewallForControlplane_BASE(internalNetCidr, bootstrap),
	)
}

func firewallRuleWorkerPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) (securityRules []*armnetwork.SecurityRule) {
	return convertToProviderSpecific(
		helpers.FirewallForWorkerplane_BASE(internalNetCidr, bootstrap),
	)
}

func firewallRuleLoadBalancer() (securityRules []*armnetwork.SecurityRule) {
	return convertToProviderSpecific(
		helpers.FirewallForLoadBalancer_BASE(),
	)
}

func firewallRuleDataStore(internalNetCidr string) (securityRules []*armnetwork.SecurityRule) {
	return convertToProviderSpecific(
		helpers.FirewallForDataStore_BASE(internalNetCidr),
	)
}
