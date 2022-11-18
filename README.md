# ksctl

A Kubernetes Distribution that can create clusters as well as High-Available clusters in local as well as on cloud platforms

<img src="/img/ksctl-dark.png" style="height: auto!important;width: 600px !important;"/>

# Prerequisites

- Docker installed (if using docker container to run the CLI and for Local clusters)

# Supported Platforms

Platform | Status
--|--
Linux | :heavy_check_mark:
macOS | :heavy_check_mark:
Windows | :heavy_check_mark:

# Project Scope

There are many cloud providers that offer their own flavor of Kubernetes. Each provider has their own unique cli tool which is used to create and manage clusters on that particular cloud. When working in a multi-cloud enviornment, it can get difficult to create and manage so many clusters. ksctl is a single cli tool which can interact with a multitude of cloud providers, making it easy for you to mange multi cloud clusters, with just a single cli tool. Currently, we support Azure, Civo and AWS.

You can also use ksctl to create clusters locally using docker.

# Contribution Guidelines
[contribution Guide](CONTRIBUTION-GUIDE.md)

# Software Requirement Specification Docs

[Google Doc Link](https://docs.google.com/document/d/1qLGcJly0qWK0dnno6tKXUsm3dd_BpyKl7oi7PLqi6J0/edit?usp=sharing)

# Demo Screenshot
<!-- Add the demo screenshots-->

# Setup CLI (Local)
## Host Machine (LINUX)
### Install
```zsh
make install_linux
```
## Host Machine (macOS)
### Install
```zsh
make install_macos
```

### Uninstall
```zsh
make uninstall
```

## Inside Container

### Install

```zsh
make docker_builder docker_run
```
### Uninstall

```zsh
make docker_clean
```

