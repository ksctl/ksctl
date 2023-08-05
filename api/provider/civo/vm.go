package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// NewVM implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewVM(state resources.StorageInfrastructure) error {
	fmt.Printf("[civo] creating %s VM...\n", obj.Metadata.ResName)
	return nil
}

// DelVM implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelVM(state resources.StorageInfrastructure) error {
	fmt.Printf("[civo] delete %s VM...\n", obj.Metadata.ResName)
	return nil
}
