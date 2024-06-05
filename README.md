# Ksctl: Cloud Agnostic Kubernetes Management tool

![CoverPage Social Media](./img/ksctl-cover.png)

<h4 align="center">
    <a href="https://discord.com/invite/kubesimplify">Discord</a> |
    <a href="https://docs.ksctl.com/">Website</a><br/><br/>
    <a href="https://docs.ksctl.com/docs/intro">Intro</a> |
    <a href="https://docs.ksctl.com/docs/contributions">Contribute</a> |
    <a href="https://docs.ksctl.com/docs/roadmap">Roadmap</a><br/><br/>
</h4>
<br>
<div align="center">
    <a href="https://pkg.go.dev/github.com/ksctl/ksctl"><img src="https://pkg.go.dev/badge/github.com/ksctl/ksctl.svg" alt="Go Reference"></a>
   <img src="https://img.shields.io/github/issues/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/issues-pr/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/issues-pr-closed-raw/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/license/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/forks/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/stars/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/contributors/ksctl/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/last-commit/ksctl/ksctl?style=for-the-badge" />
   <br>

   [![ci-test-go](https://github.com/ksctl/ksctl/actions/workflows/go-fmt.yaml/badge.svg)](https://github.com/ksctl/ksctl/actions/workflows/go-fmt.yaml)
  [![codecov](https://codecov.io/gh/ksctl/ksctl/branch/main/graph/badge.svg?token=QM61IOCPKC)](https://codecov.io/gh/ksctl/ksctl)
  [![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7469/badge)](https://bestpractices.coreinfrastructure.org/projects/7469)
</div>


# Project Scope

Many cloud providers offer their flavor of Kubernetes. Each provider has its unique CLI tool which is used to create and manage clusters on that particular cloud. When working in a multi-cloud environment, it can get difficult to create and manage so many clusters using CLI from each cloud provider. ksctl is a **single CLI tool** that can interact with a multitude of cloud providers, making it easy for you to **manage multi-cloud clusters, with just a single CLI tool**

# Purpose

The ksctl project by kubesimplify is a Cloud Agnostic Kubernetes Management tool that helps developers and administrators manage Kubernetes clusters running on different environment

It provides a simple and intuitive interface for performing common tasks such as creating, deleting, and managing Kubernetes resources. ksctl is designed to be easy to use, even for developers who are new to Kubernetes.

Here are some of the specific features of ksctl:

- It provides a simple and intuitive interface for managing Kubernetes clusters.
- It is designed to be efficient and can perform tasks quickly and without the need for additional tools.
- It is a powerful tool that can be used to perform a wide range of tasks.
- It is currently under development, but it is already a valuable tool for developers who want to manage Kubernetes clusters.

## Repositories
Repo | Description
-|-
[Ksctl](https://github.com/ksctl/ksctl) | It provides the core components aka the internals of ksctl features
[Ksctl CLI](https://github.com/ksctl/cli) | It uses the ksctl repo to make a CLI tool
[Ksctl Docs](https://github.com/ksctl/docs) | It's for documentation purpose and to host the ksctl website

# Prerequisites

- Go (if building from source)
- Docker (if going to use Local provider)

# Supported Platforms

Platform | Status
--|--
Linux | `OK`
macOS | `OK`
Windows | `OK`

# Getting Started guide

[Getting Started guide](https://docs.ksctl.com/docs/getting-started/)

# Usage

Please refer to the [CLI Reference guide](https://docs.ksctl.com/docs/reference/cli/)
# Contribution Guidelines
Please refer to our [contribution guide](https://docs.ksctl.com/docs/contribution-guidelines/) if you wish to contribute to the project :smile:

## Thanks to all the contributors ❤️
[Link to Contributors](https://github.com/ksctl/ksctl/graphs/contributors)

<a href="https://github.com/ksctl/ksctl/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=ksctl/ksctl" />
</a>
