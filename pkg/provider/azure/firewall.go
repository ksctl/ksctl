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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/firewall"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func (p *Provider) DelFirewall() error {
	role := <-p.chRole

	p.l.Debug(p.ctx, "Printing", "role", role)

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName
	case consts.RoleWp:
		nsg = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName
	case consts.RoleLb:
		nsg = p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName
	case consts.RoleDs:
		nsg = p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName
	}

	if len(nsg) == 0 {
		p.l.Print(p.ctx, "skipped firewall already deleted")
		return nil
	}

	pollerResponse, err := p.client.BeginDeleteSecurityGrp(nsg, nil)
	if err != nil {
		return err
	}
	p.l.Print(p.ctx, "firewall deleting...", "name", nsg)

	_, err = p.client.PollUntilDoneDelNSG(p.ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch role {
	case consts.RoleCp:
		p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName = ""
		p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID = ""
	case consts.RoleWp:
		p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID = ""
		p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName = ""
	case consts.RoleLb:
		p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID = ""
		p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName = ""
	case consts.RoleDs:
		p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID = ""
		p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName = ""
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "Deleted network security group", "name", nsg)

	return nil
}

func (p *Provider) NewFirewall() error {
	name := <-p.chResName
	role := <-p.chRole

	p.l.Debug(p.ctx, "Printing", "name", name, "role", role)

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName
	case consts.RoleWp:
		nsg = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName
	case consts.RoleLb:
		nsg = p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName
	case consts.RoleDs:
		nsg = p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName
	}
	if len(nsg) != 0 {
		p.l.Success(p.ctx, "skipped firewall already created", "name", nsg)
		return nil
	}
	netCidr := p.state.CloudInfra.Azure.NetCidr
	kubernetesDistro := consts.KsctlKubernetes(p.state.BootstrapProvider)

	var securityRules []*armnetwork.SecurityRule
	switch role {
	case consts.RoleCp:
		securityRules = p.firewallRuleControlPlane(netCidr, kubernetesDistro)
	case consts.RoleWp:
		securityRules = p.firewallRuleWorkerPlane(netCidr, kubernetesDistro)
	case consts.RoleLb:
		securityRules = p.firewallRuleLoadBalancer()
	case consts.RoleDs:
		securityRules = p.firewallRuleDataStore(netCidr)
	}

	p.l.Debug(p.ctx, "Printing", "firewallrule", securityRules)

	parameters := armnetwork.SecurityGroup{
		Location: utilities.Ptr(p.Region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}

	pollerResponse, err := p.client.BeginCreateSecurityGrp(name, parameters, nil)
	if err != nil {
		return err
	}

	switch role {
	case consts.RoleCp:
		p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName = name
	case consts.RoleWp:
		p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName = name
	case consts.RoleLb:
		p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName = name
	case consts.RoleDs:
		p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName = name
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Print(p.ctx, "creating firewall...", "name", name)

	resp, err := p.client.PollUntilDoneCreateNSG(p.ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch role {
	case consts.RoleCp:
		p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID = *resp.ID
	case consts.RoleWp:
		p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID = *resp.ID
	case consts.RoleLb:
		p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID = *resp.ID
	case consts.RoleDs:
		p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID = *resp.ID
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "Created network security group", "name", *resp.Name)

	return nil
}

func (p *Provider) convertToProviderSpecific(_rules []firewall.FirewallRule) []*armnetwork.SecurityRule {
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
			destCidr = p.state.CloudInfra.Azure.NetCidr
		case consts.FirewallActionEgress:
			direction = armnetwork.SecurityRuleDirectionOutbound
			destCidr = _r.Cidr
			srcCidr = p.state.CloudInfra.Azure.NetCidr
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

func (p *Provider) firewallRuleControlPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) (securityRules []*armnetwork.SecurityRule) {
	return p.convertToProviderSpecific(
		firewall.FirewallforcontrolplaneBase(internalNetCidr, bootstrap),
	)
}

func (p *Provider) firewallRuleWorkerPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) (securityRules []*armnetwork.SecurityRule) {
	return p.convertToProviderSpecific(
		firewall.FirewallforworkerplaneBase(internalNetCidr, bootstrap),
	)
}

func (p *Provider) firewallRuleLoadBalancer() (securityRules []*armnetwork.SecurityRule) {
	return p.convertToProviderSpecific(
		firewall.FirewallforloadbalancerBase(),
	)
}

func (p *Provider) firewallRuleDataStore(internalNetCidr string) (securityRules []*armnetwork.SecurityRule) {
	return p.convertToProviderSpecific(
		firewall.FirewallfordatastoreBase(internalNetCidr),
	)
}
