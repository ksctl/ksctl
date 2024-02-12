# about ksctl components

should we create crd for only autoscaler or a generic crd for all the ksctl tasks and make it such that it can accomodate all

the domain choosen is `ksctl.kubesimplify.com`
```bash
kubebuilder init --domain ksctl.kubesimplify.com
```

as of now plan is the autoscaler we can have
    `ksctl-cluster/v1` as the apiVersion
    `KsctlAutoScaler` as the kind

as of now plan is the installapplication we can have

for **Installing**
    `ksctl-apps/v1` as the apiVersion
    `KsctlInstall` as the kind

for **Uninstalling**
    `ksctl-apps/v1` as the apiVersion
    `KsctlRemove` as the kind

as of now plan for the ksctl agent will be having a controller for getting latest changes
from the above types

## AutoScaler controller

Refer to this doc [Link](./autoScalerController/idea.md)