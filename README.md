# ksctl

A Kubernetes Distribution that can create clusters as well as High-Available clusters in local as well as on cloud platforms

<img src="/img/ksctl-dark.png" style="height: auto!important;width: 600px !important;"/>

# Prerequisites

- Docker installed (if using docker container to run the CLI and for Local clusters)

# Supported Platforms

Platform | Status
--|--
Linux | :heavy_check_mark:
macOS | :heavy_check_mark:
Windows | :heavy_check_mark:

# Project Scope

There are many cloud providers that offer their own flavor of Kubernetes. Each provider has their own unique cli tool which is used to create and manage clusters on that particular cloud. When working in a multi-cloud enviornment, it can get difficult to create and manage so many clusters. ksctl is a single cli tool which can interact with a multitude of cloud providers, making it easy for you to mange multi cloud clusters, with just a single cli tool. Currently, we support Civo and Local clusters.

You can also use ksctl to create clusters locally using docker.

# Contribution Guidelines
Please refer to our [contribution guide](CONTRIBUTING.md) if you wish to contribute to the project :smile:

# Software Requirement Specification Docs

[Google Doc Link](https://docs.google.com/document/d/1qLGcJly0qWK0dnno6tKXUsm3dd_BpyKl7oi7PLqi6J0/edit?usp=sharing)

# Demo Screenshot
<!-- Add the demo screenshots-->

# Setup CLI (Local)
## Host Machine (LINUX)
### Install
```zsh
make install_linux
```
## Host Machine (macOS)
### Install
```zsh
make install_macos
```

### Uninstall
```zsh
make uninstall
```

## Inside Container

### Install

```zsh
make docker_builder docker_run
```
### Uninstall

```zsh
make docker_clean
```

# Usage

After you have built ksctl locally, you can use the following steps to set it up. We will use connect to a Civo account for this example, but the steps will remain the same for any cloud provider.

- [Register your cloud credentials](#register-credentials)
- [Creating a cluster](#create-a-cluster)
- [Saving the Kubeconfig file](#saving-the-kubeconfig)
- [Switching between clusters](#switching-between-multiple-clusters)
- [Deleting Cluster](#deleteing-your-cluster)

#### Register Credentials

- From your Civo dashboard, under profile > Security tab, copy your API key
![](https://i.imgur.com/jexwOeu.png)
- Open a terminal and type in `ksctl cred`
 ![](https://i.imgur.com/fIWyqlH.png)
 - Select the cloud provider you wish to register. In this case we are using Civo, so we will type `3` and enter.
 - Paste your API key when prompted and hit enter.


Now ksctl is connected to your civo account. We can now move ahead and create a cluster.

#### Create a Cluster

- After connection to your cloud account, you can use `ksctl create [cloud-provider-name]` to create a cluster. This will have some options specific to the cloud provider. Below, we'll take a look at how you can create clusters in civo.

> Note: Please ensure that you have added a credit card to your cloud provider. Without it, you might not be able to create clusters.

- Apart from using the `ksctl create civo` command, we will also need to insert some flags. Some flags which will be common accross every provider are `Cluster name`, `Node size`, `Region`, and `Number of nodes`.

- We will create a medium sized cluster, named `demo-cluster`, and we will pre install ArgoCD from the civo marketplace.
`ksctl create-cluster civo --name demo --nodeSize g4s.kube.medium --nodes 3 --region LON1 --apps argo-cd`

- After you create your cluster, you will get a prompt asking if you want to PRINT your kubeconfig.

> Note: It will only print the kubeconfig, not save it to your local `.kube` folder.

![](https://i.imgur.com/vJunAfl.png)

#### Saving the Kubeconfig

- As of right now, ksctl cannot save the kubeconfig file to your local system. We do have a workaround for this. In the above step, type `y` when asked if you want to print your kubeconfig.
- Copy the contents of the printed kubeconfig.
- Navigate to your `.kube` directory, and create a new file named `config` if you don't have it already, and paste the contents of the kubeconfig in this file.
- If you did this correctly, you should be able to access your cluster using `kubectl` now.

If we now check the civo dashboard, we should be able to see our `demo-cluster`

![](https://i.imgur.com/tDJma3C.png)

#### Deleteing your cluster

Let's say you are done with Kubernetes, and have finished your work and now you want to delete the cluster. You can do this relatively easily.

 We can use `ksctl delete [provider-name] --name [cluster-name] -r [region]`. For example, if we want to delete our `demo-cluster` we can do it easily by using the command `ksctl delete civo --name demo-cluster -r LON1`.