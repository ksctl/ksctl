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

package kubeadm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/ssh"

	"github.com/ksctl/ksctl/v2/pkg/consts"
)

func (p *Kubeadm) JoinWorkerplane(noOfWP int) error {
	p.mu.Lock()
	idx := noOfWP
	sshExecutor := ssh.NewSSHExecutor(p.ctx, p.l, p.state) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	p.l.Note(p.ctx, "configuring Workerplane", "number", strconv.Itoa(idx))

	if err := func() error {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.l.Print(p.ctx, "Checking validity Kubeadm Bootstrap Token")

		tN := time.Now().UTC()
		tM := p.state.K8sBootstrap.Kubeadm.BootstrapTokenExpireTimeUtc
		tDiff := tM.Sub(tN)

		p.l.Debug(p.ctx, "printing debug", "tNow", tN, "tExpire", tM, "tDiff", tDiff)

		// time.After means expire time is after the current time
		if tM.After(tN) && tDiff.Minutes() > 10 {
			p.l.Success(p.ctx, "Valid Kubeadm Bootstrap Token")
			return nil
		} else {
			p.l.Note(p.ctx, "Regenerating Kubeadm Bootstrap Token ttl is near")
			timeCreationBootStrapToken := time.Now().UTC()
			if err := sshExecutor.Flag(consts.UtilExecWithOutput).
				Script(scriptToRenewBootStrapToken()).
				IPv4(p.state.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
				SSHExecute(); err != nil {
				return err
			}
			p.state.K8sBootstrap.Kubeadm.BootstrapToken = strings.Trim(sshExecutor.GetOutput()[0], "\n")
			p.state.K8sBootstrap.Kubeadm.BootstrapTokenExpireTimeUtc = timeCreationBootStrapToken
		}

		return nil
	}(); err != nil {
		return err
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	script := scriptJoinWorkerplane(
		scriptInstallKubeadmAndOtherTools(*p.state.Versions.Kubeadm),
		p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer,
		p.state.K8sBootstrap.Kubeadm.BootstrapToken,
		p.state.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash,
	)
	p.l.Print(p.ctx, "Installing Kubeadm and Joining WorkerNode to existing cluster")

	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(script).
		IPv4(p.state.K8sBootstrap.B.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).
		SSHExecute(); err != nil {
		return err
	}

	p.l.Success(p.ctx, "configured WorkerPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptJoinWorkerplane(collection ssh.ExecutionPipeline, privateIPLb, token, cacertSHA string) ssh.ExecutionPipeline {

	collection.Append(ssh.Script{
		Name:           "Join K3s workerplane",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s &>> ksctl.log
`, privateIPLb, token, cacertSHA),
	})

	return collection
}
