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

### Create TLS certs (openssl)
```bash
cd openssl
openssl genrsa -out ca-key.pem 2048
openssl req -new -key ca-key.pem -out ca-csr.pem -subj "/CN=etcd cluster"

openssl x509 -req -in ca-csr.pem -out ca.pem -days 3650 -signkey ca-key.pem -sha256


################################

openssl genrsa -out etcd-key.pem 2048
openssl req -new -key etcd-key.pem -out etcd-csr.pem -subj "/CN=etcd"
```

```bash
echo subjectAltName = DNS:localhost,IP:192.168.1.6,IP:127.0.0.1 > extfile.cnf
openssl x509 -req -in etcd-csr.pem -CA ca.pem -CAkey ca-key.pem -CAcreateserial -days 3650 -out etcd.pem -sha256 -extfile extfile.cnf
```

> `etcd-key.pem` -> Client key
> `etcd.pem` -> Client certificate
> `ca.pem` -> CA certificate

```bash
etcd --name infra0 --initial-advertise-peer-urls https://192.168.1.6:2380 \
  --listen-peer-urls https://192.168.1.6:2380 \
  --listen-client-urls https://192.168.1.6:2379,https://127.0.0.1:2379 \
  --advertise-client-urls https://192.168.1.6:2379 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster-state new \
  --force-new-cluster \
  --peer-auto-tls \
  --wal-dir=wal \
  --data-dir=data \
  --client-cert-auth \
  --trusted-ca-file=ca.pem \
  --cert-file=etcd.pem \
  --key-file=etcd-key.pem
```

```bash
etcdctl --endpoints=https://192.168.1.6:2379 --key=etcd-key.pem --cert=etcd.pem --cacert=ca.pem  member list
```

---


### Create VMs for Datastore
lets create 3 datastore

Install etcd on datastore server
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


mkdir -p /var/lib/etcd
```

> save /etc/systemd/system/etcd.service


Run on the localsystem
```bash
cd openssl
openssl genrsa -out ca-key.pem 2048
openssl req -new -key ca-key.pem -out ca-csr.pem -subj "/CN=etcd cluster"
openssl x509 -req -in ca-csr.pem -out ca.pem -days 3650 -signkey ca-key.pem -sha256
openssl genrsa -out etcd-key.pem 2048
openssl req -new -key etcd-key.pem -out etcd-csr.pem -subj "/CN=etcd"

echo subjectAltName = DNS:localhost,IP:192.168.1.2,IP:192.168.1.3,IP:192.168.1.4,IP:127.0.0.1 > extfile.cnf
openssl x509 -req -in etcd-csr.pem -CA ca.pem -CAkey ca-key.pem -CAcreateserial -days 3650 -out etcd.pem -sha256 -extfile extfile.cnf
```

> Important to note that when using the ssh access we have to use scp

> when going to use golang we can use cat <<EOF thing with string

> copy it to all `controlplane` nodes and also to the `datastore`

> ALso mkdir `/var/lib/etcd` or any other directory where you want to keep the cert files in controlplane nodes

```bash
scp -v ca.pem etcd.pem etcd-key.pem root@<pub-ip>:/usr/lib/etcd/
```

#### etcd-1

```bash

cat <<EOF > /etc/systemd/system/etcd.service

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
  --initial-advertise-peer-urls https://192.168.1.3:2380 \
  --listen-peer-urls https://192.168.1.3:2380 \\
  --listen-client-urls https://192.168.1.3:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://192.168.1.3:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380 \\
  --log-outputs=/var/lib/etcd/etcd.log \\
  --initial-cluster-state new \\
  --peer-auto-tls \\
  --wal-dir=/var/lib/etcd/wal \\
  --client-cert-auth \\
  --trusted-ca-file=/var/lib/etcd/ca.pem \\
  --cert-file=/var/lib/etcd/etcd.pem \\
  --key-file=/var/lib/etcd/etcd-key.pem \\
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

cat <<EOF > /etc/systemd/system/etcd.service

[Unit]
Description=etcd

[Service]
ExecStart=/usr/local/bin/etcd \\
  --name infra2 \\
  --initial-advertise-peer-urls https://192.168.1.4:2380 \
  --listen-peer-urls https://192.168.1.4:2380 \\
  --listen-client-urls https://192.168.1.4:2379,https://127.0.0.1:2379 \\
  --advertise-client-urls https://192.168.1.4:2379 \\
  --initial-cluster-token etcd-cluster-1 \\
  --initial-cluster infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380 \\
  --log-outputs=/var/lib/etcd/etcd.log \\
  --initial-cluster-state new \\
  --peer-auto-tls \\
  --snapshot-count '10000' \\
  --client-cert-auth \\
  --trusted-ca-file=/var/lib/etcd/ca.pem \\
  --cert-file=/var/lib/etcd/etcd.pem \\
  --key-file=/var/lib/etcd/etcd-key.pem \\
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
etcdctl \
  --cacert=/var/lib/etcd/ca.pem \
  --cert=/var/lib/etcd/etcd.pem \
  --key=/var/lib/etcd/etcd-key.pem \
  endpoint health \
  -w=table \
  --cluster

etcdctl \
  --cacert=/var/lib/etcd/ca.pem \
  --cert=/var/lib/etcd/etcd.pem \
  --key=/var/lib/etcd/etcd-key.pem \
  endpoint status \
  -w=table \
  --cluster

etcdctl \
  --cacert=/var/lib/etcd/ca.pem \
  --cert=/var/lib/etcd/etcd.pem \
  --key=/var/lib/etcd/etcd-key.pem \
  member list \
  -w=table

etcdctl \
  --cacert=/var/lib/etcd/ca.pem \
  --cert=/var/lib/etcd/etcd.pem \
  --key=/var/lib/etcd/etcd-key.pem \
  get / --prefix --keys-only

```

### K3s

#### Create VMs for controlplane
lets create 2 controlplane

```bash
curl -sfL https://get.k3s.io | sh -s - server \
  --datastore-endpoint "https://192.168.1.2:2379,https://192.168.1.3:2379,https://192.168.1.4:2379" \
  --datastore-cafile=/var/lib/etcd/ca.pem \
  --datastore-keyfile=/var/lib/etcd/etcd-key.pem \
  --datastore-certfile=/var/lib/etcd/etcd.pem \
  --tls-san "74.220.22.9"
```

cat /var/lib/rancher/k3s/server/token
```bash
curl -sfL https://get.k3s.io | sh -s - server \
  --token "<>" \
  --datastore-endpoint "https://192.168.1.2:2379,https://192.168.1.3:2379,https://192.168.1.4:2379" \
  --datastore-cafile=/var/lib/etcd/ca.pem \
  --datastore-keyfile=/var/lib/etcd/etcd-key.pem \
  --datastore-certfile=/var/lib/etcd/etcd.pem \
  --tls-san "<ip>"
```

#### Create VMs for workerplane
lets create 1 workerplane
```bash
curl -sfL https://get.k3s.io | sh -s - agent \
  --token "<>" \
  --server "https://<ip>:6443"
```

> Now Testing demo workload
```bash
scp -i <> root@<pub-ip>:/etc/rancher/k3s/k3s.yaml config
#edit the config with the pub ip of the loadbalanbcer
```
```bash
# workload
kubectl run nginx --image=nginx
kubectl expose pod nginx --port=80 --type=LoadBalancer --name=nginx-service

watch kubectl get no,po,svc,componentstatuses -A
```


### Kubeadm

label: `TBD`
