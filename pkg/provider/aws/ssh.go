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

package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/v2/pkg/ssh"
)

func (p *Provider) CreateUploadSSHKeyPair() error {

	name := <-p.chResName
	p.l.Debug(p.ctx, "Printing", "name", name)

	if len(p.state.CloudInfra.Aws.B.SSHKeyName) != 0 {
		p.l.Success(p.ctx, "skipped ssh key already created", "name", p.state.CloudInfra.Aws.B.SSHKeyName)
		return nil
	}

	err := ssh.CreateSSHKeyPair(p.ctx, p.l, p.state)
	if err != nil {
		return err
	}

	parameter := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(name),
		PublicKeyMaterial: []byte(p.state.SSHKeyPair.PublicKey),
	}

	if err := p.client.ImportKeyPair(p.ctx, parameter); err != nil {
		return err
	}

	p.state.CloudInfra.Aws.B.SSHKeyName = name
	p.state.CloudInfra.Aws.B.SSHUser = "ubuntu"

	if err := p.store.Write(p.state); err != nil {
		return err
	}
	p.l.Success(p.ctx, "created the ssh key pair", "name", p.state.CloudInfra.Aws.B.SSHKeyName)

	return nil
}

func (p *Provider) DelSSHKeyPair() error {

	if len(p.state.CloudInfra.Aws.B.SSHKeyName) == 0 {
		p.l.Success(p.ctx, "skipped already deleted the ssh key", "name", p.state.CloudInfra.Aws.B.SSHKeyName)
	} else {
		err := p.client.DeleteSSHKey(p.ctx, p.state.CloudInfra.Aws.B.SSHKeyName)
		if err != nil {
			return err
		}

		sshName := p.state.CloudInfra.Aws.B.SSHKeyName

		p.state.CloudInfra.Aws.B.SSHKeyName = ""
		p.state.CloudInfra.Aws.B.SSHUser = ""

		if err := p.store.Write(p.state); err != nil {
			return err
		}

		p.l.Success(p.ctx, "deleted the ssh key pair", "name", sshName)
	}

	return nil
}
