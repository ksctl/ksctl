// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package providers

// TODO(@dipankardas011): there are some k3s and kubeadm specific firewall rules so need to think of this

import (
	"github.com/ksctl/ksctl/pkg/consts"
)

type FirewallRule struct {
	Description string
	Name        string
	Protocol    consts.FirewallRuleProtocol
	Direction   consts.FirewallRuleDirection
	Action      consts.FirewallRuleAction

	Cidr      string
	StartPort string
	EndPort   string
}

func firewallRuleSSH() FirewallRule {
	return FirewallRule{
		Name:        "ksctl_ssh",
		Description: "SSH port for ksctl to work",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "22",
		EndPort:   "22",
	}
}

func firewallRuleOutBoundAllUDP() FirewallRule {
	return FirewallRule{
		Name:        "all_udp_outgoing",
		Description: "enable all the UDP outgoing traffic",
		Protocol:    consts.FirewallActionUDP,
		Direction:   consts.FirewallActionEgress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "1",
		EndPort:   "65535",
	}
}

func firewallRuleOutBoundAllTCP() FirewallRule {
	return FirewallRule{
		Name:        "all_tcp_outgoing",
		Description: "enable all the TCP outgoing traffic",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionEgress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "1",
		EndPort:   "65535",
	}
}

func firewallRuleKubeApiServer(cidr string) FirewallRule {
	return FirewallRule{
		Name:        "kubernetes_api_server",
		Description: "Kubernetes API Server",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "6443",
		EndPort:   "6443",
	}
}

func firewallRuleKubeletApi(cidr string) FirewallRule {
	return FirewallRule{
		Name:        "kubelet_api",
		Description: "Kubelet API",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "10250",
		EndPort:   "10250",
	}
}

func firewallRuleFlannel_VXLAN(cidr string) FirewallRule {
	return FirewallRule{
		Name:        "cni_flannel_vxlan",
		Description: "Required only for Flannel VXLAN",
		Protocol:    consts.FirewallActionUDP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "8472",
		EndPort:   "8472",
	}
}

func firewallRuleKubeProxy(cidr string) FirewallRule {
	return FirewallRule{
		Name:        "kubernetes_kube_proxy",
		Description: "kube-proxy",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "10256",
		EndPort:   "10256",
	}
}

func firewallRuleNodePort() FirewallRule {
	return FirewallRule{
		Name:        "kubernetes_nodeport",
		Description: "NodePort Services",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "30000",
		EndPort:   "32767",
	}
}

func firewallRuleEtcd(cidr string) FirewallRule {
	return FirewallRule{
		Name:        "etcd",
		Description: "For HA with external etcd",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "2379",
		EndPort:   "2380",
	}
}

func FirewallForControlplane_BASE(cidr string, distro consts.KsctlKubernetes) []FirewallRule {

	rules := []FirewallRule{
		firewallRuleKubeApiServer(cidr),
		firewallRuleKubeletApi(cidr),
		firewallRuleNodePort(),
		firewallRuleSSH(),
		firewallRuleOutBoundAllUDP(),
		firewallRuleOutBoundAllTCP(),
		firewallRuleFlannel_VXLAN(cidr),
	}

	return rules
}

func FirewallForWorkerplane_BASE(cidr string, distro consts.KsctlKubernetes) []FirewallRule {

	rules := []FirewallRule{
		firewallRuleKubeletApi(cidr),
		firewallRuleSSH(),
		firewallRuleNodePort(),
		firewallRuleOutBoundAllUDP(),
		firewallRuleOutBoundAllTCP(),
		firewallRuleFlannel_VXLAN(cidr),
	}

	if distro == consts.K8sKubeadm {
		rules = append(rules,
			firewallRuleKubeProxy(cidr),
		)
	}

	return rules
}

func FirewallForLoadBalancer_BASE() []FirewallRule {
	return []FirewallRule{
		firewallRuleKubeApiServer("0.0.0.0/0"),
		firewallRuleSSH(),
		firewallRuleOutBoundAllUDP(),
		firewallRuleOutBoundAllTCP(),
	}
}

func FirewallForDataStore_BASE(cidr string) []FirewallRule {
	return []FirewallRule{
		firewallRuleEtcd(cidr),
		firewallRuleSSH(),
		firewallRuleOutBoundAllUDP(),
		firewallRuleOutBoundAllTCP(),
	}
}
