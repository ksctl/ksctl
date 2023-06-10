---
sidebar_position: 1
---

# Getting Started

Lets begin with installtion of the tools
their are various method

## Single command method
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

:::info Install
<Tabs groupId="platform" queryString>
  <TabItem value="Linux" label="Linux" default>

```bash
bash <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.sh)
```

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

```bash
zsh <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.sh)
```

  </TabItem>
  <TabItem value="Windows" label="Windows">

```ps1
iwr -useb https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.ps1 | iex
```

  </TabItem>
</Tabs>
:::

:::info uninstall
<Tabs groupId="platform" queryString>
  <TabItem value="Linux" label="Linux" default>

```bash
bash <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.sh)
```

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

```bash
zsh <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.sh)
```

  </TabItem>
  <TabItem value="Windows" label="Windows">

```ps1
iwr -useb https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.ps1 | iex
```

  </TabItem>
</Tabs>
:::


## From Source Code


:::info install
<Tabs groupId="platform-src" queryString>
  <TabItem value="Linux" label="Linux" default>

```bash
make install_linux
```

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

```bash
# macOS on M1
make install_macos

# macOS on INTEL
make install_macos_intel
```

  </TabItem>
  <TabItem value="Windows" label="Windows">

```ps
./builder.ps1
```

  </TabItem>
</Tabs>
:::

:::info uninstall
<Tabs groupId="platform-src" queryString>
  <TabItem value="Linux" label="Linux" default>

```bash
make uninstall
```

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

```bash
make uninstall
```

  </TabItem>
  <TabItem value="Windows" label="Windows">

```ps
./uninstall.ps1
```

  </TabItem>
</Tabs>
:::


:::caution VIDEO not playable

# Installation video

<video width="320" height="240" controls>
<source src="ksctl-install.mp4" type="video/mp4" />
Your browser does not support the video tag.
</video>

:::
