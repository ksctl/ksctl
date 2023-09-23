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

| Options   | Shorthand | Description                                                                  |
| --------- | --------- | ---------------------------------------------------------------------------- |
| --help    | -h        | It provides more information on the ksctl-cli.                               |
| --toggle  | -t        | Help message for toggle                                                      |
| --version | -v        | It is the version of the `ksctl-cli` tool.                                   |

:::note NOTE
The ksctl cli tool must have access to the clusters you want it to manage. To grant it access, set the KUBECONFIG environment variable to a path to the kubeconfig file containing the necessary keys to access those clusters. To set the KUBECONFIG environment variable, use these commands:

On Linux/macOS: `export KUBECONFIG="[path to kubeconfig file from the output of creation]"`

On Windows: `$env:KUBECONFIG = "[path to kubeconfig file from the output of creation]"`
:::

## Register Credentials

Use this command to login in your selected cloud provider.

### Syntax
```
ksctl cred
```

:::note Note
This command is used to configure the credentials for your selected cloud provider. It will prompt you to enter the authentication token or credentials specific to your provider.
:::
Further select your respected cloud-provider and copy the token or enter the required credentials when prompted by the command to complete the authentication process.
After successful authentication, you should see a confirmation message.

### Example Output
```bash
[LOG] Enter TOKEN
    Enter Secret-> 
[SUCCESS] [secrets] configuration
[SUCCESS] [ksctl] Credential added
```

Now, ksctl is connected to your cloud provider account. You can proceed to create a cluster.

## Create a Cluster

Use this command to create a cluster. For example, use the create command to create a cluster in a respected cloud provider.

### Syntax

```
ksctl create-cluster <cloud-provider> --name <cluster-name> --node <Number-of-nodes> --region <deafult-region> --nodeSize <Node-size>
```

### Commands

The following are the `ksctl create [command] [options]` command.

| Command         | Description                             |
| ----------------| --------------------------------------- |
| aws             | Use to create a EKS cluster in AWS      |
| azure           | Use to create a AKS cluster in Azure    |
| civo            | Use to create a CIVO k3s cluster        |
| ha-azure        | Use to create a HA k3s cluster in Azure |
| ha-civo         | Use to create a HA CIVO k3s cluster     |
| local           | Use to create a LOCAL cluster in Docker |

### Options

| options             | Shorthand | Description                                                                    |
| --------------------| --------- | ------------------------------------------------------------------------------ |
| --name string       | -n        | Used to name a Cluster                                                         |
| --noMP int          | -N        | Used to defined Number of Nodes                                                |
| --region string     | -r        | Used to specify the region                                                     |
| --nodeSizeMP string | -s        | Used to define the Node size  of managed cluster nodes                         |
| --apps string       | -a        | Used fot PreInstalled Apps with comma seperated string                         |
| --cni string        | -c        | Used to Install CNI Plugin                                                     |
| --help              | -h        | It provides information on the create command.                                 |
| --verbose           | -v        | To get a verbose output                                                        | 
| --apps string       |           | Provides with pre-Installed Applications

### Examples

```bash
ksctl create civo --name demo-cluster --nodeSizeMP g4s.kube.medium --noMP 3 --region LON1 --apps argo-cd
```

### Example Output
```
[civo] Booted Instance demo-cluster-ksctl-managed
[NOTE] KUBECONFIG env var
[NOTE] export KUBECONFIG="/root/.ksctl/config/civo/managed/demo LON1/kubeconfig"

[SUCCESS] [ksctl] created managed cluster
```

## Delete a cluster

Use this command to delete the cluster.

### Syntax

```bash
ksctl delete-cluster civo --name demo-cluster -r LON1
```

### Options

The following are the `ksctl delete [cloud-provider] [options]` command options.

| Name             | Shorthand | Usage                                                                        |
| -----------------| --------- | ---------------------------------------------------------------------------- |
| --name string    | -n        | Name of the cluster                                                          |
| --region string  | -r        | Specify the region of the cluster                                            |
| --help           | -h        | It provides information on the delete command.                               |
| --verbose        | -v        | It provides verbose output (default true)                                    |

### Examples 

```bash
ksctl delete civo --name demo-cluster -r LON1
```

### Example Output
```
SUCCESS [ksctl] deleted managed cluster
```

## Switch 

Use this command to swicth between clusters

### Options
| Name              | Shorthand | Usge                                   |
|------------------ | --------- | -------------------------------------- |
|  --help           |    -h     | It provides help for switch cmd        |
| --name string     |    -n     | Use to define Name of the Cluster      |        
| --region string   |    -r     | Specify the region of the cluster      |
| --provider string |    -p     | Specify the cloud provider             |
| --verbose         |   -v      | Use to get a verbose output            |
     
### Example
```bash
ksctl switch --provider civo --name demo-cluster --region LON1
```

### Example Output
```
[NOTE] export KUBECONFIG="/root/.ksctl/config/civo/managed/demo-cluster LON1/kubeconfig"

[SUCCESS] [ksctl] switched cluster
```

## Get

Use to prints out all the clusters created via ksctl

### Example
```bash
kstcl get-cluster
```

### Examples Output
```
Name            Provider    Nodes  Type     K8s  
demo-cluster    civo(LON1)  wp: 3  managed  k3s 
[SUCCESS] [ksctl] get clusters
```
 