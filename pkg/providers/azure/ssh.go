// Copyright 2024 Ksctl Authors
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

package azure

import (
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *AzureProvider) CreateUploadSSHKeyPair(storage types.StorageFactory) error {
	name := <-obj.chResName
	log.Debug(azureCtx, "Printing", "name", name)

	if len(mainStateDocument.CloudInfra.Azure.B.SSHKeyName) != 0 {
		log.Print(azureCtx, "skipped ssh key already created", "name", mainStateDocument.CloudInfra.Azure.B.SSHKeyName)
		return nil
	}

	err := helpers.CreateSSHKeyPair(azureCtx, log, mainStateDocument)
	if err != nil {
		return err
	}
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: utilities.Ptr(obj.region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: utilities.Ptr(mainStateDocument.SSHKeyPair.PublicKey),
		},
	}

	log.Debug(azureCtx, "Printing", "sshConfig", parameters)

	_, err = obj.client.CreateSSHKey(name, parameters, nil)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Azure.B.SSHKeyName = name
	mainStateDocument.CloudInfra.Azure.B.SSHUser = "azureuser"

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success(azureCtx, "created the ssh key pair", "name", mainStateDocument.CloudInfra.Azure.B.SSHKeyName)

	return nil
}

func (obj *AzureProvider) DelSSHKeyPair(storage types.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Azure.B.SSHKeyName) == 0 {
		log.Print(azureCtx, "skipped ssh key already deleted", "name", mainStateDocument.CloudInfra.Azure.B.SSHKeyName)
		return nil
	}

	if _, err := obj.client.DeleteSSHKey(mainStateDocument.CloudInfra.Azure.B.SSHKeyName, nil); err != nil {
		return err
	}

	sshName := mainStateDocument.CloudInfra.Azure.B.SSHKeyName

	mainStateDocument.CloudInfra.Azure.B.SSHKeyName = ""
	mainStateDocument.CloudInfra.Azure.B.SSHUser = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "deleted the ssh key pair", "name", sshName)
	return nil
}
