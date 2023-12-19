package k3s

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kubesimplify/ksctl/internal/storage/types"

	localstate "github.com/kubesimplify/ksctl/internal/storage/local"
	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloudControlRes "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"gotest.tools/assert"
)

var (
	storeHA resources.StorageFactory

	fakeClient         *K3sDistro
	dir                = fmt.Sprintf("%s ksctl-k3s-test", os.TempDir())
	fakeStateFromCloud cloudControlRes.CloudResourceState
)

func TestMain(m *testing.M) {

	fakeClient = ReturnK3sStruct(resources.Metadata{
		ClusterName:  "fake",
		Region:       "fake",
		Provider:     consts.CloudAzure,
		IsHA:         true,
		LogVerbosity: -1,
		LogWritter:   os.Stdout,
		NoCP:         7,
		NoDS:         5,
		NoWP:         10,
		K8sDistro:    consts.K8sK3s,
	}, &types.StorageDocument{})

	storeHA = localstate.InitStorage(-1, os.Stdout)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "fake", consts.ClusterTypeHa)
	_ = storeHA.Connect(context.TODO())

	_ = os.Setenv(string(consts.KsctlCustomDirEnabled), dir)
	_ = os.Setenv(string(consts.KsctlFakeFlag), "true")

	if err := helpers.CreateSSHKeyPair(log, mainStateDocument); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	fakeStateFromCloud = cloudControlRes.CloudResourceState{
		SSHState: cloudControlRes.SSHInfo{
			PrivateKey: mainStateDocument.SSHKeyPair.PrivateKey,
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
		PrivateIPv4DataStores:    []string{"192.168.X.2"},
		PrivateIPv4LoadBalancer:  "192.168.X.1",
	}

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(os.TempDir() + helpers.PathSeparator + "ksctl-k3s-test"); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func TestK3sDistro_Version(t *testing.T) {
	forTesting := map[string]bool{
		"1.27.4":  true,
		"1.26.7":  true,
		"1.25.12": true,
		"1.27.1":  true,
		"1.27.0":  false,
	}
	for ver, expected := range forTesting {
		if ok := isValidK3sVersion(ver); ok != expected {
			t.Fatalf("Expected for %s as %v but got %v", ver, expected, ok)
		}
	}
}

func TestScriptsControlplane(t *testing.T) {

	ver := []string{"1.26.1", "1.27"}
	dbEndpoint := []string{"demo@(cd);dcwef", "mysql@demo@(cd);dcwef"}
	pubIP := []string{"192.16.9.2", "23.34.4.1"}
	for i := 0; i < len(ver); i++ {
		validWithCni := fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, ver[i], dbEndpoint[i], pubIP[i])

		validWithoutCni := fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, ver[i], dbEndpoint[i], pubIP[i])

		if validWithCni != scriptCP_1(ver[i], dbEndpoint[i], pubIP[i]) {
			t.Fatalf("scipts for Controlplane 1 failed")
		}
		if validWithoutCni != scriptCP_1WithoutCNI(ver[i], dbEndpoint[i], pubIP[i]) {
			t.Fatalf("scipts for Controlplane 1 failed")
		}
	}

	k3sToken := `#!/bin/bash
sudo cat /var/lib/rancher/k3s/server/token
`
	if k3sToken != scriptForK3sToken() {
		t.Fatalf("script for the k3s token missmatch")
	}

	sampleToken := "k3ssdcdsXXXYYYZZZ"
	for i := 0; i < len(ver); i++ {
		validWithCni := fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server --token %s --datastore-endpoint="%s" --node-taint CriticalAddonsOnly=true:NoExecute --tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ver[i], sampleToken, dbEndpoint[i], pubIP[i])

		validWithoutCni := fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--token %s \
	--datastore-endpoint="%s" \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ver[i], sampleToken, dbEndpoint[i], pubIP[i])

		if rscript := scriptCP_N(ver[i], dbEndpoint[i], pubIP[i], sampleToken); rscript != validWithCni {
			t.Fatalf("scipts for Controlplane N failed, expected \n%s \ngot \n%s", validWithCni, rscript)
		}
		if rscript := scriptCP_NWithoutCNI(ver[i], dbEndpoint[i], pubIP[i], sampleToken); rscript != validWithoutCni {
			t.Fatalf("scipts for Controlplane N failed, expected \n%s \ngot \n%s", validWithoutCni, rscript)
		}
	}
}

func TestScriptsDataStore(t *testing.T) {
	password := "$$hello**"
	valid := fmt.Sprintf(`#!/bin/bash
sudo apt update
sudo apt install -y mysql-server

sudo systemctl start mysql

sudo systemctl enable mysql

cat <<EOF > mysqld.cnf
[mysqld]
user		= mysql
bind-address		= 0.0.0.0
#mysqlx-bind-address	= 127.0.0.1

key_buffer_size		= 16M
myisam-recover-options  = BACKUP
log_error = /var/log/mysql/error.log
max_binlog_size   = 100M

EOF

sudo mv mysqld.cnf /etc/mysql/mysql.conf.d/mysqld.cnf

sudo systemctl restart mysql

sudo mysql -e "create user 'ksctl' identified by '%s';"
sudo mysql -e "create database ksctldb; grant all on ksctldb.* to 'ksctl';"

`, password)
	if valid != scriptDB(password) {
		t.Fatalf("script for configuring datastore missmatch")
	}
}

func TestScriptsLoadbalancer(t *testing.T) {
	script := `#!/bin/bash
sudo apt update
sudo apt install haproxy -y
sleep 2s
sudo systemctl start haproxy && sudo systemctl enable haproxy

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
sudo systemctl restart haproxy
`
	array := []string{"127.0.0.1:6443", "127.0.0.2:6443", "127.0.0.3:6443"}
	if script != configLBscript(array) {
		t.Fatalf("script for configuring loadbalancer missmatch")
	}
}

func TestSciprWorkerplane(t *testing.T) {
	ver := "1.18"
	token := "K#Sde43rew34"
	private := "192.20.3.3"
	script := fmt.Sprintf(`#!/bin/bash
cat <<EOF > worker-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - agent --token %s --server https://%s:6443
EOF

sudo chmod +x worker-setup.sh
sudo ./worker-setup.sh
`, ver, token, private)

	if rscript := scriptWP(ver, private, token); rscript != script {
		t.Fatalf("script for configuring the workerplane missmatch, expected \n%s \ngot \n%s", script, rscript)
	}
}

func checkCurrentStateFile(t *testing.T) {

	raw, err := storeHA.Read()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}

	assert.DeepEqual(t, mainStateDocument, raw)
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.InitState(fakeStateFromCloud, storeHA, consts.OperationStateCreate), nil, "should be initlize the state")

	fakeClient.Version("1.27.1")

	checkCurrentStateFile(t)

	err := fakeClient.ConfigureLoadbalancer(storeHA)
	if err != nil {
		t.Fatalf("Configure Loadbalancer unable to operate %v", err)
	}
	noDS := len(fakeStateFromCloud.IPv4DataStores)
	noCP := len(fakeStateFromCloud.IPv4ControlPlanes)
	noWP := len(fakeStateFromCloud.IPv4WorkerPlanes)

	for no := 0; no < noDS; no++ {
		err := fakeClient.ConfigureDataStore(no, storeHA)
		if err != nil {
			t.Fatalf("Configure Datastore unable to operate %v", err)
		}
	}

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
		string(consts.CNIFlannel): false,
		string(consts.CNICilium):  true,
	}

	for k, v := range testCases {
		got := fakeClient.CNI(k)
		assert.Equal(t, got, v, "missmatch")
	}
}
