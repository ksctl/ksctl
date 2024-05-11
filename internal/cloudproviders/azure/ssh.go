package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/types"
)

// CreateUploadSSHKeyPair implements types.CloudFactory.
func (obj *AzureProvider) CreateUploadSSHKeyPair(storage types.StorageFactory) error {
	name := <-obj.chResName
	log.Debug("Printing", "name", name)

	if len(mainStateDocument.CloudInfra.Azure.B.SSHKeyName) != 0 {
		log.Print("skipped ssh key already created", "name", mainStateDocument.CloudInfra.Azure.B.SSHKeyName)
		return nil
	}

	err := helpers.CreateSSHKeyPair(log, mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}
	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: to.Ptr(obj.region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: to.Ptr(mainStateDocument.SSHKeyPair.PublicKey),
		},
	}

	log.Debug("Printing", "sshConfig", parameters)

	_, err = obj.client.CreateSSHKey(name, parameters, nil)
	if err != nil {
		return log.NewError(err.Error())
	}

	mainStateDocument.CloudInfra.Azure.B.SSHKeyName = name
	mainStateDocument.CloudInfra.Azure.B.SSHUser = "azureuser"

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("created the ssh key pair", "name", mainStateDocument.CloudInfra.Azure.B.SSHKeyName)

	return nil
}

// DelSSHKeyPair implements types.CloudFactory.
func (obj *AzureProvider) DelSSHKeyPair(storage types.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Azure.B.SSHKeyName) == 0 {
		log.Print("skipped ssh key already deleted", "name", mainStateDocument.CloudInfra.Azure.B.SSHKeyName)
		return nil
	}

	if _, err := obj.client.DeleteSSHKey(mainStateDocument.CloudInfra.Azure.B.SSHKeyName, nil); err != nil {
		return log.NewError(err.Error())
	}

	sshName := mainStateDocument.CloudInfra.Azure.B.SSHKeyName

	mainStateDocument.CloudInfra.Azure.B.SSHKeyName = ""
	mainStateDocument.CloudInfra.Azure.B.SSHUser = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("deleted the ssh key pair", "name", sshName)
	return nil
}
