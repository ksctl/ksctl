package civo

import (
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// DelSSHKeyPair implements resources.CloudFactory.
func (obj *CivoProvider) DelSSHKeyPair(storage resources.StorageFactory) error {
	if len(civoCloudState.SSHID) == 0 {
		log.Print("skipped ssh keypair already deleted")
		return nil
	}

	_, err := obj.client.DeleteSSHKey(civoCloudState.SSHID)
	if err != nil {
		return log.NewError(err.Error())
	}
	path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

	log.Success("ssh keypair deleted", "sshID", civoCloudState.SSHID)

	civoCloudState.SSHID = ""
	civoCloudState.SSHPrivateKeyLoc = ""
	civoCloudState.SSHUser = ""

	if err := saveStateHelper(storage, path); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (obj *CivoProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	name := obj.metadata.resName
	obj.mxName.Unlock()

	if len(civoCloudState.SSHID) != 0 {
		log.Print("skipped ssh keypair already uploaded")
		return nil
	}

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, log, consts.CloudCivo, clusterDirName)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Debug("Printing", "keyapir", keyPairToUpload)
	if err := obj.uploadSSH(storage, name, keyPairToUpload); err != nil {
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

	civoCloudState.SSHID = sshResp.ID
	civoCloudState.SSHUser = "root"
	civoCloudState.SSHPrivateKeyLoc = utils.GetPath(consts.UtilSSHPath, consts.CloudCivo, clusterType, clusterDirName)

	log.Debug("Printing", "civoCloudState.SSHID", civoCloudState.SSHID, "civoCloudState.SSHUser", civoCloudState.SSHUser, "civoCloudState.SSHPrivateKeyLoc", civoCloudState.SSHPrivateKeyLoc)

	path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

	return saveStateHelper(storage, path)
}
