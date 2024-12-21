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
