package k3s

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	localstate "github.com/kubesimplify/ksctl/internal/storagelogger/local"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloudControlRes "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
	"gotest.tools/assert"
)

var (
	demoClient         *resources.KsctlClient
	fakeClient         *K3sDistro
	dir                = fmt.Sprintf("%s/ksctl-k3s-test", os.TempDir())
	fakeStateFromCloud cloudControlRes.CloudResourceState
)

func TestMain(m *testing.M) {

	demoClient = &resources.KsctlClient{}
	demoClient.Metadata.ClusterName = "fake"
	demoClient.Metadata.Region = "fake"
	demoClient.Metadata.Provider = CLOUD_AZURE
	demoClient.Distro = ReturnK3sStruct()
	fakeClient = ReturnK3sStruct()

	demoClient.Storage = localstate.InitStorage(false)
	_ = os.Setenv(string(KSCTL_CUSTOM_DIR_ENABLED), dir)
	_ = os.Setenv(string(KSCTL_FAKE_FLAG), "true")
	azHA := utils.GetPath(CLUSTER_PATH, CLOUD_AZURE, CLUSTER_TYPE_HA, "fake fake-resgrp fake-reg")

	if err := os.MkdirAll(azHA, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Created tmp directories")
	fakeStateFromCloud = cloudControlRes.CloudResourceState{
		SSHState: cloudControlRes.SSHInfo{
			PathPrivateKey: utils.GetPath(SSH_PATH, CLOUD_AZURE, CLUSTER_TYPE_HA, "fake fake-resgrp fake-reg"),
			UserName:       "fakeuser",
		},
		Metadata: cloudControlRes.Metadata{
			ClusterName: "fake",
			Provider:    CLOUD_AZURE,
			Region:      "fake-reg",
			ClusterType: CLUSTER_TYPE_HA,
			ClusterDir:  "fake fake-resgrp fake-reg",
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
	if err := os.RemoveAll(dir); err != nil {
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
		valid := fmt.Sprintf(`#!/bin/bash
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
		if valid != scriptWithoutCP_1(ver[i], dbEndpoint[i], pubIP[i]) {
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
		valid := fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server --token %s --datastore-endpoint="%s" --node-taint CriticalAddonsOnly=true:NoExecute --tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ver[i], sampleToken, dbEndpoint[i], pubIP[i])
		if rscript := scriptCP_N(ver[i], dbEndpoint[i], pubIP[i], sampleToken); rscript != valid {
			t.Fatalf("scipts for Controlplane N failed, expected \n%s \ngot \n%s", valid, rscript)
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

	raw, err := demoClient.Storage.Path(utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)).Load()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("Reason: %v", err)
	}

	assert.DeepEqual(t, k8sState, data)
}

func getState(t *testing.T) {
	t.Run("check getState()", func(t *testing.T) {
		expected, err := fakeClient.GetStateFile(demoClient.Storage)
		assert.NilError(t, err, "no error should be there for getstate")

		got, _ := json.Marshal(k8sState)
		assert.DeepEqual(t, string(got), expected)
	})

}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.InitState(fakeStateFromCloud, demoClient.Storage, OPERATION_STATE_CREATE), nil, "should be initlize the state")

	getState(t)

	fakeClient.Version("1.27.1")

	checkCurrentStateFile(t)

	_, err := utils.CreateSSHKeyPair(demoClient.Storage, CLOUD_AZURE, k8sState.ClusterDir)
	if err != nil {
		t.Fatalf("Reason: %v", err)
	}
	t.Log(utils.GetPath(SSH_PATH, CLOUD_AZURE, CLUSTER_TYPE_HA, "fake fake-resgrp fake-reg"))

	err = fakeClient.ConfigureLoadbalancer(demoClient.Storage)
	if err != nil {
		t.Fatalf("Configure Loadbalancer unable to operate %v", err)
	}
	demoClient.Metadata.NoDS = len(fakeStateFromCloud.IPv4DataStores)
	demoClient.Metadata.NoCP = len(fakeStateFromCloud.IPv4ControlPlanes)
	demoClient.Metadata.NoDS = len(fakeStateFromCloud.IPv4DataStores)

	for no := 0; no < demoClient.Metadata.NoDS; no++ {
		err := fakeClient.ConfigureDataStore(no, demoClient.Storage)
		if err != nil {
			t.Fatalf("Configure Datastore unable to operate %v", err)
		}
	}
	for no := 0; no < demoClient.Metadata.NoCP; no++ {
		err := fakeClient.ConfigureControlPlane(no, demoClient.Storage)
		if err != nil {
			t.Fatalf("Configure Controlplane unable to operate %v", err)
		}
	}

	for no := 0; no < demoClient.Metadata.NoWP; no++ {
		err := fakeClient.JoinWorkerplane(no, demoClient.Storage)
		if err != nil {
			t.Fatalf("Configure Workerplane unable to operate %v", err)
		}
	}

	t.Run("get kubeconfig test", func(t *testing.T) {
		a, b, err := fakeClient.GetKubeConfig(demoClient.Storage)
		assert.NilError(t, err, "get kubeconfig should be successful")

		path := utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)

		var raw []byte
		raw, err = demoClient.Storage.Path(path).Load()
		assert.NilError(t, err, "kubeconfig missmatch")

		assert.Equal(t, path, a, "path for the kubeconfig is missmatch")
		assert.Equal(t, string(raw), b, "the kubeconfig must match")

	})
}
