---
slug: feature-install-applications
title: Feature Install Applications
authors: [dipankar]
tags: [kubernetes-client-go, ksctl]
---

:::note Status `Experiemental`
Currently Argocd is supported
:::

It provides a way to install common applications like argocd, and many more to come

to use this feature
```bash
ksctl create ha-<cloud-provider> ...  --feature-flag application --apps argocd
```

