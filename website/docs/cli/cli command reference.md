# CLI Command Reference

This topic contains information about ksctl-cli commands, syntax, options, resource types, and a few examples of command usage.

## Syntax

Use the following syntax to run the ksctl-cli tool:

```bash
ksctl [command] [<command-arguments>] [command-options]
```

## Commands

The following table describes the syntax and descriptions for all the ksctl-cli commands.

| Operations     | Syntax                              | Description                                                        |
| ---------------| ----------------------------------- | ------------------------------------------------------------------ |
| cred           | `ksctl cred`                        | Login with your Cloud-provider Credentials                         |
| create         | `ksctl create [options]`            | Creates a cluster of ha or cloud managed types                     |
| delete         | `ksctl delete [options]`            | Delete a cluster of ha or cloud managed types                      |
| get-clusters   | `ksctl get-clusters [options]`      | Prints out all the clusters created via ksctl                      |
| switch-cluster | `ksctl switch-cluster [options]`    | Use to switch between clusters                                     |
| version        | `ksctl version`                     | Prints out ksctl binary version                                    |


## Options

The following are the ksctl-cli options.

| Options         | Shorthand | Description                                                                  |
| --------------- | --------- | ---------------------------------------------------------------------------- |
| --help          | -h        | It provides more information on the ksctl-cli.                               |
| --toggle        | -t        | Help message for toggle                                                      |
| --version       | -v        | It is the version of the `ksctl-cli` tool.                                   |
| --feature-flags | N/A       | It is a feature flag for testing latest development                          |

:::note NOTE

::::caution WARN!
this feature is being worked on
It will be used by the future releases of ksctl
::::

The ksctl cli tool must have access to the clusters you want it to manage. To grant it access, set the KUBECONFIG environment variable to a path to the kubeconfig file containing the necessary keys to access those clusters. To set the KUBECONFIG environment variable, use these commands:

On Linux/macOS: `export KUBECONFIG="[path to kubeconfig file from the output of creation]"`

On Windows: `$env:KUBECONFIG = "[path to kubeconfig file from the output of creation]"`
:::

## Register Credentials

Use this command to login in your selected cloud provider.

### Syntax
```bash
ksctl cred
```

:::note Note
This command is used to configure the credentials for your selected cloud provider. It will prompt you to enter the credentials specific to your cloud provider.
:::

Further select your respected cloud-provider and enter the required credentials when asked to complete the authentication process.
After successful authentication, you should see a confirmation message.

### Example
```bash
$ ksctl cred

1> AWS (EKS)
2> Azure (AKS)
3> Civo (K3s)

2
[LOG] Enter your SUBSCRIPTION ID
    Enter Secret->
[LOG] Enter your TENANT ID
    Enter Secret->
[LOG] Enter your CLIENT ID
    Enter Secret->
[LOG] Enter your CLIENT SECRET
    Enter Secret->
[SUCCESS] [secrets] configuration
[SUCCESS] [ksctl] Credential added
```

## Create a Cluster

Use this command to create cluster. Also have option of creating self-managed or a managed cluster

### Syntax

```bash
ksctl create-cluster <cloud-provider> --name <cluster-name> --node <Number-of-nodes> --region <deafult-region> --nodeSize <Node-size>
```

### Sub-Commands

The following are the `ksctl create [command] [options]` command.

| Command         | Description                                 |
| ----------------| ------------------------------------------- |
| aws             | Use to create a EKS cluster in AWS          |
| azure           | Use to create a AKS cluster in Azure        |
| civo            | Use to create a CIVO k3s cluster            |
| local           | Use to create a LOCAL cluster using Docker  |
| ha-azure        | Use to create a HA cluster in Azure         |
| ha-civo         | Use to create a HA CIVO cluster             |
| ha-[cloud_provider] add-nodes | Use to add more workernodes to existing cluster |

### Options

#### HA (aka self-managed Highly Available cluster)

Flags | Shorthand | Description |
-|- | - |
--approve | | approval to avoid showMsg (default true)
--apps=[string] | | Pre-Installed Applications
--cni=[string] | | CNI plugin to use
--distribution=[string] | |   Kubernetes Distribution
--feature-flags=[string]  | | Supported values with comma seperated features to be enabled
--name=[string] | -n |  Cluster Name (default "demo")
--noCP=[int] | |  Number of ControlPlane Nodes (default 3)
--noDS=[int] | | Number of DataStore Nodes (default 1)
--noWP=[int] | | Number of WorkerPlane Nodes (default 1)
--nodeSizeCP=[string] | | Node size of self-managed controlplane nodes
--nodeSizeDS=[string] | | Node size of self-managed datastore nodes
--nodeSizeLB=[string] | | Node size of self-managed loadbalancer node
--nodeSizeWP=[string] | | Node size of self-managed workerplane nodes
--region=[string] | -r | Region
--verbose | -v | for verbose output (default true)
--version=[string] | | Kubernetes Version

