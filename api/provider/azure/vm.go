package azure

import (
	"errors"
	"github.com/kubesimplify/ksctl/api/resources"
)

// DelVM implements resources.CloudFactory.
func (*AzureProvider) DelVM(state resources.StorageFactory, indexNo int) error {
	panic("unimplemented")
}

// NewVM implements resources.CloudFactory.
func (*AzureProvider) NewVM(state resources.StorageFactory, indexNo int) error {
	return errors.New("unimplemented")
}
