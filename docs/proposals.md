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

> **NOTE**
> Suggestions are most Welcome

# Goal

- Local system
	- [  ] Single Node cluster
	- [  ] Multi-node cluster
	- [  ] High-Available K8s cluster
- Cloud provider
	- [  ] multi-node cluster
	- [  ] High-Available K8s cluster
- Integrated system
	- [  ] Store each cluster's info in some encrypted format in permenanent dir
	- [  ] To have a head pointer similar to `Git` used to identify which context is been used
	- [  ] change contexts by moving `KUBECONFIG` for the target file to the location `~/.kube/config`


# CLI interface

```bash
# local K8s cluster

# create-cluster
kubesimpctl create-cluster --provider local demo-cluster

# delete-cluster
kubesimpctl delete-cluster demo-cluster # auto detect provider and delete the resources

# start-cluster
kubesimpctl start-cluster demo-cluster 	# replace the current KUBECONFIG if present with the specific cluster's kubeconfig

# stop-cluster
kubesimpctl stop-cluster demo-cluster		# replace the current KUBECONFIG with the previous one if present or else empty config file



# Cloud provider


## EKS
# create-cluster
kubesimpctl create-cluster --provider aws demo-cluster2 # other paramaters for EKS specific
# For Example: eksctl


## AKS
# create-cluster
kubesimpctl create-cluster --provider azure demo-cluster3 #other paramaters for EKS specific

```

# API proposal

```go

```


Hope you have Great time Contributing :heart:
