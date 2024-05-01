package k8sdistros

import (
	"context"
	"fmt"
	testHelper "github.com/ksctl/ksctl/test/helpers"
	"os"
	"testing"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
	cloudControlRes "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
	"gotest.tools/v3/assert"
)

func TestScriptsDataStore(t *testing.T) {
	ca, etcd, key := "-- CA_CERT --", "-- ETCD_CERT --", "-- ETCD_KEY --"
	privIPs := []string{"9.9.9.9"}
	clusterMembers := getEtcdMemberIPFieldForDatastore(privIPs)
	currIdx := 0

	testHelper.HelperTestTemplate(
		t,
		[]resources.Script{
			{
				Name:           "fetch etcd binaries and cleanup",
				ScriptExecutor: consts.LinuxBash,
				MaxRetries:     9,
				CanRetry:       true,
				ShellScript: `
ETCD_VER=v3.5.10

GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GOOGLE_URL}

sudo rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
sudo rm -rf /tmp/etcd-download-test
mkdir -p /tmp/etcd-download-test

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
`,
			},
			{
				Name:           "moving the downloaded binaries to specific location",
				ScriptExecutor: consts.LinuxBash,
				CanRetry:       false,
				ShellScript: `
ETCD_VER=v3.5.10
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1
sudo rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

sudo mv -v /tmp/etcd-download-test/etcd /usr/local/bin
sudo mv -v /tmp/etcd-download-test/etcdctl /usr/local/bin
sudo mv -v /tmp/etcd-download-test/etcdutl /usr/local/bin

sudo rm -rf /tmp/etcd-download-test
`,
			},
			{
				Name:           "store the certificate files",
				ScriptExecutor: consts.LinuxBash,
				CanRetry:       false,
				ShellScript: fmt.Sprintf(`
sudo mkdir -p /var/lib/etcd

cat <<EOF > ca.pem
%s
EOF

cat <<EOF > etcd.pem
%s
EOF

cat <<EOF > etcd-key.pem
%s
EOF

sudo mv -v ca.pem etcd.pem etcd-key.pem /var/lib/etcd
`, ca, etcd, key),
			},
			{
				Name:           "configure etcd configuration file and systemd",
				ScriptExecutor: consts.LinuxBash,
				CanRetry:       false,
				ShellScript: fmt.Sprintf(`

cat <<EOF > etcd.service

[Unit]
Description=etcd

[Service]

ExecStart=/usr/local/bin/etcd \\
  --name infra%d \\
  --initial-advertise-peer-urls https://%s:2380 \
  --listen-peer-urls https://%s:2380 \\
  --listen-client-urls https://%s:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://%s:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster %s \\
  --log-outputs=/var/lib/etcd/etcd.log \\
  --initial-cluster-state new \\
  --peer-auto-tls \\
  --snapshot-count '10000' \\
  --wal-dir=/var/lib/etcd/wal \\
  --client-cert-auth \\
  --trusted-ca-file=/var/lib/etcd/ca.pem \\
  --cert-file=/var/lib/etcd/etcd.pem \\
  --key-file=/var/lib/etcd/etcd-key.pem \\
  --data-dir=/var/lib/etcd/data
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

EOF

sudo mv -v etcd.service /etc/systemd/system
`, currIdx, privIPs[currIdx], privIPs[currIdx], privIPs[currIdx], privIPs[currIdx], clusterMembers),
			},
			{
				Name:           "restart the systemd and start etcd service",
				CanRetry:       true,
				MaxRetries:     3,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
sudo systemctl daemon-reload
sudo systemctl enable etcd

sudo systemctl start etcd
`,
			},
		},
		func() resources.ScriptCollection { // Adjust the signature to match your needs
			return scriptDB(ca, etcd, key, privIPs, currIdx)
		},
	)
}

func TestScriptsLoadbalancer(t *testing.T) {
	array := []string{"127.0.0.1:6443", "127.0.0.2:6443", "127.0.0.3:6443"}

	testHelper.HelperTestTemplate(
		t,
		[]resources.Script{
			{
				Name:       "Install haproxy",
				CanRetry:   true,
				MaxRetries: 9,
				ShellScript: `
sudo apt update -y
sudo apt install haproxy -y
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
				Name:           "create haproxy configuration",
				CanRetry:       false,
				ScriptExecutor: consts.LinuxBash,
				ShellScript: `
sudo systemctl restart haproxy
`,
			},
		},
		func() resources.ScriptCollection { // Adjust the signature to match your needs
			return scriptConfigureLoadbalancer(array)
		},
	)

}

func TestGetEtcdMemberIPFieldForDatastore(t *testing.T) {
	ips := []string{"9.9.9.9", "1.1.1.1"}
	res1 := "infra0=https://9.9.9.9:2380,infra1=https://1.1.1.1:2380"
	assert.Equal(t, res1, getEtcdMemberIPFieldForDatastore(ips), "it should be equal")

	assert.Equal(t, "", getEtcdMemberIPFieldForDatastore([]string{}), "it should be equal")
}

var (
	storeHA resources.StorageFactory

	fakeClient         *PreBootstrap
	dir                = fmt.Sprintf("%s ksctl-bootstrap-test", os.TempDir())
	fakeStateFromCloud cloudControlRes.CloudResourceState
)

func TestMain(m *testing.M) {
	log = logger.NewDefaultLogger(-1, os.Stdout)
	log.SetPackageName("bootstrap")
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

	fakeClient = NewClientHelper(resources.Metadata{
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
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-bootstrap-test"); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func NewClientHelper(m resources.Metadata, state *types.StorageDocument) *PreBootstrap {
	helper := NewPreBootStrap(m, state)
	switch o := helper.(type) {
	case *PreBootstrap:
		return o
	default:
		return nil
	}
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(fakeStateFromCloud, storeHA, consts.OperationCreate), nil, "should be initlize the state")
	noDS := len(fakeStateFromCloud.IPv4DataStores)

	err := fakeClient.ConfigureLoadbalancer(storeHA)
	if err != nil {
		t.Fatalf("Configure Datastore unable to operate %v", err)
	}

	for no := 0; no < noDS; no++ {
		err := fakeClient.ConfigureDataStore(no, storeHA)
		if err != nil {
			t.Fatalf("Configure Datastore unable to operate %v", err)
		}
	}
}
