package helpers

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"

	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"

	"gotest.tools/v3/assert"
)

var (
	dir                         = fmt.Sprintf("%s/ksctl-k3s-test", os.TempDir())
	log resources.LoggerFactory = func() resources.LoggerFactory {
		var l resources.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
		l.SetPackageName("utils")
		return l
	}()
	mainStateDoc = &types.StorageDocument{}
)

func TestConsts(t *testing.T) {
	assert.Equal(t, string(consts.CloudCivo), "civo")
	assert.Equal(t, string(consts.CloudAzure), "azure")
	assert.Equal(t, string(consts.CloudLocal), "local")
	assert.Equal(t, string(consts.K8sK3s), "k3s")
	assert.Equal(t, string(consts.CloudAws), "aws")
	assert.Equal(t, string(consts.K8sKubeadm), "kubeadm")
	assert.Equal(t, string(consts.StoreLocal), "local")
	assert.Equal(t, string(consts.StoreExtMongo), "external-mongo")
	assert.Equal(t, string(consts.RoleCp), "controlplane")
	assert.Equal(t, string(consts.RoleLb), "loadbalancer")
	assert.Equal(t, string(consts.RoleDs), "datastore")
	assert.Equal(t, string(consts.RoleWp), "workerplane")
	assert.Equal(t, string(consts.ClusterTypeHa), "ha")
	assert.Equal(t, string(consts.ClusterTypeMang), "managed")
	assert.Equal(t, string(consts.OperationCreate), "create")
	assert.Equal(t, string(consts.OperationGet), "get")
	assert.Equal(t, string(consts.OperationDelete), "delete")
	assert.Equal(t, uint8(consts.UtilClusterPath), uint8(1))
	assert.Equal(t, uint8(consts.UtilOtherPath), uint8(3))
	assert.Equal(t, uint8(consts.UtilSSHPath), uint8(2))
	assert.Equal(t, uint8(consts.UtilCredentialPath), uint8(0))
	assert.Equal(t, uint8(consts.UtilExecWithoutOutput), uint8(0))
	assert.Equal(t, uint8(consts.UtilExecWithOutput), uint8(1))

	assert.Equal(t, string(consts.CNIAzure), "azure")
	assert.Equal(t, string(consts.CNIKind), "kind")
	assert.Equal(t, string(consts.CNINone), "none")
	assert.Equal(t, string(consts.CNICilium), "cilium")
	assert.Equal(t, string(consts.CNIFlannel), "flannel")
	assert.Equal(t, string(consts.CNIKubenet), "kubenet")
	assert.Equal(t, string(consts.LinuxSh), "/bin/sh")
	assert.Equal(t, string(consts.LinuxBash), "/bin/bash")

	assert.Equal(t, int(consts.FirewallActionAllow), 0)
	assert.Equal(t, int(consts.FirewallActionDeny), 1)
	assert.Equal(t, int(consts.FirewallActionIngress), 2)
	assert.Equal(t, int(consts.FirewallActionEgress), 3)
	assert.Equal(t, int(consts.FirewallActionTCP), 4)
	assert.Equal(t, int(consts.FirewallActionUDP), 5)
}

