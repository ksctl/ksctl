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

#### etcd-1

```bash
cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]

ExecStart=/usr/local/bin/etcd \\
  --name infra0 \\
  --initial-advertise-peer-urls http://192.168.1.6:2380 \
  --listen-peer-urls http://192.168.1.6:2380 \\
  --listen-client-urls http://192.168.1.6:2379,http://127.0.0.1:2379 \\
  --advertise-client-urls http://192.168.1.6:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=http://192.168.1.6:2380,infra1=http://192.168.1.7:2380,infra2=http://192.168.1.8:2380 \\
  --log-outputs=/var/lib/etcd.log \\
  --initial-cluster-state new \\
  --data-dir=/var/lib/etcd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
```

#### etcd-2

```bash
cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
ExecStart=/usr/local/bin/etcd \\
  --name infra1 \\
  --initial-advertise-peer-urls http://192.168.1.7:2380 \
  --listen-peer-urls http://192.168.1.7:2380 \\
  --listen-client-urls http://192.168.1.7:2379,http://127.0.0.1:2379 \\
  --advertise-client-urls http://192.168.1.7:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=http://192.168.1.6:2380,infra1=http://192.168.1.7:2380,infra2=http://192.168.1.8:2380 \\
  --log-outputs=/var/lib/etcd.log \\
  --initial-cluster-state new \\
  --data-dir=/var/lib/etcd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
```

#### etcd-3

```bash
cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
ExecStart=/usr/local/bin/etcd \\
  --name infra2 \\
  --initial-advertise-peer-urls http://192.168.1.8:2380 \
  --listen-peer-urls http://192.168.1.8:2380 \\
  --listen-client-urls http://192.168.1.8:2379,http://127.0.0.1:2379 \\
  --advertise-client-urls http://192.168.1.8:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=http://192.168.1.6:2380,infra1=http://192.168.1.7:2380,infra2=http://192.168.1.8:2380 \\
  --log-outputs=/var/lib/etcd.log \\
  --initial-cluster-state new \\
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
curl -sfL https://get.k3s.io | sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "http://192.168.1.6:2379,http://192.168.1.7:2379,http://192.168.1.8:2379" \
	--tls-san "74.220.20.36"
```


```bash
curl -sfL https://get.k3s.io | sh -s - server \
    --token "K106750583e6a35a52ce92add7a0c9a9177250f8f39c49e8d6b5810f1d352a9adab::server:294adbf8a30918379c243a2567d5f3d0" \
    --datastore-endpoint "http://192.168.1.6:2379,http://192.168.1.7:2379,http://192.168.1.8:2379" \
    --tls-san "74.220.20.36"
```

### Create VMs for workerplane
lets create 1 workerplane


```bash
# workload
k3s kubectl run nginx --image=nginx
kubectl expose pod nginx --port=80 --type=LoadBalancer --name=nginx-service
```
