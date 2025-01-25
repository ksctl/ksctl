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
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/addons"

	"github.com/ksctl/ksctl/v2/pkg/ssh"
	testHelper "github.com/ksctl/ksctl/v2/pkg/ssh"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"gotest.tools/v3/assert"
)

func TestK3sDistro_Version(t *testing.T) {
	forTesting := map[string]bool{
		"":      true,
		"v1.30": false,
		"v1.31": true,
	}
	for ver, expected := range forTesting {
		v, err := fakeClient.isValidKubeadmVersion(ver)
		got := err == nil

		if got != expected {
			t.Fatalf("Expected for %s as %v but got %v", ver, expected, got)
		}
		if err == nil {
			if len(ver) == 0 {
				assert.Equal(t, v, "v1.31", "it should be equal")
			} else {
				assert.Equal(t, v, ver, "it should be equal")
			}
		}
	}
}

func TestScriptGeneratebootstrapToken(t *testing.T) {

	t.Run("scriptGeneratingBootstrapToken", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "generate bootstrap token",
					CanRetry:       false,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: `
kubeadm token generate
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptToGenerateBootStrapToken()
			},
		)
	})
}

func TestScriptRenewbootstrapToken(t *testing.T) {

	t.Run("scriptGeneratingBootstrapToken", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "renew bootstrap token",
					CanRetry:       false,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: `
kubeadm token create --ttl 20m --description "ksctl bootstrap token"
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptToRenewBootStrapToken()
			},
		)
	})
}

func TestScriptInstallKubeadmAndOtherTools(t *testing.T) {
	ver := "v1.31"

	testHelper.HelperTestTemplate(
		t,
		[]ssh.Script{
			{
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
			},
			{
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
			},
			{
				Name:           "containerd config",
				CanRetry:       false,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
sudo mkdir -p /etc/containerd
containerd config default > config.toml
sudo mv -v config.toml /etc/containerd/config.toml
`,
			},
			{
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
			},
			{
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
			},
			{
				Name:           "apt mark kubenetes tool as hold",
				CanRetry:       false,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
		sudo apt-mark hold kubelet kubeadm kubectl

		`,
			},
		},
		func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
			return scriptInstallKubeadmAndOtherTools(ver)
		},
	)
}

func TestScriptsControlplane(t *testing.T) {

	t.Run("scriptGetCertificateKey", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:     "fetch bootstrap certificate key",
					CanRetry: false,
					ShellScript: `
sudo kubeadm certs certificate-key
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptGetCertificateKey()
			},
		)
	})

	t.Run("scriptDiscoveryTokenCACertHash", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:     "fetch discovery token ca cert hash",
					CanRetry: false,
					ShellScript: `
sudo openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptDiscoveryTokenCACertHash()
			},
		)
	})

	t.Run("scriptGetKubeconfig", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:     "fetch kubeconfig",
					CanRetry: false,
					ShellScript: `
sudo cat /etc/kubernetes/admin.conf
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptGetKubeconfig()
			},
		)
	})

	t.Run("scriptAddKubeadmControlplane0", func(t *testing.T) {
		ver := "1"
		bootstrapToken := "abcd"
		certificateKey := "key"
		publicIPLb := "1.1.1.1"
		privateIPLb := "5.1.1.1"
		privateIPDs := []string{"8.8.8.8"}
		etcdConf := generateExternalEtcdConfig(privateIPDs)

		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:       "store configuration for Controlplane0",
					CanRetry:   true,
					MaxRetries: 3,
					ShellScript: fmt.Sprintf(`
cat <<EOF > kubeadm-config.yml
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: %s
  ttl: 20m
  description: "ksctl bootstrap token"
  usages:
  - signing
  - authentication

certificateKey: %s
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
    - "%s"
    - "%s"
    - "127.0.0.1"
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns: {}
etcd:
  external:
    endpoints:
%s
    caFile: "/etcd/kubernetes/pki/etcd/ca.pem"
    certFile: "/etcd/kubernetes/pki/etcd/etcd.pem"
    keyFile: "/etcd/kubernetes/pki/etcd/etcd-key.pem"
imageRepository: registry.k8s.io
kubernetesVersion: %s.0
controlPlaneEndpoint: "%s:6443"
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
  podSubnet: 10.244.0.0/16
scheduler: {}
EOF

`, bootstrapToken, certificateKey, publicIPLb, privateIPLb, etcdConf, ver, publicIPLb),
				},
				{
					Name:       "kubeadm init",
					CanRetry:   true,
					MaxRetries: 3,
					ShellScript: `
sudo kubeadm init --config kubeadm-config.yml --upload-certs  &>> ksctl.log
#### Adding the below for the kubeconfig to be set so that otken renew can work
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptAddKubeadmControlplane0(ver, bootstrapToken, certificateKey, publicIPLb, privateIPLb, privateIPDs)
			},
		)
	})

	t.Run("script for joining controlplane", func(t *testing.T) {
		privateIPLb := "1.1.1.1"
		token := "abcd"
		cacertSHA := "x2r23erd23"
		crtKey := "xxyy"
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "Join Controlplane [1..N]",
					CanRetry:       true,
					MaxRetries:     3,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: fmt.Sprintf(`
sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s --control-plane --certificate-key %s  &>> ksctl.log
`, privateIPLb, token, cacertSHA, crtKey),
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptJoinControlplane(privateIPLb, token, cacertSHA, crtKey)
			},
		)
	})

	t.Run("scriptTransferEtcdCerts", func(t *testing.T) {
		ca, etcd, key := "-- CA_CERT --", "-- ETCD_CERT --", "-- ETCD_KEY --"
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "save etcd certificate",
					CanRetry:       false,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: fmt.Sprintf(`
sudo mkdir -vp /etcd/kubernetes/pki/etcd/

cat <<EOF > ca.pem
%s
EOF

cat <<EOF > etcd.pem
%s
EOF

cat <<EOF > etcd-key.pem
%s
EOF

sudo mv -v ca.pem etcd.pem etcd-key.pem /etcd/kubernetes/pki/etcd
`, ca, etcd, key),
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptTransferEtcdCerts(ssh.NewExecutionPipeline(), ca, etcd, key)
			},
		)
	})
}

