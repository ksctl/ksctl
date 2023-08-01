package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// DelFirewall implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelFirewall(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] delete %s Firewall....\n", obj.Metadata.ResName)
	return nil
}

// NewFirewall implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewFirewall(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] create %s Firewall....\n", obj.Metadata.ResName)
	return nil
}
