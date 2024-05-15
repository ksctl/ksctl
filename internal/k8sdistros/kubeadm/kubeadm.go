package kubeadm

import (
	"context"
	"fmt"
	"strings"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	log               types.LoggerFactory
	kubeadmCtx        context.Context
)

type Kubeadm struct {
	KubeadmVer string
	Cni        string
	mu         *sync.Mutex
}

func (p *Kubeadm) Setup(storage types.StorageFactory, operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap.Kubeadm = &storageTypes.StateConfigurationKubeadm{}
		mainStateDocument.BootstrapProvider = consts.K8sKubeadm
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	return nil
}

func (p *Kubeadm) K8sVersion(ver string) types.KubernetesBootstrap {
	if err := isValidKubeadmVersion(ver); err == nil {
		// valid
		p.KubeadmVer = ver
		log.Debug(kubeadmCtx, "Printing", "kubeadm.KubeadmVersion", p.KubeadmVer)
		return p
	} else {
		log.Error(kubeadmCtx, err.Error())
		return nil
	}
}

func (p *Kubeadm) CNI(cni string) (externalCNI bool) {
	log.Debug(kubeadmCtx, "Printing", "cni", cni)
	switch consts.KsctlValidCNIPlugin(cni) {
	case "":
		p.Cni = ""
	default:
		p.Cni = string(consts.CNINone)
	}
	return true // if its empty string we will install the default cni as flannel
}

func isValidKubeadmVersion(ver string) error {
	validVersion := []string{"1.28", "1.29", "1.30"}

	for _, vver := range validVersion {
		if vver == ver {
			return nil
		}
	}
	return log.NewError(kubeadmCtx, "invalid kubeadm version", "valid versions", strings.Join(validVersion, " "))
}

func NewClient(parentCtx context.Context,
	parentLog types.LoggerFactory,
	state *storageTypes.StorageDocument) *Kubeadm {
	kubeadmCtx = context.WithValue(parentCtx, consts.ContextModuleNameKey, string(consts.K8sKubeadm))
	log = parentLog

	mainStateDocument = state
	return &Kubeadm{mu: &sync.Mutex{}}
}

func scriptInstallKubeadmAndOtherTools(ver string) types.ScriptCollection {
	collection := helpers.NewScriptCollection()

	collection.Append(types.Script{
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

	collection.Append(types.Script{
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

	collection.Append(types.Script{
		Name:           "containerd config",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo mkdir -p /etc/containerd
containerd config default > config.toml
sudo mv -v config.toml /etc/containerd/config.toml
`,
	})

	collection.Append(types.Script{
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

	collection.Append(types.Script{
		Name:           "install kubeadm, kubectl, kubelet",
		CanRetry:       true,
		MaxRetries:     9,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo apt-get update -y

sudo apt-get install -y apt-transport-https ca-certificates curl gpg

curl -fsSL https://pkgs.k8s.io/core:/stable:/v%s/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg --yes

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v%s/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update -y
sudo apt-get install -y kubelet kubeadm kubectl
sudo systemctl enable kubelet
`, ver, ver),
	})

	collection.Append(types.Script{
		Name:           "apt mark kubenetes tool as hold",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
		sudo apt-mark hold kubelet kubeadm kubectl

		`,
	})

	return collection
}
