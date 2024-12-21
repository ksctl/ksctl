// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
