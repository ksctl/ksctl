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

import "github.com/ksctl/ksctl/pkg/ssh"

func (p *Provider) DelSSHKeyPair() error {
	if len(p.state.CloudInfra.Civo.B.SSHID) == 0 {
		p.l.Print(p.ctx, "skipped ssh keypair already deleted")
		return nil
	}

	_, err := p.client.DeleteSSHKey(p.state.CloudInfra.Civo.B.SSHID)
	if err != nil {
		return err
	}

	p.l.Success(p.ctx, "ssh keypair deleted", "sshID", p.state.CloudInfra.Civo.B.SSHID)

	p.state.CloudInfra.Civo.B.SSHID = ""
	p.state.CloudInfra.Civo.B.SSHUser = ""
	p.state.SSHKeyPair.PrivateKey, p.state.SSHKeyPair.PrivateKey = "", ""

	return p.store.Write(p.state)
}

func (p *Provider) CreateUploadSSHKeyPair() error {
	name := <-p.chResName

	if len(p.state.CloudInfra.Civo.B.SSHID) != 0 {
		p.l.Print(p.ctx, "skipped ssh keypair already uploaded")
		return nil
	}

	err := ssh.CreateSSHKeyPair(p.ctx, p.l, p.state)
	if err != nil {
		return err
	}
	p.l.Debug(p.ctx, "Printing", "keypair", p.state.SSHKeyPair.PublicKey)

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	if err := p.uploadSSH(name, p.state.SSHKeyPair.PublicKey); err != nil {
		return err
	}
	p.l.Success(p.ctx, "ssh keypair created and uploaded", "sshKeyPairName", name)
	return nil
}

func (p *Provider) uploadSSH(resName, pubKey string) error {
	sshResp, err := p.client.NewSSHKey(resName, pubKey)
	if err != nil {
		return err
	}

	p.state.CloudInfra.Civo.B.SSHID = sshResp.ID
	p.state.CloudInfra.Civo.B.SSHUser = "root"

	p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.B.SSHID", p.state.CloudInfra.Civo.B.SSHID, "p.state.CloudInfra.Civo.B.SSHUser", p.state.CloudInfra.Civo.B.SSHUser)

	return p.store.Write(p.state)
}
