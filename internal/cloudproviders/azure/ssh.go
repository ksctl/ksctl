package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
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

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, log, CloudAzure, clusterDirName)
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
	azureCloudState.SSHPrivateKeyLoc = utils.GetPath(UtilSSHPath, CloudAzure, clusterType, clusterDirName)

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

	azureCloudState.SSHKeyName = ""
	azureCloudState.SSHUser = ""
	azureCloudState.SSHPrivateKeyLoc = ""

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("deleted the ssh key pair", "name", azureCloudState.SSHKeyName)
	return nil
}
