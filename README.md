# ![CoverPage Social Media](./img/ksctl-cover.png)

<h4 align="center">
    <a href="https://discord.com/invite/kubesimplify">Discord</a> |
    <a href="https://kubesimplify.github.io/ksctl/">Website</a><br/><br/>
    <a href="https://kubesimplify.github.io/ksctl/docs/intro">Intro</a> |
    <a href="https://kubesimplify.github.io/ksctl/docs/contributions">Contribute</a> |
    <a href="https://kubesimplify.github.io/ksctl/docs/roadmap">Roadmap</a><br/><br/>
</h4>

**CLI which can manage Kubernetes cluster on different environment**

<div align="center">

   <img src="https://img.shields.io/github/repo-size/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/issues/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/issues-pr/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/issues-pr-closed-raw/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/license/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/forks/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/stars/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/contributors/kubesimplify/ksctl?style=for-the-badge" />
   <img src="https://img.shields.io/github/last-commit/kubesimplify/ksctl?style=for-the-badge" />

   <br>

   [![ci-test-go](https://github.com/kubesimplify/ksctl/actions/workflows/go-fmt.yaml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/go-fmt.yaml)
  [![Testing API](https://github.com/kubesimplify/ksctl/actions/workflows/testingAPI.yaml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/testingAPI.yaml)
  [![Testing Build process](https://github.com/kubesimplify/ksctl/actions/workflows/testBuilder.yaml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/testBuilder.yaml)
  [![goreleaser](https://github.com/kubesimplify/ksctl/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/kubesimplify/ksctl/actions/workflows/goreleaser.yml)
  [![codecov](https://codecov.io/gh/kubesimplify/ksctl/branch/main/graph/badge.svg?token=QM61IOCPKC)](https://codecov.io/gh/kubesimplify/ksctl)
  [![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7469/badge)](https://bestpractices.coreinfrastructure.org/projects/7469)
  [![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=kubesimplify_ksctl&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=kubesimplify_ksctl)
  [![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=kubesimplify_ksctl&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=kubesimplify_ksctl)
  [![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=kubesimplify_ksctl&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=kubesimplify_ksctl)
  [![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=kubesimplify_ksctl&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=kubesimplify_ksctl)
  [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kubesimplify_ksctl&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kubesimplify_ksctl)
  [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=kubesimplify_ksctl&metric=bugs)](https://sonarcloud.io/summary/new_code?id=kubesimplify_ksctl)

</div>




# Project Scope

Many cloud providers offer their flavor of Kubernetes. Each provider has its unique CLI tool which is used to create and manage clusters on that particular cloud. When working in a multi-cloud environment, it can get difficult to create and manage so many clusters using CLI from each cloud provider. ksctl is a **single CLI tool** that can interact with a multitude of cloud providers, making it easy for you to **manage multi-cloud clusters, with just a single CLI tool**

# Purpose

The ksctl project by kubesimplify is a Generic Kubernetes Management command-line tool that helps developers and administrators manage Kubernetes clusters running on different environment

It provides a simple and intuitive interface for performing common tasks such as creating, deleting, and managing Kubernetes resources. ksctl is designed to be easy to use, even for developers who are new to Kubernetes.

Here are some of the specific features of ksctl:

- It provides a simple and intuitive interface for managing Kubernetes clusters.
- It is designed to be efficient and can perform tasks quickly and without the need for additional tools.
- It is a powerful tool that can be used to perform a wide range of tasks.
- It is currently under development, but it is already a valuable tool for developers who want to manage Kubernetes clusters.

# Documentations

Link to the [docs website](https://kubesimplify.github.io/ksctl/)

## Getting Started Azure
Connect ksctl cli to you [Azure](https://kubesimplify.github.io/ksctl/docs/providers/azure) account.

## Getting Started Civo
Connect ksctl cli to your [Civo](https://kubesimplify.github.io/ksctl/docs/providers/civo) account. Watch the installation video here


# Prerequisites

- Go (if building from source)
- Docker (if going to use Local provider)
- Go version >1.20 for build process

# Supported Platforms

Platform | Status
--|--
Linux | :heavy_check_mark:
macOS | :heavy_check_mark:
Windows | :heavy_check_mark:

# Getting Started guide

[Getting Started guide](https://kubesimplify.github.io/ksctl/docs/category/getting-started)

# Usage

Please refer to the [CLI Reference guide](https://kubesimplify.github.io/ksctl/docs/cli/CLI%20command%20reference) also [Docs Website](https://kubesimplify.github.io/ksctl) to know how you can use ksctl

# Contribution Guidelines
Please refer to our [contribution guide](CONTRIBUTING.md) if you wish to contribute to the project :smile:

## Thanks to all the contributors ❤️
[Link to Contributors](https://github.com/kubesimplify/ksctl/graphs/contributors)

<a href="https://github.com/kubesimplify/ksctl/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=kubesimplify/ksctl" />
</a>
