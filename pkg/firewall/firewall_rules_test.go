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

package firewall

import (
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"gotest.tools/v3/assert"
	"testing"
)

func TestFirewallRules(t *testing.T) {

	cidr := "x.y.z.a/b"
	expectedEtcd := FirewallRule{
		Name:        "etcd",
		Description: "For HA with external etcd",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "2379",
		EndPort:   "2380",
	}

	expectedSSH := FirewallRule{
		Name:        "ksctl_ssh",
		Description: "SSH port for ksctl to work",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "22",
		EndPort:   "22",
	}
	expectedUdp := FirewallRule{
		Name:        "all_udp_outgoing",
		Description: "enable all the UDP outgoing traffic",
		Protocol:    consts.FirewallActionUDP,
		Direction:   consts.FirewallActionEgress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "1",
		EndPort:   "65535",
	}
	expectedTcp := FirewallRule{
		Name:        "all_tcp_outgoing",
		Description: "enable all the TCP outgoing traffic",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionEgress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "1",
		EndPort:   "65535",
	}
	expectedK8sApiServer := FirewallRule{
		Name:        "kubernetes_api_server",
		Description: "Kubernetes API Server",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "6443",
		EndPort:   "6443",
	}
	expectedKubeletApi := FirewallRule{
		Name:        "kubelet_api",
		Description: "Kubelet API",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "10250",
		EndPort:   "10250",
	}
	expectedFlannelVXLan := FirewallRule{
		Name:        "cni_flannel_vxlan",
		Description: "Required only for Flannel VXLAN",
		Protocol:    consts.FirewallActionUDP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "8472",
		EndPort:   "8472",
	}
	expectedKubeProxy := FirewallRule{
		Name:        "kubernetes_kube_proxy",
		Description: "kube-proxy",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      cidr,
		StartPort: "10256",
		EndPort:   "10256",
	}
	expectedNodePort := FirewallRule{
		Name:        "kubernetes_nodeport",
		Description: "NodePort Services",
		Protocol:    consts.FirewallActionTCP,
		Direction:   consts.FirewallActionIngress,
		Action:      consts.FirewallActionAllow,

		Cidr:      "0.0.0.0/0",
		StartPort: "30000",
		EndPort:   "32767",
	}

	t.Run("ssh rule", func(t *testing.T) {
		got := firewallRuleSSH()
		assert.Equal(t, got, expectedSSH)
	})
	t.Run("allow all udp", func(t *testing.T) {
		got := firewallRuleOutBoundAllUDP()
		assert.Equal(t, got, expectedUdp)
	})
	t.Run("allow all tcp", func(t *testing.T) {
		got := firewallRuleOutBoundAllTCP()
		assert.Equal(t, got, expectedTcp)
	})

	t.Run("kube api server", func(t *testing.T) {
		got := firewallRuleKubeApiServer(cidr)
		assert.Equal(t, got, expectedK8sApiServer)
	})

	t.Run("kubelet api", func(t *testing.T) {
		got := firewallRuleKubeletApi(cidr)
		assert.Equal(t, got, expectedKubeletApi)
	})

	t.Run("cni flannel vxlan", func(t *testing.T) {
		got := firewallRuleFlannel_VXLAN(cidr)
		assert.Equal(t, got, expectedFlannelVXLan)
	})

	t.Run("kubernetes kube proxy", func(t *testing.T) {
		got := firewallRuleKubeProxy(cidr)
		assert.Equal(t, got, expectedKubeProxy)
	})

	t.Run("kubernetes nodeport", func(t *testing.T) {
		got := firewallRuleNodePort()
		assert.Equal(t, got, expectedNodePort)
	})

	t.Run("etcd", func(t *testing.T) {
		got := firewallRuleEtcd(cidr)
		assert.Equal(t, got, expectedEtcd)
	})

	t.Run("firewallRule for ControlPlane", func(t *testing.T) {
		assert.DeepEqual(t,
			FirewallforcontrolplaneBase(
				cidr, consts.K8sK3s),
			[]FirewallRule{
				expectedK8sApiServer,
				expectedKubeletApi,
				expectedNodePort,
				expectedSSH,
				expectedUdp,
				expectedTcp,
				expectedFlannelVXLan,
			})
		assert.DeepEqual(t,
			FirewallforcontrolplaneBase(
				cidr, consts.K8sKubeadm),
			[]FirewallRule{
				expectedK8sApiServer,
				expectedKubeletApi,
				expectedNodePort,
				expectedSSH,
				expectedUdp,
				expectedTcp,
				expectedFlannelVXLan,
			})
	})

	t.Run("firewallRule for WorkerPlane", func(t *testing.T) {
		assert.DeepEqual(t,
			FirewallforworkerplaneBase(
				cidr, consts.K8sK3s),
			[]FirewallRule{
				expectedKubeletApi,
				expectedSSH,
				expectedNodePort,
				expectedUdp,
				expectedTcp,
				expectedFlannelVXLan,
			})
		assert.DeepEqual(t,
			FirewallforworkerplaneBase(
				cidr, consts.K8sKubeadm),
			[]FirewallRule{
				expectedKubeletApi,
				expectedSSH,
				expectedNodePort,
				expectedUdp,
				expectedTcp,
				expectedFlannelVXLan,
				expectedKubeProxy,
			})
	})
	t.Run("firewallRule for LoadBalancer", func(t *testing.T) {
		assert.DeepEqual(t,
			FirewallforloadbalancerBase(),
			[]FirewallRule{
				{
					Name:        "kubernetes_api_server",
					Description: "Kubernetes API Server",
					Protocol:    consts.FirewallActionTCP,
					Direction:   consts.FirewallActionIngress,
					Action:      consts.FirewallActionAllow,

					Cidr:      "0.0.0.0/0",
					StartPort: "6443",
					EndPort:   "6443",
				},

				expectedSSH,
				expectedUdp,
				expectedTcp,
			})
	})
	t.Run("firewallRule for DataStore", func(t *testing.T) {
		assert.DeepEqual(t,
			FirewallfordatastoreBase(cidr),
			[]FirewallRule{
				expectedEtcd,
				expectedSSH,
				expectedUdp,
				expectedTcp,
			})
	})
}
