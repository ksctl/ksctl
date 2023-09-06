package civo

import (
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// DelFirewall implements resources.CloudFactory.
func (obj *CivoProvider) DelFirewall(storage resources.StorageFactory) error {
	role := obj.metadata.role
	obj.mxRole.Unlock()

	var firewallID string
	switch role {
	case utils.ROLE_CP:
		if len(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode) == 0 {
			storage.Logger().Success("[skip] firewall for controlplane already deleted")
			return nil
		}
		firewallID = civoCloudState.NetworkIDs.FirewallIDControlPlaneNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode)
		if err != nil {
			return err
		}

		civoCloudState.NetworkIDs.FirewallIDControlPlaneNode = ""

	case utils.ROLE_WP:
		if len(civoCloudState.NetworkIDs.FirewallIDWorkerNode) == 0 {
			storage.Logger().Success("[skip] firewall for workerplane already deleted")
			return nil
		}

		firewallID = civoCloudState.NetworkIDs.FirewallIDWorkerNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDWorkerNode)
		if err != nil {
			return err
		}

		civoCloudState.NetworkIDs.FirewallIDWorkerNode = ""
	case utils.ROLE_DS:
		if len(civoCloudState.NetworkIDs.FirewallIDDatabaseNode) == 0 {
			storage.Logger().Success("[skip] firewall for datastore already deleted")
			return nil
		}

		firewallID = civoCloudState.NetworkIDs.FirewallIDDatabaseNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDDatabaseNode)
		if err != nil {
			return err
		}

		civoCloudState.NetworkIDs.FirewallIDDatabaseNode = ""
	case utils.ROLE_LB:
		if len(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode) == 0 {
			storage.Logger().Success("[skip] firewall for loadbalancer already deleted")
			return nil
		}

		firewallID = civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode

		_, err := obj.client.DeleteFirewall(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode)
		if err != nil {
			return err
		}

		civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode = ""

	}

	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)

	storage.Logger().Success("[civo] Deleted firewall", firewallID)
	return saveStateHelper(storage, path)
}

// NewFirewall implements resources.CloudFactory.
func (obj *CivoProvider) NewFirewall(storage resources.StorageFactory) error {

	name := obj.metadata.resName
	role := obj.metadata.role
	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	firewallConfig := &civogo.FirewallConfig{
		Name:      name,
		Region:    obj.region,
		NetworkID: civoCloudState.NetworkIDs.NetworkID,
	}

	switch role {
	case utils.ROLE_CP:
		if len(civoCloudState.NetworkIDs.FirewallIDControlPlaneNode) != 0 {
			storage.Logger().Success("[skip] firewall for controlplane found", civoCloudState.NetworkIDs.FirewallIDControlPlaneNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleControlPlane()

	case utils.ROLE_WP:
		if len(civoCloudState.NetworkIDs.FirewallIDWorkerNode) != 0 {
			storage.Logger().Success("[skip] firewall for workerplane found", civoCloudState.NetworkIDs.FirewallIDWorkerNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleWorkerPlane()

	case utils.ROLE_DS:
		if len(civoCloudState.NetworkIDs.FirewallIDDatabaseNode) != 0 {
			storage.Logger().Success("[skip] firewall for datastore found", civoCloudState.NetworkIDs.FirewallIDDatabaseNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleDataStore()

	case utils.ROLE_LB:
		if len(civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode) != 0 {
			storage.Logger().Success("[skip] firewall for loadbalancer found", civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode)
			return nil
		}

		firewallConfig.Rules = firewallRuleLoadBalancer()

	}

	firew, err := obj.client.NewFirewall(firewallConfig)
	if err != nil {
		return err
	}

	switch role {
	case utils.ROLE_CP:
		civoCloudState.NetworkIDs.FirewallIDControlPlaneNode = firew.ID
	case utils.ROLE_WP:
		civoCloudState.NetworkIDs.FirewallIDWorkerNode = firew.ID
	case utils.ROLE_DS:
		civoCloudState.NetworkIDs.FirewallIDDatabaseNode = firew.ID
	case utils.ROLE_LB:
		civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode = firew.ID
	}

	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)

	storage.Logger().Success("[civo] Created firewall", name)
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
