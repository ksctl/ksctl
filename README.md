![CoverPage Social Media][cover-img-loc]

# Ksctl: Simplified Kubernetes Clusters Lifecycle Management


<h3>Let's Make Kubernetes accessible to Developers</h3>
<h3>Visit <a href="https://docs.ksctl.com" target="_blank">ksctl docs</a> for the full documentation,
examples and guides.</h3>

[![Discord](https://img.shields.io/badge/discord-ksctl-brightgreen.svg)][discord-link] [![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)][license] [![X/Twitter][x-badge]][x-link]


It aims to simplify a collection of kubernetes clusters running on different cloud providers. It provides a simple and intuitive interface for managing Kubernetes clusters. It is designed to be efficient and can perform tasks quickly and without the need for additional tools. It is a powerful tool that can be used to perform a wide range of tasks.

It is already a valuable tool for developers who want to manage Kubernetes clusters using our CLI. And Get started with Kubernetes without thinking about the infrastructure & configurations. Just run `ksctl create` and your cluster is ready to be used be it a local cluster or a cloud provider managed cluster. It makes the developers skip the cluster setup step as well as day 0 work. Soon we will work on day 1 operations and so on 🙂

So It helps you to avoid using Aws, Azure cli and just create and manage the cluster using a single CLI interface


#### ksctl Components

The main components of ksctl include:

- [**ksctl**][ksctl-gh-link]

It is home to cluster provisioner, Kubernetes Bootstrap, cost & optimization management, addon trigger, interface for cli to use

- [**ksctl/cli (cli)**][cli-gh-link]

It contains the end-user CLI interface. It can perform cluster: create, delete, connect, scaleup, scaledown, list, get; addons: enable, disable

- [**Ksctl Cluster Management (kcm)**][kcm-gh-link]

It contains Kubernetes Controller for manageming ksctl specific cluster addons for now ksctl stack is a part of it. In future we are planning for more kubernetes related addons support like unifying EKS, AKS addons as well.

- [**Ksctl Application (ka)**][ka-gh-link]

It has the controller for ksctl application stack functionality


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
* Lifecycle
  * You can now deploy wasm workloads using our Ksctl application stack
  * Common Application Stack Deployment. Example are gitops, monitoring, etc [Refer](https://docs.ksctl.com/docs/ksctl-cluster-mgt/stacks/)


[![Go Report Card](https://goreportcard.com/badge/github.com/ksctl/ksctl)](https://goreportcard.com/report/github.com/ksctl/ksctl) [![](https://pkg.go.dev/badge/github.com/ksctl/ksctl.svg)](https://pkg.go.dev/github.com/ksctl/ksctl) [![OpenSSF Best Practices](https://www.bestpractices.dev/projects/7469/badge)](https://www.bestpractices.dev/projects/7469)

![](https://img.shields.io/github/license/ksctl/ksctl?style=for-the-badge) ![](https://img.shields.io/github/issues/ksctl/ksctl?style=for-the-badge) ![](https://img.shields.io/github/forks/ksctl/ksctl?style=for-the-badge)

## 📐 Architecture

Here is the entire Ksctl system level design

![ksctl-arch][system-level]


## Getting Started guide

[Getting Started guide][docs-gettingstarted]

## 👋 Community

We welcome contributions from the wider community! Read this [guide][contribution-link] to get started, and join our thriving community on [Discord][discord-link].

🌟 [Leave us a star](https://github.com/ksctl/ksctl), it helps the project to get discovered by others and keeps us motivated to build awesome open-source tools! 🌟

## 🙏 Sponsoring
If you like this project and would like to provide financial help, here's our [sponsoring page](https://github.com/sponsors/dipankardas011). Thanks a lot for considering it !


## 👥 Contributing

To learn about how to contribute to k0rdent, see our [contributing documentation][contribution-link].

k0rdent contributors must follow the [ksctl Code of Conduct][code-of-conduct].

To learn about k0rdent governance, see our [community governance document][governance].

<h1 id="license">📃 License</h1>

Apache License 2.0, see [LICENSE][license].


<h1 id="project resources">💼 Project Resources</h1>

- [k0rdent Community Details](https://github.com/k0rdent/community)
- Join the [Ksctl Discord][discord-link] community.
- k0rdent GitHub:  https://github.com/k0rdent
- k0rdent docs: https://k0rdent.github.io/docs/
- Monthly community call on Tuesday 5:30-6:30 PM (CET) so join our [Google Group](https://groups.google.com/g/ksctl)

## Thanks to all the contributors ❤️
[Link to Contributors](https://github.com/ksctl/ksctl/graphs/contributors)

<a href="https://github.com/ksctl/ksctl/graphs/contributors">
	<img src="https://contrib.rocks/image?repo=ksctl/ksctl" />
</a>



[cover-img-loc]:./assets/img/cover.svg
[x-badge]:https://img.shields.io/twitter/follow/ksctl_org?logo=x&style=flat
[x-link]:https://x.com/ksctl_org
[ksctl-gh-link]:https://github.com/ksctl/ksctl
[cli-gh-link]:https://github.com/ksctl/cli
[kcm-gh-link]:https://github.com/ksctl/kcm
[ka-gh-link]:https://github.com/ksctl/ka
[docs-gettingstarted]:https://docs.ksctl.com/docs/getting-started/
[system-level]:./assets/img/ksctl_solution.svg
[contribution-link]:https://docs.ksctl.com/docs/contribution-guidelines/
[discord-link]:https://discord.com/invite/pWjtKxVrMe
[code-of-conduct]:https://github.com/ksctl/ksctl/blob/main/CODE_OF_CONDUCT.md
[governance]:https://github.com/ksctl/ksctl/blob/main/GOVERNANCE.md
[license]:https://github.com/ksctl/ksctl/blob/main/LICENSE
