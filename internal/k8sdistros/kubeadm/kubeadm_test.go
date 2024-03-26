package kubeadm

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/ksctl/ksctl/internal/storage/types"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	cloudControlRes "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
	"gotest.tools/v3/assert"
)

var (
	storeHA resources.StorageFactory

	fakeClient         *Kubeadm
	dir                = fmt.Sprintf("%s ksctl-kubeadm-test", os.TempDir())
	fakeStateFromCloud cloudControlRes.CloudResourceState
)

func NewClientHelper(x cloudControlRes.CloudResourceState, storage resources.StorageFactory, m resources.Metadata, state *types.StorageDocument) *Kubeadm {

	mainStateDocument = state
	mainStateDocument.K8sBootstrap = &types.KubernetesBootstrapState{}
	var err error
	mainStateDocument.K8sBootstrap.B.CACert, mainStateDocument.K8sBootstrap.B.EtcdCert, mainStateDocument.K8sBootstrap.B.EtcdKey, err = helpers.GenerateCerts(log, x.PrivateIPv4DataStores)
	if err != nil {
		return nil
	}

	mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes = x.IPv4ControlPlanes
	mainStateDocument.K8sBootstrap.B.PrivateIPs.ControlPlanes = x.PrivateIPv4ControlPlanes

	mainStateDocument.K8sBootstrap.B.PublicIPs.DataStores = x.IPv4DataStores
	mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores = x.PrivateIPv4DataStores

	mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes = x.IPv4WorkerPlanes

	mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer = x.IPv4LoadBalancer
	mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer = x.PrivateIPv4LoadBalancer
	mainStateDocument.K8sBootstrap.B.SSHInfo = x.SSHState

	return &Kubeadm{mu: &sync.Mutex{}}
}

func TestMain(m *testing.M) {
	log = logger.NewDefaultLogger(-1, os.Stdout)
	log.SetPackageName("kubeadm")
	mainState := &types.StorageDocument{}
	if err := helpers.CreateSSHKeyPair(log, mainState); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	fakeStateFromCloud = cloudControlRes.CloudResourceState{
		SSHState: cloudControlRes.SSHInfo{
			PrivateKey: mainState.SSHKeyPair.PrivateKey,
			UserName:   "fakeuser",
		},
		Metadata: cloudControlRes.Metadata{
			ClusterName: "fake",
			Provider:    consts.CloudAzure,
			Region:      "fake",
			ClusterType: consts.ClusterTypeHa,
		},
		// public IPs
		IPv4ControlPlanes: []string{"A.B.C.4", "A.B.C.5", "A.B.C.6"},
		IPv4DataStores:    []string{"A.B.C.3"},
		IPv4WorkerPlanes:  []string{"A.B.C.2"},
		IPv4LoadBalancer:  "A.B.C.1",

		// Private IPs
		PrivateIPv4ControlPlanes: []string{"192.168.X.7", "192.168.X.9", "192.168.X.10"},
		PrivateIPv4DataStores:    []string{"192.168.5.2"},
		PrivateIPv4LoadBalancer:  "192.168.X.1",
	}

	fakeClient = NewClientHelper(fakeStateFromCloud, storeHA, resources.Metadata{
		ClusterName:  "fake",
		Region:       "fake",
		Provider:     consts.CloudAzure,
		IsHA:         true,
		LogVerbosity: -1,
		LogWritter:   os.Stdout,
		NoCP:         7,
		NoDS:         5,
		NoWP:         10,
		K8sDistro:    consts.K8sKubeadm,
	}, &types.StorageDocument{})
	if fakeClient == nil {
		panic("unable to initialize")
	}

	storeHA = localstate.InitStorage(-1, os.Stdout)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "fake", consts.ClusterTypeHa)
	_ = storeHA.Connect(context.TODO())

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	_ = os.Setenv(string(consts.KsctlFakeFlag), "true")

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-kubeadm-test"); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func TestK3sDistro_Version(t *testing.T) {
	forTesting := map[string]bool{
		"1.26.7": false,
		"1.28":   true,
	}
	for ver, expected := range forTesting {
		if ok := isValidKubeadmVersion(ver); ok != expected {
			t.Fatalf("Expected for %s as %v but got %v", ver, expected, ok)
		}
	}
}

func TestScriptInstallKubeadmAndOtherTools(t *testing.T) {
	expected := `#!/bin/bash
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

sudo mkdir -p /etc/containerd

containerd config default > config.toml

sudo mv -v config.toml /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd

sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml
sudo systemctl restart containerd

sudo apt-get update -y

sudo apt-get install -y apt-transport-https ca-certificates curl gpg

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

sudo systemctl enable kubelet
`
	assert.Equal(t, expected, scriptInstallKubeadmAndOtherTools("1"), "expected to be same")
}

