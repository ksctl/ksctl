# How all things should work

* use of OCI images for creating nodes for the cluster (Local cluster)
* use of standard API calls to cloud platform specific Kubernetes clusters(AKS, EKS, etc.)
* CLI should be able to `create-cluster`, `view-cluster`, `delete-cluster`, `switch-context`

> **NOTE**
> Suggestions are most Welcome

# WorkFlow

1. CLI will get the request from the user
2. Some processing program will check request's validation (local validation and remote validation by dry-run)
3. Then CLI will call using API to create job which in-turn will allocate the neccessary resources and configure them

# Development
1. When there is an API Change in `main` branch or in `PR` make sure to update the API part first then in second commit make the neccesssary changes to go.mod in src/cli like changing the branch and then `go mod tidy` to get the latest commit hash for that branch

> **NOTE**
> Suggestions are most Welcome

# Goal

- Local system
	- [x] Single Node cluster
	- [x] Multi-node cluster
- Cloud provider
	- [x] multi-node cluster
	- [ ] High-Available K8s cluster
- Integrated system
	- [ ] Store each cluster's info in some encrypted format in permenanent dir
	- [ ] To have a head pointer similar to `Git` used to identify which context is been used
    - [ ] change contexts by moving `KUBECONFIG` for the target file to the location `~/.kube/config`


# CLI interface

> **NOTE**
> Suggestions are most Welcome

```bash
ksctl get-contexts
ksctl set-context
ksctl init
ksctl pre-check

# local K8s cluster

# create-cluster
ksctl create-cluster --provider local -name demo-cluster

# delete-cluster
ksctl delete-cluster -name demo-cluster # auto detect provider and delete the resources

# start-cluster
ksctl start-cluster -name demo-cluster 	# replace the current KUBECONFIG if present with the specific cluster's kubeconfig

# stop-cluster
ksctl stop-cluster -name demo-cluster		# replace the current KUBECONFIG with the previous one if present or else empty config file



# Cloud provider


## EKS
# create-cluster
ksctl create-cluster --provider aws -name demo-cluster2 # other paramaters for EKS specific
# For Example: eksctl


## AKS
# create-cluster
ksctl create-cluster --provider azure -name demo-cluster3 #other paramaters for EKS specific

```


## Need for SHA encoding for cluster names
- we will be using the SHA of cluster name as a unique reference
- current-context can be stored rather than the kubeconfig file location
  For Example:
```bash
ksctl create cluster -p aws -name demo-aws-1 -n 2
```

and set CONTEXT to this SHA code

CONTEXT file
```editorconfig
[CURRCONTEXT]
1e16b17f262363de8d659731e693d1ebfb99f8cfc2cfb81fd1c0fd3487f49154
[PREVCONTEXT]

```

To store all the cluster names with SHA code as its key
Cluster.conf
```text
1e16b17f... 	demo-aws-1	aws
c07ed107... 	demo-aks-012	azure
bf67c55e... 	local-demo	local
1ea1bf64... 	mini-demo	local
```

[//]: # (TODO: view contexts)

[//]: # (TODO: set contexts)

```bash
# GET SHA256 of a particular string
echo -n foobar | openssl dgst -sha256 | awk '{print $2}'
# OR
echo -n foobar | sha256sum | awk '{print $1}'
```

## File management for `KUBECONFIG` and `CREDENTIALS`
> **NOTE**
> Suggestions are most Welcome


```prototext
~/.ksctl
    cred
      aws
      azure
			civo
    config
      aws
        |- 1ea1bf647945ff30efd1a62d0be84da659c248760cf9c0412840979a7b40a65a.yaml
        |- bf67c55e1add70240ce7df7e7d0634da60c988a0f014da0dd635efc8136d9872.yaml
      azure
        |- c07ed10784ad2ff06e24aad10a90d3bd6bfdc8216cccea3a6aecd9575d04ab5d.yaml
      local
        |- 2a97516c354b68848cdbd8f54a226a0a55b21ed138e207ad6c5cbb9c00aa5aea.yaml
```


# API proposal

```go
// General structure for API handler to consume from CLI

type Machine struct {
	Nodes uint8
	Cpu   uint8
	Mem   uint8
	Disk  uint8
}

type AwsProvider struct {
	ClusterName string
	HACluster   bool
	Region      string
	Spec        Machine
	AccessKey   string
	Secret      string
}

type AzureProvider struct {
	ClusterName         string
	HACluster           bool
	Region              string
	Spec                Machine
	SubscriptionID      string
	TenantID            string
	servicePrincipleKey string
	servicePrincipleID  string
}

type CivoProvider struct {
	ClusterName string
	APIKey      string
	HACluster   bool
	Region      string
	Spec        Machine
}

type LocalProvider struct {
	ClusterName string
	HACluster   bool
	Region      string
	Spec        Machine
}

type Providers struct {
	eks  *AwsProvider
	aks  *AzureProvider
	k3s  *CivoProvider
	mk8s *LocalProvider
}

```


Hope you have Great time Contributing :heart:
