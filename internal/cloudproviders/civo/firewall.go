package civo

import (
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

// DelFirewall implements resources.CloudFactory.
func (obj *CivoProvider) DelFirewall(storage resources.StorageFactory) error {
	role := obj.metadata.role
	obj.mxRole.Unlock()

	log.Debug("Printing", "Role", role)

	var firewallID string
	switch role {
	case consts.RoleCp:
		if len(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode) == 0 {
			log.Print("skipped firewall for controlplane already deleted")
			return nil
		}
		firewallID = civoCloudState.NetworkIDs.FirewallIDControlPlaneNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode)
		if err != nil {
			return log.NewError(err.Error())
		}

		civoCloudState.NetworkIDs.FirewallIDControlPlaneNode = ""

	case consts.RoleWp:
		if len(civoCloudState.NetworkIDs.FirewallIDWorkerNode) == 0 {
			log.Print("skipped firewall for workerplane already deleted")
			return nil
		}

		firewallID = civoCloudState.NetworkIDs.FirewallIDWorkerNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDWorkerNode)
		if err != nil {
			return log.NewError(err.Error())
		}

		civoCloudState.NetworkIDs.FirewallIDWorkerNode = ""
	case consts.RoleDs:
		if len(civoCloudState.NetworkIDs.FirewallIDDatabaseNode) == 0 {
			log.Print("skipped firewall for datastore already deleted")
			return nil
		}

		firewallID = civoCloudState.NetworkIDs.FirewallIDDatabaseNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDDatabaseNode)
		if err != nil {
			return log.NewError(err.Error())
		}

		civoCloudState.NetworkIDs.FirewallIDDatabaseNode = ""
	case consts.RoleLb:
		if len(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode) == 0 {
			log.Print("skipped firewall for loadbalancer already deleted")
			return nil
		}

		firewallID = civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode)
		if err != nil {
			return log.NewError(err.Error())
		}

		civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode = ""
	}

	path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

	log.Success("Deleted firewall", "firewallID", firewallID)
	return saveStateHelper(storage, path)
}

// NewFirewall implements resources.CloudFactory.
func (obj *CivoProvider) NewFirewall(storage resources.StorageFactory) error {

	name := obj.metadata.resName
	role := obj.metadata.role
	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	log.Debug("Printing", "Name", name)
	log.Debug("Printing", "Role", role)

	firewallConfig := &civogo.FirewallConfig{
		Name:      name,
		Region:    obj.region,
		NetworkID: civoCloudState.NetworkIDs.NetworkID,
	}

	switch role {
	case consts.RoleCp:
		if len(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode) != 0 {
			log.Print("skipped firewall for controlplane found", civoCloudState.NetworkIDs.FirewallIDControlPlaneNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleControlPlane()

	case consts.RoleWp:
		if len(civoCloudState.NetworkIDs.FirewallIDWorkerNode) != 0 {
			log.Print("skipped firewall for workerplane found", civoCloudState.NetworkIDs.FirewallIDWorkerNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleWorkerPlane()

	case consts.RoleDs:
		if len(civoCloudState.NetworkIDs.FirewallIDDatabaseNode) != 0 {
			log.Print("skipped firewall for datastore found", civoCloudState.NetworkIDs.FirewallIDDatabaseNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleDataStore()

	case consts.RoleLb:
		if len(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode) != 0 {
			log.Print("skipped firewall for loadbalancer found", civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode)
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
		civoCloudState.NetworkIDs.FirewallIDControlPlaneNode = firew.ID
	case consts.RoleWp:
		civoCloudState.NetworkIDs.FirewallIDWorkerNode = firew.ID
	case consts.RoleDs:
		civoCloudState.NetworkIDs.FirewallIDDatabaseNode = firew.ID
	case consts.RoleLb:
		civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode = firew.ID
	}

	path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)
	log.Debug("Printing", "path", path, "cloudState.networkIDS", civoCloudState.NetworkIDs)

	log.Success("Created firewall", "name", name)
	return saveStateHelper(storage, path)
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
