package azure

import "github.com/kubesimplify/ksctl/api/resources"

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (cloud *AzureProvider) CreateUploadSSHKeyPair(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudFactory.
func (*AzureProvider) DelSSHKeyPair(state resources.StorageFactory) error {
	panic("unimplemented")
}
