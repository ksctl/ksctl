package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"
)

func (obj *AwsProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {

	name := <-obj.chResName
	log.Debug("Printing", "name", name)

	if len(mainStateDocument.CloudInfra.Aws.B.SSHKeyName) != 0 {
		log.Success("[skip] ssh key already created", mainStateDocument.CloudInfra.Aws.B.SSHKeyName)
		return nil
	}

	err := helpers.CreateSSHKeyPair(log, mainStateDocument)
	if err != nil {
		return log.NewError("Error creating ssh key pair", "error", err)
	}

	parameter := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(name),
		PublicKeyMaterial: []byte(mainStateDocument.SSHKeyPair.PublicKey),
	}

	if err := obj.client.ImportKeyPair(context.Background(), parameter); err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.B.SSHKeyName = name
	mainStateDocument.CloudInfra.Aws.B.SSHUser = "ubuntu"

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("created the ssh key pair", "name", mainStateDocument.CloudInfra.Aws.B.SSHKeyName)

	return nil

}

func (obj *AwsProvider) DelSSHKeyPair(storage resources.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Aws.B.SSHKeyName) == 0 {
		log.Success("[skip] already deleted the ssh key", "", mainStateDocument.CloudInfra.Aws.B.SSHKeyName)
	} else {
		err := obj.client.DeleteSSHKey(context.Background(), mainStateDocument.CloudInfra.Aws.B.SSHKeyName)
		if err != nil {
			return err
		}

		sshName := mainStateDocument.CloudInfra.Aws.B.SSHKeyName

		mainStateDocument.CloudInfra.Aws.B.SSHKeyName = ""
		mainStateDocument.CloudInfra.Aws.B.SSHUser = ""

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError("Error writing to storage", "error", err)
		}

		log.Success("deleted the ssh key pair", "name", sshName)
	}

	return nil
}