func TestScriptsControlplane(t *testing.T) {

	t.Run("scriptGetCertificateKey", func(t *testing.T) {
		expected := `#!/bin/bash
sudo kubeadm certs certificate-key
`
		assert.Equal(t, expected, scriptGetCertificateKey(), "it should be equal")
	})

	t.Run("scriptDiscoveryTokenCACertHash", func(t *testing.T) {
		expected := `#!/bin/bash
sudo openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
`
		assert.Equal(t, expected, scriptDiscoveryTokenCACertHash(), "it should be equal")
	})

	t.Run("scriptGetKubeconfig", func(t *testing.T) {
		expected := `#!/bin/bash
sudo cat /etc/kubernetes/admin.conf
`
		assert.Equal(t, expected, scriptGetKubeconfig(), "it should be equal")
	})

	t.Run("scriptAddKubeadmControlplane0", func(t *testing.T) {
		expected := `#!/bin/bash
cat <<EOF > kubeadm-config.yml
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcd
  ttl: 24h0m0s
  usages:
  - signing
  - authentication

certificateKey: key
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
  imagePullPolicy: IfNotPresent
  taints: null
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs:
    - "1.1.1.1"
    - "127.0.0.1"
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns: {}
etcd:
  external:
    endpoints:
    - https://8.8.8.8:2379

    caFile: "/etcd/kubernetes/pki/etcd/ca.pem"
    certFile: "/etcd/kubernetes/pki/etcd/etcd.pem"
    keyFile: "/etcd/kubernetes/pki/etcd/etcd-key.pem"
imageRepository: registry.k8s.io
kubernetesVersion: 1.0
controlPlaneEndpoint: "1.1.1.1:6443"
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
scheduler: {}
EOF

sudo kubeadm init --config kubeadm-config.yml --upload-certs
`
		assert.Equal(t, expected, scriptAddKubeadmControlplane0("1", "abcd", "key", "1.1.1.1", []string{"8.8.8.8"}), "it should be equal")
	})

	t.Run("scriptTransferEtcdCerts", func(t *testing.T) {
		expected := `
echo "This script needs to be used as in concatenation to install kubeadm tools"

sudo mkdir -vp /etcd/kubernetes/pki/etcd/

cat <<EOF > ca.pem
ca
EOF

cat <<EOF > etcd.pem
etcd
EOF

cat <<EOF > etcd-key.pem
key
EOF

sudo mv -v ca.pem etcd.pem etcd-key.pem /etcd/kubernetes/pki/etcd
`

		assert.Equal(t, expected, scriptTransferEtcdCerts("ca", "etcd", "key"))
	})
}

func TestSciprWorkerplane(t *testing.T) {
	expected := `#!/bin/bash
sudo kubeadm join 192.2.2.2:6443 --token xyz --discovery-token-ca-cert-hash sha256:qewsdrfcv234 --control-plane --certificate-key abcd
`

	assert.Equal(t, expected, scriptJoinControlplane("192.2.2.2", "xyz", "qewsdrfcv234", "abcd"), "expected to be equal")
}

func checkCurrentStateFile(t *testing.T) {
	raw, err := storeHA.Read()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}

	assert.DeepEqual(t, mainStateDocument, raw)
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(storeHA, consts.OperationStateCreate), nil, "should be initlize the state")
	fakeClient.Version("1.27.1")
	checkCurrentStateFile(t)
	noCP := len(fakeStateFromCloud.IPv4ControlPlanes)
	noWP := len(fakeStateFromCloud.IPv4WorkerPlanes)
	fakeClient.CNI("flannel")
	for no := 0; no < noCP; no++ {
		err := fakeClient.ConfigureControlPlane(no, storeHA)
		if err != nil {
			t.Fatalf("Configure Controlplane unable to operate %v", err)
		}
	}

	for no := 0; no < noWP; no++ {
		err := fakeClient.JoinWorkerplane(no, storeHA)
		if err != nil {
			t.Fatalf("Configure Workerplane unable to operate %v", err)
		}
	}

}
func TestCNI(t *testing.T) {
	testCases := map[string]bool{
		"":                       false,
		string(consts.CNICilium): true,
	}

	for k, v := range testCases {
		got := fakeClient.CNI(k)
		assert.Equal(t, got, v, "missmatch")
	}
}
