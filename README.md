![ksctl - Kubernetes Simplified][cover-img-loc]

# Ksctl: Simplified Kubernetes Clusters Lifecycle Management

> [!WARNING]
> **This project is no longer maintained.**
>
> What I build now:
>
> - **[Kubmin](https://ksctl.com/product/kubmin)** — per-workload cost, efficiency, and SCI/SEI intelligence for Kubernetes. Some parts of this repo were reused there, but the latest migration is removing that dependency.
> - **[infai](https://infai.dipankar-das.com)** — a terminal control plane for local LLM inference: one profile shape for llama.cpp and vLLM, many models at once, live tokens/sec plus CPU/RAM/VRAM telemetry, hardware fit checks before launch
> - **[infaiw](https://infai.dipankar-das.com/infaiw/)** — a daemon-based agent harness: sessions keep running after you close the terminal, sidecar loops must pass an acceptance check before reporting back, every event lands on a durable branchable timeline, and approvals wait for you
>
> This repo is the core engine behind [kli](https://github.com/ksctl/kli): the cluster provisioning and lifecycle logic that the `kli` CLI wrapped. `kli` is no longer maintained either.
>
> **Want to take over maintenance of `ksctl`?** If you still depend on it and would like to become the maintainer, reach out on [Discord](https://discord.ksctl.com) or [open an issue](https://github.com/ksctl/ksctl/issues) on this repo.
>
> The code here stays available for reference under its original license.


<h3>Making Kubernetes accessible to developers</h3>

[![Discord](https://img.shields.io/badge/discord-ksctl-brightgreen.svg)][discord-link]
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)][license]
[![X/Twitter][x-badge]][x-link]
[![Go Report Card](https://goreportcard.com/badge/github.com/ksctl/ksctl)](https://goreportcard.com/report/github.com/ksctl/ksctl)
[![](https://pkg.go.dev/badge/github.com/ksctl/ksctl.svg)](https://pkg.go.dev/github.com/ksctl/ksctl)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/7469/badge)](https://www.bestpractices.dev/projects/7469)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/ksctl/ksctl)

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
