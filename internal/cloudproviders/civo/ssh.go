package civo

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"
)

// DelSSHKeyPair implements resources.CloudFactory.
func (obj *CivoProvider) DelSSHKeyPair(storage resources.StorageFactory) error {
	if len(mainStateDocument.CloudInfra.Civo.B.SSHID) == 0 {
		log.Print("skipped ssh keypair already deleted")
		return nil
	}

	_, err := obj.client.DeleteSSHKey(mainStateDocument.CloudInfra.Civo.B.SSHID)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("ssh keypair deleted", "sshID", mainStateDocument.CloudInfra.Civo.B.SSHID)

	mainStateDocument.CloudInfra.Civo.B.SSHID = ""
	mainStateDocument.CloudInfra.Civo.B.SSHUser = ""
	mainStateDocument.SSHKeyPair.PrivateKey, mainStateDocument.SSHKeyPair.PrivateKey = "", ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (obj *CivoProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	name := <-obj.chResName

	if len(mainStateDocument.CloudInfra.Civo.B.SSHID) != 0 {
		log.Print("skipped ssh keypair already uploaded")
		return nil
	}

	err := helpers.CreateSSHKeyPair(log, mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Debug("Printing", "keypair", mainStateDocument.SSHKeyPair.PublicKey)

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	if err := obj.uploadSSH(storage, name, mainStateDocument.SSHKeyPair.PublicKey); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("ssh keypair created and uploaded", "sshKeyPairName", name)
	return nil
}

func (obj *CivoProvider) uploadSSH(storage resources.StorageFactory, resName, pubKey string) error {
	sshResp, err := obj.client.NewSSHKey(resName, pubKey)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Civo.B.SSHID = sshResp.ID
	mainStateDocument.CloudInfra.Civo.B.SSHUser = "root"

	log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.B.SSHID", mainStateDocument.CloudInfra.Civo.B.SSHID, "mainStateDocument.CloudInfra.Civo.B.SSHUser", mainStateDocument.CloudInfra.Civo.B.SSHUser)

	return storage.Write(mainStateDocument)
}
