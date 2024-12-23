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

package civo

import (
	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/providers"
)

func (p *Provider) DelFirewall() error {
	role := <-p.chRole

	p.l.Debug(p.ctx, "Printing", "Role", role)

	var firewallID string
	switch role {
	case consts.RoleCp:
		if len(p.state.CloudInfra.Civo.FirewallIDControlPlanes) == 0 {
			p.l.Print(p.ctx, "skipped firewall for controlplane already deleted")
			return nil
		}
		firewallID = p.state.CloudInfra.Civo.FirewallIDControlPlanes

		_, err := p.client.DeleteFirewall(p.state.CloudInfra.Civo.FirewallIDControlPlanes)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Civo.FirewallIDControlPlanes = ""

	case consts.RoleWp:
		if len(p.state.CloudInfra.Civo.FirewallIDWorkerNodes) == 0 {
			p.l.Print(p.ctx, "skipped firewall for workerplane already deleted")
			return nil
		}

		firewallID = p.state.CloudInfra.Civo.FirewallIDWorkerNodes

		_, err := p.client.DeleteFirewall(p.state.CloudInfra.Civo.FirewallIDWorkerNodes)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Civo.FirewallIDWorkerNodes = ""
	case consts.RoleDs:
		if len(p.state.CloudInfra.Civo.FirewallIDDatabaseNodes) == 0 {
			p.l.Print(p.ctx, "skipped firewall for datastore already deleted")
			return nil
		}

		firewallID = p.state.CloudInfra.Civo.FirewallIDDatabaseNodes

		_, err := p.client.DeleteFirewall(p.state.CloudInfra.Civo.FirewallIDDatabaseNodes)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Civo.FirewallIDDatabaseNodes = ""
	case consts.RoleLb:
		if len(p.state.CloudInfra.Civo.FirewallIDLoadBalancer) == 0 {
			p.l.Print(p.ctx, "skipped firewall for loadbalancer already deleted")
			return nil
		}

		firewallID = p.state.CloudInfra.Civo.FirewallIDLoadBalancer

		_, err := p.client.DeleteFirewall(p.state.CloudInfra.Civo.FirewallIDLoadBalancer)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Civo.FirewallIDLoadBalancer = ""
	}

	p.l.Success(p.ctx, "Deleted firewall", "firewallID", firewallID)
	return p.store.Write(p.state)
}

func (p *Provider) NewFirewall() error {

	name := <-p.chResName
	role := <-p.chRole

	p.l.Debug(p.ctx, "Printing", "Name", name)
	p.l.Debug(p.ctx, "Printing", "Role", role)

	createRules := false

	firewallConfig := &civogo.FirewallConfig{
		CreateRules: &createRules,
		Name:        name,
		Region:      p.Region,
		NetworkID:   p.state.CloudInfra.Civo.NetworkID,
	}

	netCidr := p.state.CloudInfra.Civo.NetworkCIDR
	kubernetesDistro := p.state.BootstrapProvider

	switch role {
	case consts.RoleCp:
		if len(p.state.CloudInfra.Civo.FirewallIDControlPlanes) != 0 {
			p.l.Print(p.ctx, "skipped firewall for controlplane found", p.state.CloudInfra.Civo.FirewallIDControlPlanes)
			return nil
		}

		firewallConfig.Rules = firewallRuleControlPlane(netCidr, kubernetesDistro)

	case consts.RoleWp:
		if len(p.state.CloudInfra.Civo.FirewallIDWorkerNodes) != 0 {
			p.l.Print(p.ctx, "skipped firewall for workerplane found", p.state.CloudInfra.Civo.FirewallIDWorkerNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleWorkerPlane(netCidr, kubernetesDistro)

	case consts.RoleDs:
		if len(p.state.CloudInfra.Civo.FirewallIDDatabaseNodes) != 0 {
			p.l.Print(p.ctx, "skipped firewall for datastore found", p.state.CloudInfra.Civo.FirewallIDDatabaseNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleDataStore(netCidr)

	case consts.RoleLb:
		if len(p.state.CloudInfra.Civo.FirewallIDLoadBalancer) != 0 {
			p.l.Print(p.ctx, "skipped firewall for loadbalancer found", p.state.CloudInfra.Civo.FirewallIDLoadBalancer)
			return nil
		}

		firewallConfig.Rules = firewallRuleLoadBalancer()
	}

	p.l.Debug(p.ctx, "Printing", "FirewallRule", firewallConfig.Rules)

	firew, err := p.client.NewFirewall(firewallConfig)
	if err != nil {
		return err
	}

	switch role {
	case consts.RoleCp:
		p.state.CloudInfra.Civo.FirewallIDControlPlanes = firew.ID
	case consts.RoleWp:
		p.state.CloudInfra.Civo.FirewallIDWorkerNodes = firew.ID
	case consts.RoleDs:
		p.state.CloudInfra.Civo.FirewallIDDatabaseNodes = firew.ID
	case consts.RoleLb:
		p.state.CloudInfra.Civo.FirewallIDLoadBalancer = firew.ID
	}

	p.l.Success(p.ctx, "Created firewall", "name", name)
	return p.store.Write(p.state)
}

func convertToProviderSpecific(_rules []providers.FirewallRule) []civogo.FirewallRule {
	var rules []civogo.FirewallRule

	for _, _r := range _rules {

		var protocol, action, direction string

		switch _r.Action {
		case consts.FirewallActionAllow:
			action = "allow"
		case consts.FirewallActionDeny:
			action = "deny"
		default:
			action = "allow"
		}

		switch _r.Protocol {
		case consts.FirewallActionTCP:
			protocol = "tcp"
		case consts.FirewallActionUDP:
			protocol = "udp"
		default:
			protocol = "tcp"
		}

		switch _r.Direction {
		case consts.FirewallActionIngress:
			direction = "ingress"
		case consts.FirewallActionEgress:
			direction = "egress"
		default:
			direction = "ingress"
		}

		rules = append(rules, civogo.FirewallRule{
			Direction: direction,
			Action:    action,
			Protocol:  protocol,

			Label:     _r.Description,
			Cidr:      []string{_r.Cidr},
			StartPort: _r.StartPort,
			EndPort:   _r.EndPort,
		})
	}

	return rules
}

func firewallRuleControlPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) []civogo.FirewallRule {

	return convertToProviderSpecific(
		providers.FirewallForControlplane_BASE(internalNetCidr, bootstrap),
	)
}

func firewallRuleWorkerPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) []civogo.FirewallRule {

	return convertToProviderSpecific(
		providers.FirewallForWorkerplane_BASE(internalNetCidr, bootstrap),
	)
}

func firewallRuleLoadBalancer() []civogo.FirewallRule {
	return convertToProviderSpecific(
		providers.FirewallForLoadBalancer_BASE(),
	)
}

func firewallRuleDataStore(internalNetCidr string) []civogo.FirewallRule {
	return convertToProviderSpecific(
		providers.FirewallForDataStore_BASE(internalNetCidr),
	)
}