#### HA add-nodes (aka self-managed Highly Available cluster)

Flags | Shorthand | Description |
-|- | - |
--approve | | approval to avoid showMsg (default true)
--distribution=[string] | | Kubernetes Distribution
--feature-flags=[string] | | Supported values with comma seperated features to be enabled
--name=[string] | | Cluster Name (default "demo")
--noWP=[int] | | Number of WorkerPlane Nodes (default -1)
--nodeSizeWP=[string] | | Node size of self-managed workerplane nodes
--region=[string] | | Region
--verbose | | for verbose output (default true)
--version=[string] | | Kubernetes Version

#### Cloud Managed (aka managed)
Flags | Shorthand | Description |
-|- | - |
--approve| | approval to avoid showMsg (default true)
--apps=[string] | | Pre-Installed Applications
--cni=[string]   | | CNI plugin to use
--distribution=[string] | |    Kubernetes Distribution
--feature-flags=[string] | |  Supported values with comma seperated features to be enabled
--name=[string] | -n | Cluster Name
--noMP=[int] | | Number of Managed Nodes (default 1)
--nodeSizeMP=[string] | | Node size of managed cluster nodes
--region=[string] | -r | Region
--verbose | -v |  for verbose output (default true)
--version=[string] | |  Kubernetes Version

:::note IMP
Some Cloud provider's managed cluster offering dont provide options to choose pre-installed apps
but we are continuously working to enable it from our side
:::

### Examples

```bash
# Managed
ksctl create civo --name demo --region LON1
# HA
ksctl create ha-civo --name demo --region LON1
```

## Delete a cluster

Use this command to delete the cluster.
### Sub-Commands

The following are the `ksctl delete [command] [options]` command.

| Command         | Description                                 |
| ----------------| ------------------------------------------- |
| aws             | Use to delete a EKS cluster in AWS          |
| azure           | Use to delete a AKS cluster in Azure        |
| civo            | Use to delete a CIVO k3s cluster            |
| local           | Use to delete a LOCAL cluster using Docker  |
| ha-azure        | Use to delete a HA cluster in Azure         |
| ha-civo         | Use to delete a HA CIVO cluster             |
| ha-[cloud_provider] delete-nodes | Use to delete workernodes from existing cluster |

### Syntax

```bash
ksctl delete <cloud-provider> --name <cluster-name> --region <deafult-region>
```

### Options

The following are the `ksctl delete [cloud-provider] [options]` command options.

Flags | Shorthand | Description |
-|- | - |
--name=[string] | -n | Name of the cluster
--region=[string] | -r | Specify the region of the cluster
--help | -h | It provides information on the delete command.
--verbose | -v | It provides verbose output (default true)
--approve | | approval to avoid showMsg (default true)
--feature-flags=[string] | | Supported values with comma seperated features to be enabled


#### HA delete-nodes (aka self-managed Highly Available cluster)

Flags | Shorthand | Description |
-|- | - |
--approve | | approval to avoid showMsg (default true)
--distribution=[string] | | Kubernetes Distribution
--feature-flags=[string] | | Supported values with comma seperated: [autoscale]
--name=[string] | -n | Cluster Name (default "demo")
--noWP=[int] | | Number of WorkerPlane Nodes (default -1)
--region=[string] | -r | Region
--verbose |-v | for verbose output (default true)

### Examples

```bash
# Managed
ksctl delete civo --name demo --region LON1
# HA
ksctl delete ha-civo --name demo --region LON1
```

## Switch

Use this command to switch between clusters

### Syntax
```bash
ksctl switch --provider <cloud-provider> --name <cluster-name> --region <deafult-region>
```

### Options
Flags | Shorthand | Description |
-|- | - |
--name=[string] | -n | Use to define Name of the Cluster
--region=[string] | -r | Specify the region of the cluster
--provider=[string] | -p | Specify the cloud provider
--verbose | -v | Use to get a verbose output

### Example
```bash
ksctl switch --provider civo --name demo --region LON1
```

## Get

Use to prints out all the clusters created via ksctl in tabluar method

### Example
```bash
kstcl get
```
