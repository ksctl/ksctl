---
sidebar_position: 2
---

# Civo

CIVO support for HA and Managed Clusters

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

## How to add credentials to ksctl

1. Environment Variables

```bash
export CIVO_TOKEN=""
```

2. Using command line

```bash
ksctl cred
```

## Current Features

### Cluster features
#### Highly Available cluster
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

:::success DEMO

### &nbsp Managed Cluster  {#civoManaged}

<video width="360" height="202" controls>
<source src="../../videos/ksctl-civo-managed.mp4" type="video/mp4" />
Your browser does not support the video tag.
</video>

:::

:::success DEMO

### &nbsp Highly Available Cluster  {#civoHA}

<video width="360" height="202" controls>
<source src="../../videos/ksctl-civo-ha.mp4" type="video/mp4" />
Your browser does not support the video tag.
</video>

:::