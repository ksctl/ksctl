# Etcd datastore for HA POC

## Purpose
we need to generate etcd for external datastore with tls

> **Note**
> - client connection is self-signed tls
> - peer connection is auto-tls

## References
- [k3s-datastore](https://docs.k3s.io/datastore)
- [k3s-external-db](https://docs.k3s.io/datastore/ha)
- [etcd-self-signed-tls](https://github.com/etcd-io/etcd/tree/main/hack/tls-setup)
- [etcd-auto-tls](https://etcd.io/docs/v3.5/op-guide/clustering/#automatic-certificates)
- [automate-tls-certs-in-go](https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251)

> **Note**
There is configuration for the data-sir and WAL directory in etcd

## Work

> https://github.com/etcd-io/etcd/releases/tag/v3.5.10

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


Run on the localsystem (MANUAL STEP TO GENERATE TLS CERTS)
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
> `etcd-key.pem` -> Client key
> `etcd.pem` -> Client certificate
> `ca.pem` -> CA certificate

> Important to note that when using the ssh access we have to use scp

> copy it to all `controlplane` nodes and also to the `datastore`

> ALso mkdir `/var/lib/etcd` or any other directory where you want to keep the cert files in controlplane nodes

Run on the localsystem (AUTOMATED STEP TO GENERATE TLS CERTS)
```bash
go run . 192.168.1.2 192.168.1.3 192.168.1.4 # provide the private IP of the etcd to make ca only valid for SAN on them
```

```bash
scp -i <pem key> ca.pem etcd.pem etcd-key.pem root@<pub-ip>:/var/lib/etcd/
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

> **Note**
> copy the pem files to the contolplane vms before starting k3s
> make sure you mkdir the /var/lib/etcd
  ```bash
  scp -i <pem key> ca.pem etcd.pem etcd-key.pem root@<pub-ip>:/var/lib/etcd/
  ```

```bash
curl -sfL https://get.k3s.io | sh -s - server \
  --datastore-endpoint "https://192.168.1.2:2379,https://192.168.1.3:2379,https://192.168.1.4:2379" \
  --datastore-cafile=/var/lib/etcd/ca.pem \
  --datastore-keyfile=/var/lib/etcd/etcd-key.pem \
  --datastore-certfile=/var/lib/etcd/etcd.pem \
  --tls-san "<ip>"
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

[Refer-kubeadm-config.v1beta3](https://kubernetes.io/docs/reference/config-api/kubeadm-config.v1beta3/)

you can create a cluster-cert using $ kubeadm certs certificate-key You can also specify a custom --certificate-key during init that can later be used by join

> kubeadm init --config <> --upload-certs
```yaml
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcdef.0123456789abcdef
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
localAPIEndpoint:
  advertiseAddress: 192.168.1.3
  bindPort: 6443
certificateKey: b1f5ee0874004360b4eed04c275724a84c360de52bfd22a961b006e577fb9ebd
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
  imagePullPolicy: IfNotPresent
  taints: null
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs:
    - "74.220.22.236"
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns: {}
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: registry.k8s.io
kubernetesVersion: 1.28.0
controlPlaneEndpoint: "74.220.22.236:6443"
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
scheduler: {}
```

```
to get --discovery-token-ca-cert-hash  the copy it
openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
```

```

  kubeadm join 74.220.22.236:6443 --token abcdef.0123456789abcdef \
        --discovery-token-ca-cert-hash sha256:4ec0af85bce7b36812e89d5e8853df4429cd1c89ef03cb02a4d939d872a9d3ed \
        --control-plane --certificate-key b1f5ee0874004360b4eed04c275724a84c360de52bfd22a961b006e577fb9ebd


Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 74.220.22.236:6443 --token abcdef.0123456789abcdef \
        --discovery-token-ca-cert-hash sha256:4ec0af85bce7b36812e89d5e8853df4429cd1c89ef03cb02a4d939d872a9d3ed
```


> Refer to https://gist.github.com/saiyam1814/d87598cf55c71953e288cd22858c0593
```bash
echo "step1- install kubectl,kubeadm and kubelet 1.28.0"

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
echo "kubeadm install"
sudo apt update -y
sudo apt -y install vim git curl wget kubelet=1.28.0-00 kubeadm=1.28.0-00 kubectl=1.28.0-00

echo "memory swapoff"
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
sudo swapoff -a
sudo modprobe overlay
sudo modprobe br_netfilter

echo "Containerd setup"
sudo tee /etc/modules-load.d/containerd.conf <<EOF
overlay
br_netfilter
EOF
sudo tee /etc/sysctl.d/kubernetes.conf<<EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF
sysctl --system
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt update -y
echo -ne '\n' | sudo apt-get -y install containerd
mkdir -p /etc/containerd
containerd config default > /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd
sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable kubelet
echo "image pull and cluster setup"
sudo kubeadm config images pull --cri-socket unix:///run/containerd/containerd.sock --kubernetes-version v1.28.0
```
