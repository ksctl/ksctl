package kubeadm

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"
	"testing"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	testHelper "github.com/ksctl/ksctl/test/helpers"

	"github.com/ksctl/ksctl/pkg/logger"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	cloudControlRes "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	"gotest.tools/v3/assert"
)

var (
	storeHA types.StorageFactory

	fakeClient         *Kubeadm
	dir                = fmt.Sprintf("%s ksctl-kubeadm-test", os.TempDir())
	fakeStateFromCloud cloudControlRes.CloudResourceState
	parentCtx          context.Context
	parentLogger       types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
)

func NewClientHelper(x cloudControlRes.CloudResourceState, state *storageTypes.StorageDocument) *Kubeadm {
	kubeadmCtx = parentCtx
	log = parentLogger

	mainStateDocument = state
	mainStateDocument.K8sBootstrap = &storageTypes.KubernetesBootstrapState{}
	var err error
	mainStateDocument.K8sBootstrap.B.CACert, mainStateDocument.K8sBootstrap.B.EtcdCert, mainStateDocument.K8sBootstrap.B.EtcdKey, err = helpers.GenerateCerts(parentCtx, parentLogger, x.PrivateIPv4DataStores)
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
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)
	parentCtx = context.WithValue(parentCtx, consts.KsctlTestFlagKey, "true")

	mainState := &storageTypes.StorageDocument{}
	if err := helpers.CreateSSHKeyPair(parentCtx, parentLogger, mainState); err != nil {
		log.Error(parentCtx, err.Error())
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

	fakeClient = NewClientHelper(fakeStateFromCloud, &storageTypes.StorageDocument{})
	if fakeClient == nil {
		panic("unable to initialize")
	}

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "fake", consts.ClusterTypeHa)
	_ = storeHA.Connect()

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
		"1.29":   true,
		"1.30":   true,
	}
	for ver, expected := range forTesting {
		err := isValidKubeadmVersion(ver)
		got := err == nil

		if got != expected {
			t.Fatalf("Expected for %s as %v but got %v", ver, expected, got)
		}
	}
}

func TestGeneratebootstrapToken(t *testing.T) {

	got, err := generatebootstrapToken()
	assert.Assert(t, err == nil, "there shouldn't be error")
	pattern := regexp.MustCompile(`\A([a-z0-9]{6})\.([a-z0-9]{16})\z`)

	if pattern.MatchString(got) {
		fmt.Println("Pattern matches")
		match := pattern.FindStringSubmatch(got)
		fmt.Println("Full match:", match[0])
		fmt.Println("First group:", match[1])
		fmt.Println("Second group:", match[2])
	} else {
		t.Fatalf("regex didn't match the helper-gen token")
	}
}

func TestScriptInstallKubeadmAndOtherTools(t *testing.T) {
	ver := "1"

	testHelper.HelperTestTemplate(
		t,
		[]types.Script{
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

curl -fsSL https://pkgs.k8s.io/core:/stable:/v%s/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg --yes

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v%s/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

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
		func() types.ScriptCollection { // Adjust the signature to match your needs
			return scriptInstallKubeadmAndOtherTools(ver)
		},
	)
}

func TestScriptsControlplane(t *testing.T) {

	t.Run("scriptGetCertificateKey", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]types.Script{
				{
					Name:     "fetch bootstrap certificate key",
					CanRetry: false,
					ShellScript: `
sudo kubeadm certs certificate-key
`,
				},
			},
			func() types.ScriptCollection { // Adjust the signature to match your needs
				return scriptGetCertificateKey()
			},
		)
	})

	t.Run("scriptDiscoveryTokenCACertHash", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]types.Script{
				{
					Name:     "fetch discovery token ca cert hash",
					CanRetry: false,
					ShellScript: `
sudo openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
`,
				},
			},
			func() types.ScriptCollection { // Adjust the signature to match your needs
				return scriptDiscoveryTokenCACertHash()
			},
		)
	})

	t.Run("scriptGetKubeconfig", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]types.Script{
				{
					Name:     "fetch kubeconfig",
					CanRetry: false,
					ShellScript: `
sudo cat /etc/kubernetes/admin.conf
`,
				},
			},
			func() types.ScriptCollection { // Adjust the signature to match your needs
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
			[]types.Script{
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
  ttl: 24h0m0s
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
`,
				},
			},
			func() types.ScriptCollection { // Adjust the signature to match your needs
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
			[]types.Script{
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
			func() types.ScriptCollection { // Adjust the signature to match your needs
				return scriptJoinControlplane(privateIPLb, token, cacertSHA, crtKey)
			},
		)
	})

	t.Run("scriptTransferEtcdCerts", func(t *testing.T) {
		ca, etcd, key := "-- CA_CERT --", "-- ETCD_CERT --", "-- ETCD_KEY --"
		testHelper.HelperTestTemplate(
			t,
			[]types.Script{
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
			func() types.ScriptCollection { // Adjust the signature to match your needs
				return scriptTransferEtcdCerts(helpers.NewScriptCollection(), ca, etcd, key)
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
		[]types.Script{
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
		func() types.ScriptCollection { // Adjust the signature to match your needs
			return scriptJoinWorkerplane(helpers.NewScriptCollection(), privateIPLb, token, cacertSHA)
		},
	)
}

func checkCurrentStateFile(t *testing.T) {
	raw, err := storeHA.Read()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}

	assert.DeepEqual(t, mainStateDocument, raw)
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(storeHA, consts.OperationCreate), nil, "should be initlize the state")
	fakeClient.K8sVersion("1.27.1")
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
		"":                       true,
		string(consts.CNICilium): true,
	}

	for k, v := range testCases {
		got := fakeClient.CNI(k)
		assert.Equal(t, got, v, "missmatch")
	}
}