func TestSciprWorkerplane(t *testing.T) {
	privateIPLb := "1.1.1.1"
	token := "abcd"
	cacertSHA := "x2r23erd23"
	testHelper.HelperTestTemplate(
		t,
		[]ssh.Script{
			{
				Name:           "Join K3s workerplane",
				CanRetry:       true,
				MaxRetries:     3,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: fmt.Sprintf(`
sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s &>> ksctl.log
`, privateIPLb, token, cacertSHA),
			},
		},
		func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
			return scriptJoinWorkerplane(ssh.NewExecutionPipeline(), privateIPLb, token, cacertSHA)
		},
	)
}

func checkCurrentStateFile(t *testing.T) {
	raw, err := storeHA.Read()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}

	assert.DeepEqual(t, fakeClient.state, raw)
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(consts.OperationCreate), nil, "should be initlize the state")
	fakeClient.K8sVersion("")
	noCP := len(fakeStateFromCloud.IPv4ControlPlanes)
	noWP := len(fakeStateFromCloud.IPv4WorkerPlanes)
	fakeClient.CNI(nil)
	for no := 0; no < noCP; no++ {
		err := fakeClient.ConfigureControlPlane(no)
		if err != nil {
			t.Fatalf("Configure Controlplane unable to operate %v", err)
		}
	}

	checkCurrentStateFile(t)
	assert.Equal(t, *fakeClient.state.Versions.Kubeadm, "v1.31", "should be equal")

	for no := 0; no < noWP; no++ {
		err := fakeClient.JoinWorkerplane(no)
		if err != nil {
			t.Fatalf("Configure Workerplane unable to operate %v", err)
		}
	}

}

func TestCNI(t *testing.T) {
	testCases := []struct {
		Addon addons.ClusterAddons
		Valid bool
	}{
		{
			addons.ClusterAddons{
				{
					Label: "ksctl",
					Name:  "cilium",
					IsCNI: true,
				},
				{
					Label: "kubeadm",
					Name:  "none",
					IsCNI: true,
				},
			}, true,
		},
		{
			addons.ClusterAddons{}, true,
		},
		{
			nil, true,
		},
	}

	for _, v := range testCases {
		got := fakeClient.CNI(v.Addon)
		assert.Equal(t, got, v.Valid, "missmatch in return value")
	}
}
