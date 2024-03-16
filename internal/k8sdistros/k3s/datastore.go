package k3s

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/resources"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

// ConfigureDataStore implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureDataStore(idx int, storage resources.StorageFactory) error {
	log.Print("configuring Datastore", "number", strconv.Itoa(idx))

	err := k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(
		scriptDB(
			mainStateDocument.K8sBootstrap.K3s.B.CACert,
			mainStateDocument.K8sBootstrap.K3s.B.EtcdCert,
			mainStateDocument.K8sBootstrap.K3s.B.EtcdKey,
			mainStateDocument.K8sBootstrap.K3s.B.PrivateIPs.DataStores,
			idx)).
		IPv4(mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.DataStores[idx]).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = storage.Write(mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("configured DataStore", "number", strconv.Itoa(idx))

	return nil
}

func scriptDB(ca, etcd, key string, privIPs []string, currIdx int) string {

	clusterMembers := getEtcdMemberIPFieldForDatastore(privIPs)

	return fmt.Sprintf(`#!/bin/bash
set -xe

ETCD_VER=v3.5.10

GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GOOGLE_URL}

sudo rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
sudo rm -rf /tmp/etcd-download-test && mkdir -p /tmp/etcd-download-test

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1

sudo rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

sudo mv -v /tmp/etcd-download-test/etcd /usr/local/bin
sudo mv -v /tmp/etcd-download-test/etcdctl /usr/local/bin
sudo mv -v /tmp/etcd-download-test/etcdutl /usr/local/bin

sudo rm -rf /tmp/etcd-download-test

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

sudo systemctl daemon-reload
sudo systemctl enable etcd

sudo systemctl start etcd

`, ca, etcd, key, currIdx, privIPs[currIdx], privIPs[currIdx], privIPs[currIdx], privIPs[currIdx], clusterMembers)
}
