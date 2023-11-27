package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (obj *AzureProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	name := obj.metadata.resName
	obj.mxName.Unlock()
	log.Debug("Printing", "name", name)

	if len(azureCloudState.SSHKeyName) != 0 {
		log.Print("skipped ssh key already created", "name", azureCloudState.SSHKeyName)
		return nil
	}

	keyPairToUpload, err := helpers.CreateSSHKeyPair(storage, log, consts.CloudAzure, clusterDirName)
	if err != nil {
		return log.NewError(err.Error())
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: to.Ptr(obj.region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: to.Ptr(keyPairToUpload),
		},
	}

	log.Debug("Printing", "sshConfig", parameters)

	_, err = obj.client.CreateSSHKey(name, parameters, nil)
	if err != nil {
		return log.NewError(err.Error())
	}

	azureCloudState.SSHKeyName = name
	azureCloudState.SSHUser = "azureuser"
	azureCloudState.SSHPrivateKeyLoc = helpers.GetPath(consts.UtilSSHPath, consts.CloudAzure, clusterType, clusterDirName)

	log.Debug("Printing", "azureCloudState.SSHKeyName", azureCloudState.SSHKeyName, "azureCloudState.SSHPrivateKeyLoc", azureCloudState.SSHPrivateKeyLoc, "azureCloudState.SSHUser", azureCloudState.SSHUser)

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("created the ssh key pair", "name", azureCloudState.SSHKeyName)

	return nil
}

// DelSSHKeyPair implements resources.CloudFactory.
func (obj *AzureProvider) DelSSHKeyPair(storage resources.StorageFactory) error {

	if len(azureCloudState.SSHKeyName) == 0 {
		log.Print("skipped ssh key already deleted", "name", azureCloudState.SSHKeyName)
		return nil
	}

	if _, err := obj.client.DeleteSSHKey(azureCloudState.SSHKeyName, nil); err != nil {
		return log.NewError(err.Error())
	}

	sshName := azureCloudState.SSHKeyName

	azureCloudState.SSHKeyName = ""
	azureCloudState.SSHUser = ""
	azureCloudState.SSHPrivateKeyLoc = ""

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("deleted the ssh key pair", "name", sshName)
	return nil
}
