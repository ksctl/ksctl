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

	sshClient, err := obj.Client.SSHKeyClient()
	if err != nil {
		return err
	}

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, utils.CLOUD_AZURE, clusterDirName)
	if err != nil {
		return err
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: to.Ptr(obj.Region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: to.Ptr(keyPairToUpload),
		},
	}

	_, err = obj.Client.CreateSSHKey(ctx, sshClient, azureCloudState.ResourceGroupName,
		obj.Metadata.ResName, parameters, nil)

	azureCloudState.SSHKeyName = obj.Metadata.ResName

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

	sshClient, err := obj.Client.SSHKeyClient()
	if err != nil {
		return err
	}
	_, err = obj.Client.DeleteSSHKey(ctx, sshClient, azureCloudState.ResourceGroupName,
		azureCloudState.SSHKeyName, nil)
	if err != nil {
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
