package civo

import (
	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// DelFirewall implements resources.CloudFactory.
func (obj *CivoProvider) DelFirewall(storage resources.StorageFactory) error {
	role := <-obj.chRole

	log.Debug("Printing", "Role", role)

	var firewallID string
	switch role {
	case consts.RoleCp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes) == 0 {
			log.Print("skipped firewall for controlplane already deleted")
			return nil
		}
		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes

		_, err := obj.client.DeleteFirewall(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes)
		if err != nil {
			return log.NewError(err.Error())
		}

		mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes = ""

	case consts.RoleWp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes) == 0 {
			log.Print("skipped firewall for workerplane already deleted")
			return nil
		}

		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes

		_, err := obj.client.DeleteFirewall(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes)
		if err != nil {
			return log.NewError(err.Error())
		}

		mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes = ""
	case consts.RoleDs:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes) == 0 {
			log.Print("skipped firewall for datastore already deleted")
			return nil
		}

		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes

		_, err := obj.client.DeleteFirewall(mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes)
		if err != nil {
			return log.NewError(err.Error())
		}

		mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes = ""
	case consts.RoleLb:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer) == 0 {
			log.Print("skipped firewall for loadbalancer already deleted")
			return nil
		}

		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer

		_, err := obj.client.DeleteFirewall(mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer)
		if err != nil {
			return log.NewError(err.Error())
		}

		mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer = ""
	}

	log.Success("Deleted firewall", "firewallID", firewallID)
	return storage.Write(mainStateDocument)
}

// NewFirewall implements resources.CloudFactory.
func (obj *CivoProvider) NewFirewall(storage resources.StorageFactory) error {

	name := <-obj.chResName
	role := <-obj.chRole

	log.Debug("Printing", "Name", name)
	log.Debug("Printing", "Role", role)

	createRules := false

	firewallConfig := &civogo.FirewallConfig{
		CreateRules: &createRules,
		Name:        name,
		Region:      obj.region,
		NetworkID:   mainStateDocument.CloudInfra.Civo.NetworkID,
	}

	netCidr := mainStateDocument.CloudInfra.Civo.NetworkCIDR
	kubernetesDistro := consts.KsctlKubernetes(mainStateDocument.CloudInfra.Civo.B.KubernetesDistro)

	switch role {
	case consts.RoleCp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes) != 0 {
			log.Print("skipped firewall for controlplane found", mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes)
			return nil
		}

		firewallConfig.Rules = firewallRuleControlPlane(netCidr, kubernetesDistro)

	case consts.RoleWp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes) != 0 {
			log.Print("skipped firewall for workerplane found", mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleWorkerPlane(netCidr, kubernetesDistro)

	case consts.RoleDs:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes) != 0 {
			log.Print("skipped firewall for datastore found", mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleDataStore(netCidr)

	case consts.RoleLb:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer) != 0 {
			log.Print("skipped firewall for loadbalancer found", mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer)
			return nil
		}

		firewallConfig.Rules = firewallRuleLoadBalancer(netCidr)
	}

	log.Debug("Printing", "FirewallRule", firewallConfig.Rules)

	firew, err := obj.client.NewFirewall(firewallConfig)
	if err != nil {
		return log.NewError(err.Error())
	}

	switch role {
	case consts.RoleCp:
		mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes = firew.ID
	case consts.RoleWp:
		mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes = firew.ID
	case consts.RoleDs:
		mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes = firew.ID
	case consts.RoleLb:
		mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer = firew.ID
	}

	log.Success("Created firewall", "name", name)
	return storage.Write(mainStateDocument)
}

// need the CIdr range of the network so that internal network can be securied
func firewallRuleControlPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) []civogo.FirewallRule {

	return convertToProviderSpecific(
		helpers.FirewallForControlplane_BASE(internalNetCidr, bootstrap),
	)
}

func firewallRuleWorkerPlane(internalNetCidr string, bootstrap consts.KsctlKubernetes) []civogo.FirewallRule {

	return convertToProviderSpecific(
		helpers.FirewallForWorkerplane_BASE(internalNetCidr, bootstrap),
	)
}

func firewallRuleLoadBalancer(internalNetCidr string) []civogo.FirewallRule {
	return convertToProviderSpecific(
		helpers.FirewallForLoadBalancer_BASE(),
	)
}

func convertToProviderSpecific(_rules []helpers.FirewallRule) []civogo.FirewallRule {
	rules := []civogo.FirewallRule{}

	for _, _r := range _rules {

		var protocol, action, direction string

		switch _r.Action {
		case consts.FirewallActionAllow:
			action = "allow"
		case consts.FirewallActionDeny:
			action = "deny"
		}

		switch _r.Protocol {
		case consts.FirewallActionTCP:
			protocol = "tcp"
		case consts.FirewallActionUDP:
			protocol = "udp"
		}

		switch _r.Direction {
		case consts.FirewallActionIngress:
			protocol = "ingress"
		case consts.FirewallActionEgress:
			protocol = "egress"
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

func firewallRuleDataStore(internalNetCidr string) []civogo.FirewallRule {
	return convertToProviderSpecific(
		helpers.FirewallForDataStore_BASE(internalNetCidr),
	)
}
