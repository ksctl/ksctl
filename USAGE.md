# Installation

Please refer to the [installation instructions](README.md#setup-cli-local) for installing ksctl

# Usage

Once you have installed ksctl, you can start using it to create and manage your clusters. In the below walkthrough, we will see how you can connect ksctl with a Civo account and manage clusters there.

- [Register your cloud credentials](#register-credentials)
- [Creating a cluster](#create-a-cluster)
- [Saving the Kubeconfig file](#saving-the-kubeconfig)
- [Switching between clusters](#switching-between-multiple-clusters)
- [Deleting Cluster](#deleteing-your-cluster)
- [Creating a High Avaliability cluster](#creating-a-ha-cluster)

#### Register Credentials

- From your Civo dashboard, under profile > Security tab, copy your API key
![](https://i.imgur.com/jexwOeu.png)
- Open a terminal and type in `ksctl cred`
 ![](https://i.imgur.com/fIWyqlH.png)
 - Select the cloud provider you wish to register. In this case, we are using Civo, so we will type `3` and enter.
 - Paste your API key when prompted and hit enter.


Now, ksctl is connected to your civo account. We can now move ahead and create a cluster.

#### Create a Cluster

- After adding your cloud cred to ksctl, you can use `ksctl create [cloud-provider-name]` to create a cluster. This will have some options specific to the cloud provider. Below, we'll take a look at how you can create clusters in civo.

> Note: Please ensure that you have sufficient credits available of your cloud account. 

- Apart from using the `ksctl create civo` command, we will also need to insert some flags. Some flags which will be common across every provider are `Cluster name`, `Node size`, `Region`, and `Number of nodes`.

- We will create a medium-sized cluster, named `demo-cluster`, and we will pre-install ArgoCD from the civo marketplace.
`ksctl create-cluster civo --name demo --nodeSize g4s.kube.medium --nodes 3 --region LON1 --apps argo-cd`

- After you create your cluster, you will get a prompt asking if you want to PRINT your kubeconfig.

> Note: It will only print the kubeconfig, not save it to your local `.kube` folder.

![](https://i.imgur.com/vJunAfl.png)

#### Saving the Kubeconfig

Once your cluster is up and running, you will get a prompt, asking if you want to print your kubeconfig file. This will print your kubeconfig file if you want to view its contents. 

To connect `kubectl` with our cluster, we will need to export the kubeconfig file. Regardless of what option you select in the above step, you will get a prompt to run an `export` command. Run this command in your terminal and you can start using your cluster.

The export command will look something like `export KUBECONFIG='/home/siddhant/.ksctl/config/civo/demo LON1/config'`

If we now check the civo dashboard, we should be able to see our `demo-cluster`

![](https://i.imgur.com/tDJma3C.png)

#### Deleteing your cluster

To delete your cluster

 We can use `ksctl delete [provider-name] --name [cluster-name] -r [region]`.
 
For example, if we want to delete our `demo-cluster`  use `ksctl delete civo --name demo-cluster -r LON1`.

 #### Creating a HA Cluster 

 :construction: In Progress :construction: