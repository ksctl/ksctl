package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// NewNetwork implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewNetwork(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] Creating %s network...\n", obj.Metadata.ResName)
	return nil
}

// DelNetwork implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelNetwork(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] delete %s network...\n", obj.Metadata.ResName)
	return nil
}
