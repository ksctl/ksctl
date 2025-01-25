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
	"github.com/ksctl/ksctl/v2/pkg/ssh"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func (p *Provider) CreateUploadSSHKeyPair() error {
	name := <-p.chResName
	p.l.Debug(p.ctx, "Printing", "name", name)

	if len(p.state.CloudInfra.Azure.B.SSHKeyName) != 0 {
		p.l.Print(p.ctx, "skipped ssh key already created", "name", p.state.CloudInfra.Azure.B.SSHKeyName)
		return nil
	}

	err := ssh.CreateSSHKeyPair(p.ctx, p.l, p.state)
	if err != nil {
		return err
	}
	if err := p.store.Write(p.state); err != nil {
		return err
	}

	parameters := armcompute.SSHPublicKeyResource{
		Location: utilities.Ptr(p.Region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{
			PublicKey: utilities.Ptr(p.state.SSHKeyPair.PublicKey),
		},
	}

	p.l.Debug(p.ctx, "Printing", "sshConfig", parameters)

	_, err = p.client.CreateSSHKey(name, parameters, nil)
	if err != nil {
		return err
	}

	p.state.CloudInfra.Azure.B.SSHKeyName = name
	p.state.CloudInfra.Azure.B.SSHUser = "azureuser"

	if err := p.store.Write(p.state); err != nil {
		return err
	}
	p.l.Success(p.ctx, "created the ssh key pair", "name", p.state.CloudInfra.Azure.B.SSHKeyName)

	return nil
}

func (p *Provider) DelSSHKeyPair() error {

	if len(p.state.CloudInfra.Azure.B.SSHKeyName) == 0 {
		p.l.Print(p.ctx, "skipped ssh key already deleted", "name", p.state.CloudInfra.Azure.B.SSHKeyName)
		return nil
	}

	if _, err := p.client.DeleteSSHKey(p.state.CloudInfra.Azure.B.SSHKeyName, nil); err != nil {
		return err
	}

	sshName := p.state.CloudInfra.Azure.B.SSHKeyName

	p.state.CloudInfra.Azure.B.SSHKeyName = ""
	p.state.CloudInfra.Azure.B.SSHUser = ""

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "deleted the ssh key pair", "name", sshName)
	return nil
}
