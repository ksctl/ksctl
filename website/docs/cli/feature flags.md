# Feature Flags

## Auto-Scaling

:::note Status `UnderDevelopment`
Currently planned to support auto-scaling for self-managed clusters
:::

During creation of the cluster it installs necessary configurations and statefile to the cluster
and also creates a slim version of ksctl core api (aka scaleup and scaledown)

Added then the controller will use certain metrics from metrics server to determine when to call HTTP requests to the ksctl to scaleup or scaledown

### Here is deatiled view

![Propsal design](/img/ksctl-auto-scaling-fp.excalidraw.svg)
