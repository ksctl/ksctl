---
slug: feature-auto-scaling
title: Feature Auto Scaling
authors: [dipankar]
tags: [kubernetes-client-go, ksctl]
---

:::note Status `UnderDevelopment`
Currently planned to support auto-scaling for self-managed clusters

(Controllers part is under development)
:::

During creation of the cluster it installs necessary configurations and statefile to the cluster
and also creates a slim version of ksctl core api (aka scaleup and scaledown)

Added then the controller will use certain metrics from metrics server to determine when to call HTTP requests to the ksctl to scaleup or scaledown

### Here is deatiled view

![Propsal design](/img/ksctl-auto-scaling-fp.excalidraw.svg)
