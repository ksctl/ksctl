package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *AwsProvider) CreateUploadSSHKeyPair(storage types.StorageFactory) error {

	name := <-obj.chResName
	log.Debug(awsCtx, "Printing", "name", name)

	if len(mainStateDocument.CloudInfra.Aws.B.SSHKeyName) != 0 {
		log.Success(awsCtx, "skipped ssh key already created", "name", mainStateDocument.CloudInfra.Aws.B.SSHKeyName)
		return nil
	}

	err := helpers.CreateSSHKeyPair(awsCtx, log, mainStateDocument)
	if err != nil {
		return err
	}

	parameter := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(name),
		PublicKeyMaterial: []byte(mainStateDocument.SSHKeyPair.PublicKey),
	}

	if err := obj.client.ImportKeyPair(awsCtx, parameter); err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.B.SSHKeyName = name
	mainStateDocument.CloudInfra.Aws.B.SSHUser = "ubuntu"

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success(awsCtx, "created the ssh key pair", "name", mainStateDocument.CloudInfra.Aws.B.SSHKeyName)

	return nil
}

func (obj *AwsProvider) DelSSHKeyPair(storage types.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Aws.B.SSHKeyName) == 0 {
		log.Success(awsCtx, "skipped already deleted the ssh key", "name", mainStateDocument.CloudInfra.Aws.B.SSHKeyName)
	} else {
		err := obj.client.DeleteSSHKey(awsCtx, mainStateDocument.CloudInfra.Aws.B.SSHKeyName)
		if err != nil {
			return err
		}

		sshName := mainStateDocument.CloudInfra.Aws.B.SSHKeyName

		mainStateDocument.CloudInfra.Aws.B.SSHKeyName = ""
		mainStateDocument.CloudInfra.Aws.B.SSHUser = ""

		if err := storage.Write(mainStateDocument); err != nil {
			return err
		}

		log.Success(awsCtx, "deleted the ssh key pair", "name", sshName)
	}

	return nil
}
