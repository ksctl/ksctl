package azure

import "github.com/kubesimplify/ksctl/api/resources"

// DelNetwork implements resources.CloudFactory.
func (*AzureProvider) DelNetwork(state resources.StorageFactory) error {
	panic("unimplemented")
}

// NewNetwork implements resources.CloudFactory.
func (*AzureProvider) NewNetwork(state resources.StorageFactory) error {
	panic("unimplemented")
}
