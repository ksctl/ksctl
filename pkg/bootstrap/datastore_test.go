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
	"fmt"
	"github.com/ksctl/ksctl/pkg/ssh"
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	testHelper "github.com/ksctl/ksctl/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestLatestVersionEtcd(t *testing.T) {
	verLatest, err := getLatestVersionEtcd()
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	assert.Equal(t, verLatest, "v3.5.15", "it should be equal")
}

func TestScriptsDataStore(t *testing.T) {
	ca, etcd, key := "-- CA_CERT --", "-- ETCD_CERT --", "-- ETCD_KEY --"
	privIPs := []string{"9.9.9.9"}
	clusterMembers := getEtcdMemberIPFieldForDatastore(privIPs)
	currIdx := 0

	testHelper.HelperTestTemplate(
		t,
		[]ssh.Script{
			{
				Name:           "fetch etcd binaries and cleanup",
				ScriptExecutor: consts.LinuxBash,
				MaxRetries:     9,
				CanRetry:       true,
				ShellScript: `
ETCD_VER=v3.5.15

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
ETCD_VER=v3.5.15
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
		func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
			return scriptDB("v3.5.15", ca, etcd, key, privIPs, currIdx)
		},
	)
}

func TestGetEtcdMemberIPFieldForDatastore(t *testing.T) {
	ips := []string{"9.9.9.9", "1.1.1.1"}
	res1 := "infra0=https://9.9.9.9:2380,infra1=https://1.1.1.1:2380"
	assert.Equal(t, res1, getEtcdMemberIPFieldForDatastore(ips), "it should be equal")

	assert.Equal(t, "", getEtcdMemberIPFieldForDatastore([]string{}), "it should be equal")
}
