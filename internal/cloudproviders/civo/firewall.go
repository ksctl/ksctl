package civo

import (
	"github.com/civo/civogo"
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

	switch role {
	case consts.RoleCp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes) != 0 {
			log.Print("skipped firewall for controlplane found", mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes)
			return nil
		}

		firewallConfig.Rules = firewallRuleControlPlane(netCidr)

	case consts.RoleWp:
		if len(mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes) != 0 {
			log.Print("skipped firewall for workerplane found", mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes)
			return nil
		}

		firewallConfig.Rules = firewallRuleWorkerPlane(netCidr)

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
func firewallRuleControlPlane(internalNetCidr string) []civogo.FirewallRule {
	return []civogo.FirewallRule{
		// add ingress for the nodePort access
		{
			Label:     "K3s supervisor and Kubernetes API Server",
			Protocol:  "tcp",
			StartPort: "6443",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Required only for Flannel VXLAN",
			Protocol:  "tcp",
			StartPort: "8472",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Kubelet metrics",
			Protocol:  "tcp",
			StartPort: "10250",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "NodePort",
			Protocol:  "tcp",
			StartPort: "30000",
			EndPort:   "35000",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Required only ksctl bootstrappnig",
			Protocol:  "tcp",
			StartPort: "22",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Protocol:  "tcp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
		{
			Protocol:  "udp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
	}
}

func firewallRuleWorkerPlane(internalNetCidr string) []civogo.FirewallRule {
	return []civogo.FirewallRule{
		{
			Label:     "Required only for Flannel VXLAN",
			Protocol:  "udp",
			StartPort: "8472",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Kubelet metrics",
			Protocol:  "tcp",
			StartPort: "10250",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Required only ksctl bootstrappnig",
			Protocol:  "tcp",
			StartPort: "22",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Protocol:  "tcp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
		{
			Protocol:  "udp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
	}
}

func firewallRuleLoadBalancer(internalNetCidr string) []civogo.FirewallRule {
	return []civogo.FirewallRule{
		{
			Label:     "K3s supervisor and Kubernetes API Server",
			Protocol:  "tcp",
			StartPort: "6443",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "NodePort",
			Protocol:  "tcp",
			StartPort: "30000",
			EndPort:   "35000",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Required only ksctl bootstrappnig",
			Protocol:  "tcp",
			StartPort: "22",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Protocol:  "tcp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
		{
			Protocol:  "udp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
	}
}

func firewallRuleDataStore(internalNetCidr string) []civogo.FirewallRule {
	return []civogo.FirewallRule{
		{
			Label:     "Required only for HA with external etcd",
			Protocol:  "tcp",
			StartPort: "2379",
			EndPort:   "2380",
			Cidr:      []string{internalNetCidr},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Label:     "Required only ksctl bootstrappnig",
			Protocol:  "tcp",
			StartPort: "22",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "ingress",
		},
		{
			Protocol:  "tcp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
		{
			Protocol:  "udp",
			StartPort: "1",
			EndPort:   "65535",
			Cidr:      []string{"0.0.0.0/0"},
			Action:    "allow",
			Direction: "egress",
		},
	}
}
