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

package bootstrap

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/ssh"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func getLatestVersionHAProxy() (string, error) {
	// currently no method to get the latest LTS version of HAProxy
	// Refer: https://haproxy.debian.net/#distribution=Ubuntu&release=jammy&version=3.0
	return "3.0", nil
}

func (p *PreBootstrap) ConfigureLoadbalancer() error {
	p.l.Note(p.ctx, "configuring Loadbalancer")
	p.mu.Lock()
	sshExecutor := ssh.NewSSHExecutor(p.ctx, p.l, p.state) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	controlPlaneIPs := utilities.DeepCopySlice[string](p.state.K8sBootstrap.B.PrivateIPs.ControlPlanes)

	haProxyVer, err := getLatestVersionHAProxy()
	if err != nil {
		return err
	}

	err = sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		scriptConfigureLoadbalancer(haProxyVer, controlPlaneIPs)).
		IPv4(p.state.K8sBootstrap.B.PublicIPs.LoadBalancer).
		FastMode(true).SSHExecute()
	if err != nil {
		return err
	}

	p.state.Versions.HAProxy = utilities.Ptr(haProxyVer)
	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "configured LoadBalancer")
	return nil
}

func scriptConfigureLoadbalancer(haProxyVer string, controlPlaneIPs []string) ssh.ExecutionPipeline {
	collection := ssh.NewExecutionPipeline()
	// HA proxy repo https://haproxy.debian.net/
	collection.Append(ssh.Script{
		Name:       "Install haproxy",
		CanRetry:   true,
		MaxRetries: 9,
		ShellScript: fmt.Sprintf(`
sudo DEBIAN_FRONTEND=noninteractive apt update -y
sudo DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends software-properties-common -y
sudo DEBIAN_FRONTEND=noninteractive add-apt-repository ppa:vbernat/haproxy-%s -y
sudo DEBIAN_FRONTEND=noninteractive apt-get install haproxy=%s.\* -y
`, haProxyVer, haProxyVer),
		ScriptExecutor: consts.LinuxBash,
	})

	collection.Append(ssh.Script{
		Name:           "enable and start systemd service for haproxy",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo systemctl start haproxy
sudo systemctl enable haproxy
`,
	})

	serverScript := ""
	for index, controlPlaneIP := range controlPlaneIPs {
		serverScript += fmt.Sprintf(`  server k3sserver-%d %s:%d check
`, index+1, controlPlaneIP, 6443)
	}

	collection.Append(ssh.Script{
		Name:           "create haproxy configuration",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
cat <<EOF > haproxy.cfg
frontend kubernetes-frontend
  bind *:6443
  mode tcp
  option tcplog
  timeout client 10s
  default_backend kubernetes-backend

backend kubernetes-backend
  timeout connect 10s
  timeout server 10s
  mode tcp
  option tcp-check
  balance roundrobin
%s
EOF

sudo mv haproxy.cfg /etc/haproxy/haproxy.cfg
`, serverScript),
	})

	collection.Append(ssh.Script{
		Name:           "restarting haproxy",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo systemctl restart haproxy
`,
	})

	return collection
}
