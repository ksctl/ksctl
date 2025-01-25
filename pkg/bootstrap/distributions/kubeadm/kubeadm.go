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
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/ssh"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"
	"github.com/ksctl/ksctl/v2/pkg/utilities"

	"github.com/ksctl/ksctl/v2/pkg/poller"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
)

type Kubeadm struct {
	ctx   context.Context
	l     logger.Logger
	state *statefile.StorageDocument
	mu    *sync.Mutex
	store storage.Storage

	// Cni string
}

func (p *Kubeadm) Setup(operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		p.state.K8sBootstrap.Kubeadm = &statefile.StateConfigurationKubeadm{}
		p.state.BootstrapProvider = consts.K8sKubeadm
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}
	return nil
}

func (p *Kubeadm) K8sVersion(ver string) distributions.KubernetesDistribution {
	if v, err := p.isValidKubeadmVersion(ver); err == nil {
		p.state.Versions.Kubeadm = utilities.Ptr(v)
		p.l.Debug(p.ctx, "Printing", "kubeadm.KubeadmVersion", v)
		return p
	} else {
		p.l.Error(err.Error())
		return nil
	}
}

func (p *Kubeadm) CNI(cni addons.ClusterAddons) (externalCNI bool) {
	p.l.Debug(p.ctx, "Printing", "cni", cni)

	_ = cni.GetAddons("kubeadm")

	externalCNI = true

	return
}

func (p *Kubeadm) isValidKubeadmVersion(ver string) (string, error) {

	validVersion, err := poller.GetSharedPoller().Get("kubernetes", "kubernetes")
	if err != nil {
		return "", err
	}

	for i, v := range validVersion {
		_v := strings.Split(v, ".")
		// v1.30.1 -> [v1, 30, 1]
		if len(_v) == 3 {
			validVersion[i] = fmt.Sprintf("%s.%s", _v[0], _v[1])
		}
	}

	if ver == "" {
		return validVersion[0], nil
	}

	for _, vver := range validVersion {
		if vver == ver {
			return vver, nil
		}
	}
	return "", ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidVersion,
		p.l.NewError(p.ctx, "invalid kubeadm version", "valid versions", strings.Join(validVersion, " ")),
	)
}

func NewClient(
	parentCtx context.Context,
	parentLog logger.Logger,
	storage storage.Storage,
	state *statefile.StorageDocument,
) *Kubeadm {
	p := &Kubeadm{mu: &sync.Mutex{}}
	p.ctx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.K8sKubeadm))
	p.l = parentLog
	p.state = state
	p.store = storage

	return p
}

func scriptInstallKubeadmAndOtherTools(ver string) ssh.ExecutionPipeline {
	collection := ssh.NewExecutionPipeline()

	collection.Append(ssh.Script{
		Name:           "disable swap and some kernel module adjustments",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
sudo swapoff -a

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sudo sysctl --system

sudo lsmod | grep br_netfilter
sudo lsmod | grep overlay
sudo sysctl net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward
`,
	})

	collection.Append(ssh.Script{
		Name:           "install containerd",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo apt-get update -y
sudo apt-get install ca-certificates curl gnupg -y

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg --yes
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update -y
sudo apt-get install containerd.io -y
`,
	})

	collection.Append(ssh.Script{
		Name:           "containerd config",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo mkdir -p /etc/containerd
containerd config default > config.toml
sudo mv -v config.toml /etc/containerd/config.toml
`,
	})

	collection.Append(ssh.Script{
		Name:           "restart containerd systemd",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo systemctl restart containerd
sudo systemctl enable containerd

sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
`,
	})

	collection.Append(ssh.Script{
		Name:           "install kubeadm, kubectl, kubelet",
		CanRetry:       true,
		MaxRetries:     9,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo apt-get update -y

sudo apt-get install -y apt-transport-https ca-certificates curl gpg

curl -fsSL https://pkgs.k8s.io/core:/stable:/%s/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg --yes

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/%s/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update -y
sudo apt-get install -y kubelet kubeadm kubectl
sudo systemctl enable kubelet
`, ver, ver),
	})

	collection.Append(ssh.Script{
		Name:           "apt mark kubenetes tool as hold",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
		sudo apt-mark hold kubelet kubeadm kubectl

		`,
	})

	return collection
}
