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

package bootstrap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ksctl/ksctl/poller"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func getLatestVersionEtcd() (string, error) {
	latestVersion, err := poller.GetSharedPoller().Get("etcd-io", "etcd")
	if err != nil {
		return "", err
	}

	return latestVersion[0], nil
}

func (p *PreBootstrap) ConfigureDataStore(no int, store types.StorageFactory) error {
	p.mu.Lock()
	idx := no
	sshExecutor := helpers.NewSSHExecutor(bootstrapCtx, log, mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	log.Note(bootstrapCtx, "configuring Datastore", "number", strconv.Itoa(idx))

	etcdVer, err := getLatestVersionEtcd()
	if err != nil {
		return err
	}

	err = sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		scriptDB(
			etcdVer,
			mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey,
			mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores,
			idx)).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.DataStores[idx]).
		FastMode(true).SSHExecute()
	if err != nil {
		return err
	}

	mainStateDocument.K8sBootstrap.B.EtcdVersion = etcdVer
	if err := store.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(bootstrapCtx, "configured DataStore", "number", strconv.Itoa(idx))

	return nil
}

func getEtcdMemberIPFieldForDatastore(ips []string) string {
	var tempDS []string
	for idx, ip := range ips {
		newValue := fmt.Sprintf("infra%d=https://%s:2380", idx, ip)
		tempDS = append(tempDS, newValue)
	}

	return strings.Join(tempDS, ",")
}

func scriptDB(etcdLatestVer, ca, etcd, key string, privIPs []string, currIdx int) types.ScriptCollection {
	collection := helpers.NewScriptCollection()

	collection.Append(types.Script{
		Name:           "fetch etcd binaries and cleanup",
		ScriptExecutor: consts.LinuxBash,
		MaxRetries:     9,
		CanRetry:       true,
		ShellScript: fmt.Sprintf(`
ETCD_VER=%s

GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GOOGLE_URL}

sudo rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
sudo rm -rf /tmp/etcd-download-test
mkdir -p /tmp/etcd-download-test

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
`, etcdLatestVer),
	})

	collection.Append(types.Script{
		Name:           "moving the downloaded binaries to specific location",
		ScriptExecutor: consts.LinuxBash,
		CanRetry:       false,
		ShellScript: fmt.Sprintf(`
ETCD_VER=%s
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1
sudo rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

sudo mv -v /tmp/etcd-download-test/etcd /usr/local/bin
sudo mv -v /tmp/etcd-download-test/etcdctl /usr/local/bin
sudo mv -v /tmp/etcd-download-test/etcdutl /usr/local/bin

sudo rm -rf /tmp/etcd-download-test
`, etcdLatestVer),
	})

	collection.Append(types.Script{
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
	})

	clusterMembers := getEtcdMemberIPFieldForDatastore(privIPs)

	collection.Append(types.Script{
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
	})

	collection.Append(types.Script{
		Name:           "restart the systemd and start etcd service",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo systemctl daemon-reload
sudo systemctl enable etcd

sudo systemctl start etcd
`,
	})

	return collection
}
