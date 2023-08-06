package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// NewVM implements resources.CloudFactory.
func (obj *CivoProvider) NewVM(state resources.StorageFactory) error {
	fmt.Printf("[civo] creating %s VM...\n", obj.Metadata.ResName)
	return nil
}

// DelVM implements resources.CloudFactory.
func (obj *CivoProvider) DelVM(state resources.StorageFactory) error {
	fmt.Printf("[civo] delete %s VM...\n", obj.Metadata.ResName)
	return nil
}
