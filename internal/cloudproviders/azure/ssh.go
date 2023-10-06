package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (obj *AzureProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	name := obj.metadata.resName
	obj.mxName.Unlock()

	if len(azureCloudState.SSHKeyName) != 0 {
		storage.Logger().Success("[skip] ssh key already created", azureCloudState.SSHKeyName)
		return nil
	}

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, CLOUD_AZURE, clusterDirName)
	if err != nil {
		return err
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: to.Ptr(obj.region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: to.Ptr(keyPairToUpload),
		},
	}

	_, err = obj.client.CreateSSHKey(name, parameters, nil)

	azureCloudState.SSHKeyName = name
	azureCloudState.SSHUser = "azureuser"
	azureCloudState.SSHPrivateKeyLoc = utils.GetPath(SSH_PATH, CLOUD_AZURE, clusterType, clusterDirName)

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] created the ssh key pair", azureCloudState.SSHKeyName)

	return nil
}

// DelSSHKeyPair implements resources.CloudFactory.
func (obj *AzureProvider) DelSSHKeyPair(storage resources.StorageFactory) error {

	if len(azureCloudState.SSHKeyName) == 0 {
		storage.Logger().Success("[skip] ssh key already deleted", azureCloudState.SSHKeyName)
		return nil
	}

	if _, err := obj.client.DeleteSSHKey(azureCloudState.SSHKeyName, nil); err != nil {
		return err
	}

	azureCloudState.SSHKeyName = ""
	azureCloudState.SSHUser = ""
	azureCloudState.SSHPrivateKeyLoc = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[azure] deleted the ssh key pair", azureCloudState.SSHKeyName)
	return nil
}
