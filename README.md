# ![CoverPage Social Media](./img/ksctl-cover.png)


**CLI which can manage Kubernetes cluster on different environment**

[![ci-test-go](https://github.com/kubesimplify/ksctl/actions/workflows/go-fmt.yaml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/go-fmt.yaml) [![Testing API](https://github.com/kubesimplify/ksctl/actions/workflows/testingAPI.yaml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/testingAPI.yaml) [![Testing Build process](https://github.com/kubesimplify/ksctl/actions/workflows/testBuilder.yaml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/testBuilder.yaml) [![goreleaser](https://github.com/kubesimplify/ksctl/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/goreleaser.yml)


# Project Scope

Many cloud providers offer their flavor of Kubernetes. Each provider has its unique CLI tool which is used to create and manage clusters on that particular cloud. When working in a multi-cloud environment, it can get difficult to create and manage so many clusters using CLI from each cloud provider. ksctl is a **single CLI tool** that can interact with a multitude of cloud providers, making it easy for you to **manage multi-cloud clusters, with just a single CLI tool**. Currently, we support Civo and Local clusters.

# Demo Screenshot
<!-- TODO: Add the demo screenshots-->



# Prerequisites

- Go (if building from source)
- Administrative permission
- Docker (if going to use Local provider)
- Go version >1.19 for build process

# Supported Platforms

Platform | Status
--|--
Linux | :heavy_check_mark:
macOS | :heavy_check_mark:
Windows | :heavy_check_mark:

# Single command install

## Linux and MacOS

```bash
bash <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.sh)
```


## Windows
```powershell
iwr -useb https://raw.githubusercontent.com/kubesimplify/ksctl/main/install.ps1 | iex
```

# Uninstall?

## Linux & MacOs

```bash
bash <(curl -s https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.sh)
```

## Windows
```powershell
iwr -useb https://raw.githubusercontent.com/kubesimplify/ksctl/main/uninstall.ps1 | iex
```

# Build from src
## Linux
### Install


```zsh
make install_linux
```

## macOS
### Install

```zsh
# macOS on M1
make install_macos

# macOS on INTEL
make install_macos_intel
```

### Uninstall
```zsh
make uninstall
```

## Windows
### Install

```ps
./builder.ps1
```

### Uninstall

```ps
./uninstall.ps1
```

# Usage

Please refer to the [usage guide](USAGE.md) to know how you can use ksctl


# Contribution Guidelines
Please refer to our [contribution guide](CONTRIBUTING.md) if you wish to contribute to the project :smile:


# Software Requirement Specification Docs

[Google Doc Link](https://docs.google.com/document/d/1qLGcJly0qWK0dnno6tKXUsm3dd_BpyKl7oi7PLqi6J0/edit?usp=sharing)

## Thanks to all the contributors ❤️
<a href = "https://github.com/kubesimplify/ksctl/graphs/contributors">
  <img src = "https://contrib.rocks/image?repo=kubesimplify/ksctl"/>
</a>
