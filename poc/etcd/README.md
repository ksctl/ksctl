# Etcd datastore for HA POC

## Purpose
we need to generate etcd for external datastore with tls

> **Note**
for now we are testing with k3s

## References
- [k3s-datastore](https://docs.k3s.io/datastore)
- [k3s-external-db](https://docs.k3s.io/datastore/ha)
- [etcd-self-signed-tls](https://github.com/etcd-io/etcd/tree/main/hack/tls-setup)
- [etcd-auto-tls](https://etcd.io/docs/v3.5/op-guide/clustering/#automatic-certificates)


## Work

> https://github.com/etcd-io/etcd/releases/tag/v3.5.10

### Create VMs for Datastore
lets create 3 datastore

**Installing**
```bash
ETCD_VER=v3.5.10

# choose either URL
GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GOOGLE_URL}

rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
rm -rf /tmp/etcd-download-test && mkdir -p /tmp/etcd-download-test

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1

rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

mv -v /tmp/etcd-download-test/etcd /usr/local/bin
mv -v /tmp/etcd-download-test/etcdctl /usr/local/bin
mv -v /tmp/etcd-download-test/etcdutl /usr/local/bin

rm -rf /tmp/etcd-download-test

etcd --version
etcdctl version
etcdutl version
```

save /etc/systemd/system/etcd.service


// TODO: need to add for cssl thing https://github.com/etcd-io/etcd/tree/main/hack/tls-setup

**etcd-1**
Private: 192.168.1.2
Ip: 74.220.19.167

```bash
cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
Environment="INTERNAL_IP=192.168.1.2"
Environment="CLUSTER_NAME=infra0"
Environment="NODE_1=192.168.1.2"
Environment="NODE_2=192.168.1.3"
Environment="NODE_3=192.168.1.4"
Environment="NODE_NAME_1=infra0"
Environment="NODE_NAME_2=infra1"
Environment="NODE_NAME_3=infra2"

ExecStart=/usr/local/bin/etcd \\
  --name ${CLUSTER_NAME} \\
  --initial-advertise-peer-urls https://${INTERNAL_IP}:2380 \
  --listen-peer-urls https://${INTERNAL_IP}:2380 \\
  --listen-client-urls https://${INTERNAL_IP}:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://${INTERNAL_IP}:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster ${NODE_NAME_1}=https://${NODE_1}:2380,${NODE_NAME_2}=https://${NODE_2}:2380,${NODE_NAME_3}=https://${NODE_3}:2380 \\
  --initial-cluster-state new \\
  --client-cert-auth \\
  --trusted-ca-file=/path/to/ca-client.crt \\
  --cert-file=/path/to/infra0-client.crt \\
  --key-file=/path/to/infra0-client.key \\
  --peer-client-cert-auth \\
  --peer-trusted-ca-file=ca-peer.crt \\
  --peer-cert-file=/path/to/infra0-peer.crt \\
  --peer-key-file=/path/to/infra0-peer.key \\
  --data-dir=/var/lib/etcd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
```

**etcd-2**
Private: 192.168.1.3
Ip: 74.220.19.191

```bash
cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
Environment="INTERNAL_IP=192.168.1.3"
Environment="CLUSTER_NAME=infra1"
Environment="NODE_1=192.168.1.2"
Environment="NODE_2=192.168.1.3"
Environment="NODE_3=192.168.1.4"
Environment="NODE_NAME_1=infra0"
Environment="NODE_NAME_2=infra1"
Environment="NODE_NAME_3=infra2"

ExecStart=/usr/local/bin/etcd \\
  --name ${CLUSTER_NAME} \\
  --initial-advertise-peer-urls https://${INTERNAL_IP}:2380 \
  --listen-peer-urls https://${INTERNAL_IP}:2380 \\
  --listen-client-urls https://${INTERNAL_IP}:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://${INTERNAL_IP}:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster ${NODE_NAME_1}=https://${NODE_1}:2380,${NODE_NAME_2}=https://${NODE_2}:2380,${NODE_NAME_3}=https://${NODE_3}:2380 \\
  --initial-cluster-state new \\
  --client-cert-auth \\
  --trusted-ca-file=/path/to/ca-client.crt \\
  --cert-file=/path/to/infra0-client.crt \\
  --key-file=/path/to/infra0-client.key \\
  --peer-client-cert-auth \\
  --peer-trusted-ca-file=ca-peer.crt \\
  --peer-cert-file=/path/to/infra0-peer.crt \\
  --peer-key-file=/path/to/infra0-peer.key \\
  --data-dir=/var/lib/etcd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
```

**etcd-3**
Private: 192.168.1.4
Ip: 74.220.21.81

```bash
cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
Environment="INTERNAL_IP=192.168.1.4"
Environment="CLUSTER_NAME=infra2"
Environment="NODE_1=192.168.1.2"
Environment="NODE_2=192.168.1.3"
Environment="NODE_3=192.168.1.4"
Environment="NODE_NAME_1=infra0"
Environment="NODE_NAME_2=infra1"
Environment="NODE_NAME_3=infra2"

ExecStart=/usr/local/bin/etcd \\
  --name ${CLUSTER_NAME} \\
  --initial-advertise-peer-urls https://${INTERNAL_IP}:2380 \
  --listen-peer-urls https://${INTERNAL_IP}:2380 \\
  --listen-client-urls https://${INTERNAL_IP}:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://${INTERNAL_IP}:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster ${NODE_NAME_1}=https://${NODE_1}:2380,${NODE_NAME_2}=https://${NODE_2}:2380,${NODE_NAME_3}=https://${NODE_3}:2380 \\
  --initial-cluster-state new \\
  --client-cert-auth \\
  --trusted-ca-file=/path/to/ca-client.crt \\
  --cert-file=/path/to/infra0-client.crt \\
  --key-file=/path/to/infra0-client.key \\
  --peer-client-cert-auth \\
  --peer-trusted-ca-file=ca-peer.crt \\
  --peer-cert-file=/path/to/infra0-peer.crt \\
  --peer-key-file=/path/to/infra0-peer.key \\
  --data-dir=/var/lib/etcd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
```


**For all**

```bash
sudo systemctl start etcd
```

### Create VMs for controlplane
lets create 2 controlplane

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "https://etcd-host-1:2379,https://etcd-host-2:2379,https://etcd-host-3:2379" \
	--tls-san <public ip of loadbalancer or controlplane>
```


```bash

#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
    --token "from /var/lib/rancher/k3s/server/token" \
    --datastore-endpoint="https://etcd-host-1:2379,https://etcd-host-2:2379,https://etcd-host-3:2379" \
    --node-taint CriticalAddonsOnly=true:NoExecute \
    --tls-san <public ip of loadbalancer or controlplane>
EOF
```

### Create VMs for workerplane
lets create 1 workerplane
