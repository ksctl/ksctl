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


> **Note**
There is configuration for the data-sir and WAL directory in etcd

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
mkdir -p /var/lib/etcd

cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]

ExecStart=/usr/local/bin/etcd \\
  --name infra0 \\
  --initial-advertise-peer-urls https://192.168.1.2:2380 \
  --listen-peer-urls https://192.168.1.2:2380 \\
  --listen-client-urls http://192.168.1.2:2379,http://127.0.0.1:2379 \\
  --advertise-client-urls http://192.168.1.2:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380 \\
  --log-outputs=/var/lib/etcd/etcd.log \\
  --initial-cluster-state new \\
  --peer-auto-tls \\
  --snapshot-count '10000' \\
  --wal-dir=/var/lib/etcd/wal \\
  --data-dir=/var/lib/etcd/data
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
mkdir -p /var/lib/etcd

cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
ExecStart=/usr/local/bin/etcd \\
  --name infra1 \\
  --initial-advertise-peer-urls https://192.168.1.3:2380 \
  --listen-peer-urls https://192.168.1.3:2380 \\
  --listen-client-urls http://192.168.1.3:2379,http://127.0.0.1:2379 \\
  --advertise-client-urls http://192.168.1.3:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380 \\
  --log-outputs=/var/lib/etcd/etcd.log \\
  --initial-cluster-state new \\
  --peer-auto-tls \\
  --wal-dir=/var/lib/etcd/wal \\
  --snapshot-count '10000' \\
  --data-dir=/var/lib/etcd/data
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
mkdir -p /var/lib/etcd

cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
ExecStart=/usr/local/bin/etcd \\
  --name infra2 \\
  --initial-advertise-peer-urls https://192.168.1.4:2380 \
  --listen-peer-urls https://192.168.1.4:2380 \\
  --listen-client-urls http://192.168.1.4:2379,http://127.0.0.1:2379 \\
  --advertise-client-urls http://192.168.1.4:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380 \\
  --log-outputs=/var/lib/etcd/etcd.log \\
  --initial-cluster-state new \\
  --peer-auto-tls \\
  --snapshot-count '10000' \\
  --wal-dir=/var/lib/etcd/wal \\
  --data-dir=/var/lib/etcd/data
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

```bash
etcdctl endpoint health -w=table --cluster
etcdctl endpoint status -w=table --cluster
```

### K3s

#### Create VMs for controlplane
lets create 2 controlplane

```bash
curl -sfL https://get.k3s.io | sh -s - server \
	--datastore-endpoint "http://192.168.1.2:2379,http://192.168.1.3:2379,http://192.168.1.4:2379" \
	--tls-san "<publicip>"
```


```bash
curl -sfL https://get.k3s.io | sh -s - server \
    --token "<token>" \
    --datastore-endpoint "http://192.168.1.2:2379,http://192.168.1.3:2379,http://192.168.1.4:2379" \
    --tls-san "<publicip>"
```

#### Create VMs for workerplane
lets create 1 workerplane


> Now Testing demo workload
```bash
# workload
k3s kubectl run nginx --image=nginx
k3s kubectl expose pod nginx --port=80 --type=LoadBalancer --name=nginx-service
```


### Kubeadm

label: `TBD`
