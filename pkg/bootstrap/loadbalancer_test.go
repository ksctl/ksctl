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
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	testHelper "github.com/ksctl/ksctl/test/helpers"
)

func TestScriptsLoadbalancer(t *testing.T) {
	array := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"}

	testHelper.HelperTestTemplate(
		t,
		[]types.Script{
			{
				Name:       "Install haproxy",
				CanRetry:   true,
				MaxRetries: 9,
				ShellScript: `
sudo DEBIAN_FRONTEND=noninteractive apt update -y
sudo DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends software-properties-common -y
sudo DEBIAN_FRONTEND=noninteractive add-apt-repository ppa:vbernat/haproxy-3.0 -y
sudo DEBIAN_FRONTEND=noninteractive apt-get install haproxy=3.0.\* -y
`,
				ScriptExecutor: consts.LinuxBash,
			},
			{
				Name:           "enable and start systemd service for haproxy",
				CanRetry:       true,
				MaxRetries:     3,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
sudo systemctl start haproxy
sudo systemctl enable haproxy
`,
			},
			{
				Name:           "create haproxy configuration",
				CanRetry:       false,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
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
  server k3sserver-1 127.0.0.1:6443 check
  server k3sserver-2 127.0.0.2:6443 check
  server k3sserver-3 127.0.0.3:6443 check

EOF

sudo mv haproxy.cfg /etc/haproxy/haproxy.cfg
`,
			},
			{
				Name:           "restarting haproxy",
				CanRetry:       true,
				MaxRetries:     3,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
sudo systemctl restart haproxy
`,
			},
		},
		func() types.ScriptCollection { // Adjust the signature to match your needs
			return scriptConfigureLoadbalancer("3.0", array)
		},
	)

}
