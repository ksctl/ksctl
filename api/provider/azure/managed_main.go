package azure

import "github.com/kubesimplify/ksctl/api/resources"

// DelManagedCluster implements resources.CloudFactory.
func (*AzureProvider) DelManagedCluster(state resources.StorageFactory) error {
	panic("unimplemented")
}

// NewManagedCluster implements resources.CloudFactory.
func (*AzureProvider) NewManagedCluster(state resources.StorageFactory, noOfNodes int) error {
	panic("unimplemented")
}
