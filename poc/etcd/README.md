# Etcd datastore for HA POC

## Purpose
we need to generate etcd for external auto-tls

> **Note**
for now we are testing with k3s

## References
- [k3s-datastore](https://docs.k3s.io/datastore)
- [k3s-external-db](https://docs.k3s.io/datastore/ha)
- [etcd-self-signed-tls](https://github.com/etcd-io/etcd/tree/main/hack/tls-setup)
- [etcd-auto-tls](https://etcd.io/docs/v3.5/op-guide/clustering/#automatic-certificates)


## Work

> We are going to use Azure cloud

### Step 1: need to generate the ssh-keypair

```bash
ssh-keygen
```

### Step 2: Create VMs for Datastore
lets create 3 datastore

### Step 3: Create VMs for controlplane
lets create 2 controlplane

### Step 4: Create VMs for workerplane
lets create 1 workerplane
