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
<Tabs groupId="install-platform" queryString>
  <TabItem value="Linux" label="Linux" default>

    $ bash <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.sh)

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

    $ zsh <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.sh)

  </TabItem>
  <TabItem value="Windows" label="Windows">

    $ iwr -useb https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.ps1 | iex

  </TabItem>
</Tabs>
:::

:::info uninstall
<Tabs groupId="uninstall-platform" queryString>
  <TabItem value="Linux" label="Linux" default>

    $ bash <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.sh)

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

    $ zsh <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.sh)

  </TabItem>
  <TabItem value="Windows" label="Windows">

    $ iwr -useb https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.ps1 | iex

  </TabItem>
</Tabs>
:::


## From Source Code


:::info install
<Tabs groupId="install-platform-src" queryString>
  <TabItem value="Linux" label="Linux" default>

    $ make install_linux

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

    # macOS on M1
    $ make install_macos

    # macOS on INTEL
    $ make install_macos_intel

  </TabItem>
  <TabItem value="Windows" label="Windows">

    $ ./builder.ps1

  </TabItem>
</Tabs>
:::

:::info uninstall
<Tabs groupId="uninstall-platform-src" queryString>
  <TabItem value="Linux" label="Linux" default>

    $ make uninstall

  </TabItem>
  <TabItem value="MacOS" label="MacOS">

    $ make uninstall

  </TabItem>
  <TabItem value="Windows" label="Windows">

    $ ./uninstall.ps1

  </TabItem>
</Tabs>
:::