func TestGenerateCerts(t *testing.T) {
	if ca, etcd, key, err := GenerateCerts(log, []string{"192.168.1.1"}); err != nil {
		t.Fatalf("it shouldn't fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}

	if ca, etcd, key, err := GenerateCerts(log, []string{"192,168.1.1"}); err == nil ||
		len(ca) != 0 ||
		len(etcd) != 0 ||
		len(key) != 0 {
		t.Fatalf("it should fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}
}

func TestGetUsername(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, os.Getenv(UserDir), GetUserName(), "Unable to fetch correct username")
	} else {
		assert.Equal(t, os.Getenv(UserDir), GetUserName(), "Unable to fetch correct username")
	}
}

func TestCNIValidation(t *testing.T) {
	cnitests := map[string]bool{
		string(consts.CNIAzure):   true,
		string(consts.CNIKubenet): true,
		string(consts.CNIFlannel): true,
		string(consts.CNICilium):  true,
		string(consts.CNIKind):    true,
		"abcd":                    false,
		"":                        true,
	}
	for k, v := range cnitests {
		assert.Equal(t, v, ValidCNIPlugin(consts.KsctlValidCNIPlugin(k)), "")
	}
}

func TestCreateSSHKeyPair(t *testing.T) {
	err := CreateSSHKeyPair(log, mainStateDoc)
	if err != nil {
		t.Fatal(err)
	}
	dump.Println(mainStateDoc.SSHKeyPair)
}

func TestIsValidClusterName(T *testing.T) {
	errorStr := fmt.Errorf("CLUSTER NAME INVALID")
	assert.Equal(T, nil, IsValidName("demo"), "Returns false for valid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("Dem-o234").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, nil, IsValidName("d-234"), "Returns false for valid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("234").Error(), "Returns true for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("-2342").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("demo-").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("dscdscsd-#$#$#").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("ds@#$#$#").Error(), "Returns True for invalid cluster name")
}

func TestSSHExecute(t *testing.T) {

	var sshTest SSHCollection = &SSHPayload{}
	sshTest.Username("fake")
	sshTest.PrivateKey(mainStateDoc.SSHKeyPair.PrivateKey)
	assert.Assert(t, sshTest.Flag(consts.UtilExecWithoutOutput).Script(NewScriptCollection()).
		IPv4("A.A.A.A").
		FastMode(true).SSHExecute(log) != nil, "ssh should fail")

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

}

func TestIsValidVersion(t *testing.T) {
	testCases := map[string]bool{
		"1.1.1":  true,
		"latest": true,
		"v1":     true,
		"v1.1":   true,
		"v1.1.1": true,
		"1.1":    true,
		"1":      true,
		"v":      false,
		"stable": true,
	}

	for ver, expected := range testCases {
		err := IsValidVersion(ver)
		var got bool = err == nil
		assert.Equal(t, got, expected, fmt.Sprintf("Ver: %s, got: %v, expected: %v", ver, got, expected))
	}
}

func TestToApplicationTempl(t *testing.T) {
	testCases := []struct {
		inp               string
		Expected          types.Application
		ExpectedIsInvalid bool
	}{
		{
			inp:      "abcd@latest",
			Expected: types.Application{Name: "abcd", Version: "latest"},
		},
		{
			inp:      "abcd@123",
			Expected: types.Application{Name: "abcd", Version: "123"},
		},
		{
			inp:      "abcd",
			Expected: types.Application{Name: "abcd", Version: "latest"},
		},
		{
			inp:               "",
			ExpectedIsInvalid: true,
		},
		{
			inp:               "abcd@lald@",
			ExpectedIsInvalid: true,
		},
	}

	for _, testcase := range testCases {
		got, err := ToApplicationTempl([]string{testcase.inp})
		gotErr := err != nil

		t.Logf("App: %v\n", testcase.inp)

		assert.Check(t, gotErr == testcase.ExpectedIsInvalid,
			fmt.Sprintf("app: %v, got: %v, expected: %v", testcase.inp, gotErr, testcase.ExpectedIsInvalid))
		if len(got) != 0 {
			assert.DeepEqual(t, got[0], testcase.Expected)
		}
	}
}

func TestScriptCollection(t *testing.T) {
	var scripts *Scripts = func() *Scripts {
		o := NewScriptCollection()
		switch v := o.(type) {
		case *Scripts:
			return v
		default:
			return nil
		}
	}()

	t.Run("init state test", func(t *testing.T) {
		assert.Equal(t, scripts.currIdx, -1, "must be initialized with -1")
		assert.Assert(t, scripts.mu != nil, "the mutext variable should be initialized!")
		assert.Assert(t, scripts.data == nil)
		assert.Equal(t, scripts.IsCompleted(), false, "must be empty")
	})

	t.Run("append scripts", func(t *testing.T) {
		datas := []resources.Script{
			{
				ScriptExecutor: consts.LinuxBash,
				CanRetry:       false,
				Name:           "test",
				MaxRetries:     0,
				ShellScript:    "script",
			},
			{
				ScriptExecutor: consts.LinuxSh,
				CanRetry:       true,
				Name:           "x test",
				MaxRetries:     9,
				ShellScript:    "demo",
			},
		}

		for idx, data := range datas {
			scripts.Append(data)
			data.ShellScript = "#!" + string(data.ScriptExecutor) + "\n" + data.ShellScript

			assert.Equal(t, scripts.currIdx, 0, "the first element added so index should be 0")
			assert.DeepEqual(t, scripts.data[idx], data)
		}

	})

	t.Run("get script", func(t *testing.T) {
		v := scripts.NextScript()

		expected := &resources.Script{
			ScriptExecutor: consts.LinuxBash,
			CanRetry:       false,
			Name:           "test",
			MaxRetries:     0,
			ShellScript:    "#!/bin/bash\nscript",
		}

		assert.DeepEqual(t, v, expected)
		assert.Equal(t, scripts.currIdx, 1, "the index must increment")
	})
}

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
			FirewallForControlplane_BASE(
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
			FirewallForControlplane_BASE(
				cidr, consts.K8sKubeadm),
			[]FirewallRule{
				expectedK8sApiServer,
				expectedKubeletApi,
				expectedNodePort,
				expectedSSH,
				expectedUdp,
				expectedTcp,
			})
	})

	t.Run("firewallRule for WorkerPlane", func(t *testing.T) {
		assert.DeepEqual(t,
			FirewallForWorkerplane_BASE(
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
			FirewallForWorkerplane_BASE(
				cidr, consts.K8sKubeadm),
			[]FirewallRule{
				expectedKubeletApi,
				expectedSSH,
				expectedNodePort,
				expectedUdp,
				expectedTcp,
				expectedKubeProxy,
			})
	})
	t.Run("firewallRule for LoadBalancer", func(t *testing.T) {
		assert.DeepEqual(t,
			FirewallForLoadBalancer_BASE(),
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
			FirewallForDataStore_BASE(cidr),
			[]FirewallRule{
				expectedEtcd,
				expectedSSH,
				expectedUdp,
				expectedTcp,
			})
	})
}
