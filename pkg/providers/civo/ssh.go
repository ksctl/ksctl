package civo

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *CivoProvider) DelSSHKeyPair(storage types.StorageFactory) error {
	if len(mainStateDocument.CloudInfra.Civo.B.SSHID) == 0 {
		log.Print(civoCtx, "skipped ssh keypair already deleted")
		return nil
	}

	_, err := obj.client.DeleteSSHKey(mainStateDocument.CloudInfra.Civo.B.SSHID)
	if err != nil {
		return err
	}

	log.Success(civoCtx, "ssh keypair deleted", "sshID", mainStateDocument.CloudInfra.Civo.B.SSHID)

	mainStateDocument.CloudInfra.Civo.B.SSHID = ""
	mainStateDocument.CloudInfra.Civo.B.SSHUser = ""
	mainStateDocument.SSHKeyPair.PrivateKey, mainStateDocument.SSHKeyPair.PrivateKey = "", ""

	return storage.Write(mainStateDocument)
}

func (obj *CivoProvider) CreateUploadSSHKeyPair(storage types.StorageFactory) error {
	name := <-obj.chResName

	if len(mainStateDocument.CloudInfra.Civo.B.SSHID) != 0 {
		log.Print(civoCtx, "skipped ssh keypair already uploaded")
		return nil
	}

	err := helpers.CreateSSHKeyPair(civoCtx, log, mainStateDocument)
	if err != nil {
		return err
	}
	log.Debug(civoCtx, "Printing", "keypair", mainStateDocument.SSHKeyPair.PublicKey)

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	if err := obj.uploadSSH(storage, name, mainStateDocument.SSHKeyPair.PublicKey); err != nil {
		return err
	}
	log.Success(civoCtx, "ssh keypair created and uploaded", "sshKeyPairName", name)
	return nil
}

func (obj *CivoProvider) uploadSSH(storage types.StorageFactory, resName, pubKey string) error {
	sshResp, err := obj.client.NewSSHKey(resName, pubKey)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Civo.B.SSHID = sshResp.ID
	mainStateDocument.CloudInfra.Civo.B.SSHUser = "root"

	log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.B.SSHID", mainStateDocument.CloudInfra.Civo.B.SSHID, "mainStateDocument.CloudInfra.Civo.B.SSHUser", mainStateDocument.CloudInfra.Civo.B.SSHUser)

	return storage.Write(mainStateDocument)
}
