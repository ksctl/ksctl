---
sidebar_position: 2
---

# Civo

CIVO support for HA and managed clusters

:::note
we need credentials to access clusters
:::

:::caution
these are confidential information so shouldn't be shared with anyone
:::

## Getting credentials

### under settings look for the profile
![](/img/civo/civo-settings.png)
![](/img/civo/profile.png)

### copy the credentials
![](/img/civo/security-api.png)


## Current Features

### Cluster features
#### High Avalibility cluster
clusters which are managed by the user not by cloud provider

    using K3s kubernetes distribution which is lightweight

custom components being used
- MySQL database instance
- HAProxy loadbalancer instance for controlplane nodes
- controlplane instances
- workerplane instances

#### Managed Cluster
clusters which are managed by the cloud provider

### Other capabilities

#### Create, Update, Delete, Switch

:::info Update the cluster infrastructure
**Managed cluster**: till now it's not supported

**HA cluster**
- addition and deletion of new workerplane node
- SSH access to each cluster node (DB, LB, Controplane) _Public Access_, secured by private key
- SSH access to each workplane _Private Access_ via local network, secured by private key
:::

:::caution Creation of HA cluster
when the cluster is created you need to run a command which fixes the issue [#105](https://github.com/kubesimplify/ksctl/issues/105)
:::