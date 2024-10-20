![CoverPage Social Media](./img/cover.svg)

# Ksctl: Simplified Kubernetes Clusters Lifecycle Management


It aims to simplify a collection of kubernetes clusters running on different cloud providers. It provides a simple and intuitive interface for managing Kubernetes clusters. It is designed to be efficient and can perform tasks quickly and without the need for additional tools. It is a powerful tool that can be used to perform a wide range of tasks. 

It is already a valuable tool for developers who want to manage Kubernetes clusters using our CLI. And Get started with Kubernetes without thinking about the infrastructure & configurations. Just run `ksctl create` and your cluster is ready to be used be it a local cluster or a cloud provider managed cluster. It makes the developers skip the cluster setup step as well as day 0 work. Soon we will work on day 1 operations and so on 🙂

So It helps you to avoid using Aws, Azure cli and just create and manage the cluster using a single CLI interface


### So far what have we achieved?
* Cluster Operations
  * Create
  * Delete
  * Get Cluster infra details
  * Storage of state in not just local system but also mongodb
  * Manually Scaleup and Scaledown using the CLI interface
  * Switch Between Clusters
  * Wasm, application stack deployment
* Type Of Cluster
  * Self-Managed HA Cluster
    * K3s
    * Kubeadm
  * Cloud Managed Cluster
    * AKS
    * EKS
    * Civo K3s
* Lifecycle
  * You can now deploy wasm workloads using our Ksctl application stack
  * Common Application Stack Deployment. Example are Argocd, Argorollouts,Kube-Prometheus,etc
  * Initial Prototype of Production Ready Stack (**InProgress**)
  * Etcd Diaster Management (**TBD**)
  * import any cluster (**TBD**)
  * Improved Monitoring of clusters (**In Roadmap**) to make the cluster even more easy to use for someone new to K8s ecosystem


[![Go Report Card](https://goreportcard.com/badge/github.com/ksctl/ksctl)](https://goreportcard.com/report/github.com/ksctl/ksctl) [![](https://pkg.go.dev/badge/github.com/ksctl/ksctl.svg)](https://pkg.go.dev/github.com/ksctl/ksctl) [![OpenSSF Best Practices](https://www.bestpractices.dev/projects/7469/badge)](https://www.bestpractices.dev/projects/7469) [![codecov](https://codecov.io/gh/ksctl/ksctl/branch/main/graph/badge.svg?token=QM61IOCPKC)](https://codecov.io/gh/ksctl/ksctl)

![](https://img.shields.io/github/license/ksctl/ksctl?style=for-the-badge) ![](https://img.shields.io/github/issues/ksctl/ksctl?style=for-the-badge) ![](https://img.shields.io/github/forks/ksctl/ksctl?style=for-the-badge) 


## All Repositories under (Ksctl Org)
Repo | Description
-|-
[Ksctl](https://github.com/ksctl/ksctl) | It provides the core components aka the internals of ksctl features
[Ksctl CLI](https://github.com/ksctl/cli) | It uses the ksctl repo to make a CLI tool
[Ksctl Docs](https://github.com/ksctl/docs) | It's for documentation purpose and to host the ksctl website

## Getting Started guide

[Getting Started guide](https://docs.ksctl.com/docs/getting-started/)

## Usage

Please refer to the [CLI Reference guide](https://docs.ksctl.com/docs/reference/cli/)

## 🙏 Sponsoring
If you like this project and would like to provide financial help, here's our [sponsoring page](https://github.com/sponsors/dipankardas011). Thanks a lot for considering it !

## Contribution Guidelines
Please refer to our [contribution guide](https://docs.ksctl.com/docs/contribution-guidelines/) if you wish to contribute to the project :smile:

[![GitHub repo Good Issues for newbies](https://img.shields.io/github/issues/ksctl/ksctl/good%20first%20issue?style=flat&logo=github&logoColor=green&label=Good%20First%20issues)](https://github.com/ksctl/ksctl/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) [![GitHub Help Wanted issues](https://img.shields.io/github/issues/ksctl/ksctl/help%20wanted?style=flat&logo=github&logoColor=b545d1&label=%22Help%20Wanted%22%20issues)](https://github.com/ksctl/ksctl/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) [![GitHub Help Wanted PRs](https://img.shields.io/github/issues-pr/ksctl/ksctl/help%20wanted?style=flat&logo=github&logoColor=b545d1&label=%22Help%20Wanted%22%20PRs)](https://github.com/ksctl/ksctl/pulls?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) [![GitHub repo Issues](https://img.shields.io/github/issues/ksctl/ksctl?style=flat&logo=github&logoColor=red&label=Issues)](https://github.com/ksctl/ksctl/issues?q=is%3Aopen)

## Interact with the team
* meet us every week tuesday 5:30-6:00 PM (CET) on [Google Group](https://groups.google.com/g/ksctl)

## Thanks to all the contributors ❤️
[Link to Contributors](https://github.com/ksctl/ksctl/graphs/contributors)

<a href="https://github.com/ksctl/ksctl/graphs/contributors">
	<img src="https://contrib.rocks/image?repo=ksctl/ksctl" />
</a>
