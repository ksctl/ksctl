package helpers

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
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
		Protocol:    consts.FirewallActionTCP,
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

func firewallRuleNodePort(cidr string) FirewallRule {
	return FirewallRule{
		Name:        "kubernetes_nodeport",
		Description: "NodePort Services",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "30000",
		EndPort:   "35000",
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

func firewallRuleNodePortWorkerNodes(cidr string) FirewallRule {
	rule := firewallRuleNodePort(cidr)
	rule.Description = "NodePort Services for kubeadm"
	rule.EndPort = "32767"

	return rule
}

func FirewallForControlplane_BASE(cidr string, distro consts.KsctlKubernetes) []FirewallRule {

	rules := []FirewallRule{
		firewallRuleKubeApiServer(cidr),
		firewallRuleKubeletApi(cidr),
		firewallRuleNodePort(cidr),
		firewallRuleSSH(),
		firewallRuleOutBoundAllUDP(),
		firewallRuleOutBoundAllTCP(),
	}

	if distro == consts.K8sK3s {
		rules = append(rules, firewallRuleFlannel_VXLAN(cidr))
	}

	return rules
}

func FirewallForWorkerplane_BASE(cidr string, distro consts.KsctlKubernetes) []FirewallRule {

	rules := []FirewallRule{
		firewallRuleKubeletApi(cidr),
		firewallRuleSSH(),
		firewallRuleOutBoundAllUDP(),
		firewallRuleOutBoundAllTCP(),
	}

	switch distro {
	case consts.K8sK3s:
		rules = append(rules, firewallRuleFlannel_VXLAN(cidr))
	case consts.K8sKubeadm:
		rules = append(rules,
			firewallRuleKubeProxy(cidr),
			firewallRuleNodePortWorkerNodes(cidr),
		)
	}

	return rules
}

func FirewallForLoadBalancer_BASE() []FirewallRule {
	return []FirewallRule{
		firewallRuleKubeApiServer("0.0.0.0/0"),
		firewallRuleNodePort("0.0.0.0/0"),
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
