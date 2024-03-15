package k3s

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

// ConfigureDataStore implements resources.DistroFactory.
// TODO: Update the k3s state struct and also script to use cat <<EOF for passing the certs as well
func (k3s *K3sDistro) ConfigureDataStore(idx int, storage resources.StorageFactory) error {
	log.Print("configuring Datastore", "number", strconv.Itoa(idx))

	//if idx > 0 {
	//	log.Warn("cluster of datastore not enabled!", "number", strconv.Itoa(idx))
	//	return nil
	//}

	password, err := helpers.GenRandomString(15)
	if err != nil {
		return log.NewError("Error in generating random string", "reason", err.Error())
	}

	err = k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(
		scriptDB(password)).
		IPv4(mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.DataStores[idx]).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}
	mainStateDocument.K8sBootstrap.K3s.DataStoreEndPoint = fmt.Sprintf("mysql://ksctl:%s@tcp(%s:3306)/ksctldb", password, mainStateDocument.K8sBootstrap.K3s.B.PrivateIPs.DataStores[idx])
	log.Debug("Printing", "datastoreEndpoint", mainStateDocument.K8sBootstrap.K3s.DataStoreEndPoint)

	err = storage.Write(mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("configured DataStore", "number", strconv.Itoa(idx))

	return nil
}

func scriptDB(password string) string {
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

etcd --version
etcdctl version
etcdutl version

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
  --name infra0 \\
  --initial-advertise-peer-urls https://192.168.1.2:2380 \
  --listen-peer-urls https://192.168.1.2:2380 \\
  --listen-client-urls https://192.168.1.2:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://192.168.1.2:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380 \\
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

`)
}
