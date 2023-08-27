package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (obj *AzureProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {

	if len(azureCloudState.SSHKeyName) != 0 {
		storage.Logger().Success("[skip] ssh key already created", azureCloudState.SSHKeyName)
		return nil
	}

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, utils.CLOUD_AZURE, clusterDirName)
	if err != nil {
		return err
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: to.Ptr(obj.region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: to.Ptr(keyPairToUpload),
		},
	}

	_, err = obj.client.CreateSSHKey(obj.metadata.resName, parameters, nil)

	azureCloudState.SSHKeyName = obj.metadata.resName

	azureCloudState.SSHUser = "azureuser"
	azureCloudState.SSHPrivateKeyLoc = utils.GetPath(utils.SSH_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName)

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
