package civo

import (
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
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

	firewallConfig := &civogo.FirewallConfig{
		Name:      name,
		Region:    obj.region,
		NetworkID: mainStateDocument.CloudInfra.Civo.NetworkID,
	}

	switch role {
	case consts.RoleCp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes) != 0 {
			log.Print("skipped firewall for controlplane found", mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes)
			return nil
		}

		firewallConfig.Rules = firewallRuleControlPlane()

	case consts.RoleWp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes) != 0 {
			log.Print("skipped firewall for workerplane found", mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleWorkerPlane()

	case consts.RoleDs:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes) != 0 {
			log.Print("skipped firewall for datastore found", mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleDataStore()

	case consts.RoleLb:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer) != 0 {
			log.Print("skipped firewall for loadbalancer found", mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer)
			return nil
		}

		firewallConfig.Rules = firewallRuleLoadBalancer()

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

func firewallRuleControlPlane() []civogo.FirewallRule {
	return nil
}

func firewallRuleWorkerPlane() []civogo.FirewallRule {
	return nil
}

func firewallRuleLoadBalancer() []civogo.FirewallRule {
	return nil
}

func firewallRuleDataStore() []civogo.FirewallRule {
	return nil
}
