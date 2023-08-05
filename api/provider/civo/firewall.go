package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// DelFirewall implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelFirewall(storage resources.StorageInfrastructure) error {
	fmt.Printf("[civo] delete %s Firewall....\n", obj.Metadata.ResName)
	return nil
}

// NewFirewall implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewFirewall(storage resources.StorageInfrastructure) error {
	fmt.Printf("[civo] create %s Firewall....\n", obj.Metadata.ResName)
	return nil
}

// ///////////// REFER TO KUBERNETES DOCS for the ports to be opened///////////////
func firewallRuleControlPlane() {}

func firewallRuleWorkerPlane() {}

func firewallRuleLoadBalancer() {}

func firewallRuleDataStore() {}
