![ksctl - Kubernetes Simplified][cover-img-loc]

# Ksctl: Simplified Kubernetes Clusters Lifecycle Management

<h3>Making Kubernetes accessible to developers</h3>

[![Discord](https://img.shields.io/badge/discord-ksctl-brightgreen.svg)][discord-link]
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)][license]
[![X/Twitter][x-badge]][x-link]
[![Go Report Card](https://goreportcard.com/badge/github.com/ksctl/ksctl)](https://goreportcard.com/report/github.com/ksctl/ksctl)
[![](https://pkg.go.dev/badge/github.com/ksctl/ksctl.svg)](https://pkg.go.dev/github.com/ksctl/ksctl)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/7469/badge)](https://www.bestpractices.dev/projects/7469)

## 📖 Overview

Ksctl simplifies the management of Kubernetes clusters across different cloud providers through a unified interface. **It focuses on creating cost-efficient and sustainable Kubernetes infrastructure by intelligently optimizing resource selection and deployment.** It eliminates the need for multiple cloud-specific CLIs (AWS, Azure, etc.) and handles the complex infrastructure setup for you.

> [!TIP]
> **Visit [ksctl documentation](https://docs.ksctl.com) for complete guides and examples.**

## ✨ Key Features

- **Unified Management Interface** - Manage clusters on multiple cloud providers with a single CLI.
- **Rapid Cluster Creation** - Deploy production-ready clusters in 5-10 minutes with `ksctl create`.
- **Multi-Cloud & Cluster Types** - Supports self-managed (K3s, Kubeadm) and managed clusters (AKS, EKS, GKE) across major clouds.
- **Sustainability Focus** - Intelligently selects regions and instances with lower carbon footprints and higher renewable energy usage.
- **Cost Optimization** - Reduces cloud expenses through smart resource selection and dynamic region switching.
- **Simplified Lifecycle Management** - Create, delete, scale, and switch between clusters easily.
- **Application Stack Deployment** - Deploy common applications like GitOps tools and monitoring solutions.

## 🏗️ Architecture

![ksctl system architecture][system-level]

## 🧩 Components

The ksctl ecosystem consists of these main components:

- [**ksctl**][ksctl-gh-link] - Core component for cluster provisioning, Kubernetes bootstrap, cost optimization, and addon management
- [**ksctl/cli**][cli-gh-link] - End-user CLI interface for cluster and addon operations
- [**Ksctl Cluster Management (kcm)**][kcm-gh-link] - Kubernetes controller for managing ksctl-specific cluster addons
- [**Ksctl Application (ka)**][ka-gh-link] - Controller for ksctl application stack functionality

## 🚀 Current Capabilities

- **Cluster Operations**
  - Create, delete, and get cluster infrastructure details
  - State storage in local system or MongoDB
  - Manual scaling up and down via CLI
  - Switch between clusters
  - Wasm and application stack deployment

- **Supported Cluster Types**
  - **Self-Managed HA Clusters**:
    - K3s
    - Kubeadm
  - **Cloud Managed Clusters**:
    - AKS (Azure)
    - EKS (AWS)

- **Lifecycle Management**
  - Deploy Wasm workloads
  - Common application stack deployment (GitOps, monitoring, etc.)

## 🚦 Getting Started

Visit our [Getting Started guide][docs-gettingstarted] for installation instructions and initial setup.

```bash
# Install ksctl (Linux/macOS)
curl -sSfL https://get.ksctl.com | python3 -

# Example: Create a cluster (see documentation for full instructions)
ksctl cluster create
```

## 👥 Community & Contribution

We welcome contributions from the community!

- Read our [contribution guide][contribution-link] to get started
- Join our [Discord server][discord-link] to connect with other users and developers
- Attend our monthly community call on Tuesday 5:30-6:30 PM (CET) by joining our [Google Group](https://groups.google.com/g/ksctl)

⭐ **[Star our repository](https://github.com/ksctl/ksctl)** to help others discover the project and motivate our team! ⭐

## 💼 Sponsorship

If you find this project valuable and would like to provide financial support, please visit our [sponsorship page](https://github.com/sponsors/dipankardas011). Thank you for considering!

## 📜 License

Apache License 2.0, see [LICENSE][license].

## 🙌 Contributors

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
[discord-link]:https://discord.ksctl.com
[code-of-conduct]:https://github.com/ksctl/ksctl/blob/main/CODE_OF_CONDUCT.md
[governance]:https://github.com/ksctl/ksctl/blob/main/GOVERNANCE.md
[license]:https://github.com/ksctl/ksctl/blob/main/LICENSE
