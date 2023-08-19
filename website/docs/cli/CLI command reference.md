# CLI Command Reference

This topic contains information about ksctl cli commands, syntax, options, resource types, and a few examples of command usage.

:::note TODO
:::

## Syntax
Use the following syntax to run the kubeslice-cli tool:

```bash
ksctl [command] [<command-arguments>] [command-options]
```

## Commands
The following table describes the syntax and descriptions for all the kubeslice-cli commands.

Operations|Syntax|Description
-|-|-
create|`ksctl create [options]` |Creates cluster of ha or cloud managed types
delete|`ksctl delete [options]` |Deletes cluster of ha or cloud managed types
switch|`ksctl switch [options]` |Prints KUBECONFIG variable for using different cluster
get|`ksctl get [options]` |Prints out all the clusters created via ksctl
version|`ksctl version` |Prints out ksctl binary version


:::note NOTE
The ksctl cli tool must have access to the clusters you want it to manage. To grant it access, set the KUBECONFIG environment variable to a path to the kubeconfig file containing the necessary keys to access those clusters. To set the KUBECONFIG environment variable, use these commands:

On Linux/macOS: `export KUBECONFIG="[path to kubeconfig file from the output of creation]"`

On Windows: `$env:KUBECONFIG = "[path to kubeconfig file from the output of creation]"`
:::


## create

### Syntax

### Options

### Example

## delete

### Syntax

### Options

### Example

## switch

### Syntax

### Options

### Example

## get

### Syntax

### Options

### Example