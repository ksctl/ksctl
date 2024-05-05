package kubeadm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

type Kubeadm struct {
	KubeadmVer string
	Cni        string
	mu         *sync.Mutex
}

func (p *Kubeadm) Setup(storage resources.StorageFactory, operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap.Kubeadm = &types.StateConfigurationKubeadm{}
		mainStateDocument.BootstrapProvider = consts.K8sKubeadm
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (p *Kubeadm) Version(ver string) resources.KubernetesBootstrap {
	if isValidKubeadmVersion(ver) {
		// valid
		p.KubeadmVer = ver
		log.Debug("Printing", "kubeadm.KubeadmVersion", p.KubeadmVer)
		return p
	}
	return nil
}

func (p *Kubeadm) CNI(cni string) (externalCNI bool) {
	log.Debug("Printing", "cni", cni)
	switch consts.KsctlValidCNIPlugin(cni) {
	case "":
		p.Cni = ""
	default:
		// this tells us that CNI should be installed via the k8s client
		p.Cni = string(consts.CNINone)
	}
	return true
}

func isValidKubeadmVersion(ver string) bool {
	validVersion := []string{"1.28", "1.29", "1.30"}

	for _, vver := range validVersion {
		if vver == ver {
			return true
		}
	}
	log.Error(strings.Join(validVersion, " "))
	return false
}

var (
	mainStateDocument *types.StorageDocument
	log               resources.LoggerFactory
)

func NewClient(m resources.Metadata, state *types.StorageDocument) resources.KubernetesBootstrap {
	log = logger.NewDefaultLogger(m.LogVerbosity, m.LogWritter)
	log.SetPackageName("kubeadm")

	mainStateDocument = state
	return &Kubeadm{mu: &sync.Mutex{}}
}

func scriptInstallKubeadmAndOtherTools(ver string) resources.ScriptCollection {
	collection := helpers.NewScriptCollection()

	collection.Append(resources.Script{
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

	collection.Append(resources.Script{
		Name:           "install containerd",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install containerd.io -y
`,
	})

	collection.Append(resources.Script{
		Name:           "containerd config",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo mkdir -p /etc/containerd
containerd config default > config.toml
sudo mv -v config.toml /etc/containerd/config.toml
`,
	})

	collection.Append(resources.Script{
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

	collection.Append(resources.Script{
		Name:           "install kubeadm, kubectl, kubelet",
		CanRetry:       true,
		MaxRetries:     9,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo apt-get update -y

sudo apt-get install -y apt-transport-https ca-certificates curl gpg

curl -fsSL https://pkgs.k8s.io/core:/stable:/v%s/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v%s/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo systemctl enable kubelet
`, ver, ver),
	})

	collection.Append(resources.Script{
		Name:           "apt mark kubenetes tool as hold",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
		sudo apt-mark hold kubelet kubeadm kubectl

		`,
	})

	return collection
}
