package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"

	"gotest.tools/v3/assert"
)

var (
	dir                              = filepath.Join(os.TempDir(), "ksctl-k3s-test")
	log          types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
	mainStateDoc                     = &storageTypes.StorageDocument{}
	dummyCtx                         = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
)

func TestConsts(t *testing.T) {
	assert.Equal(t, string(consts.CloudCivo), "civo")
	assert.Equal(t, string(consts.CloudAzure), "azure")
	assert.Equal(t, string(consts.CloudLocal), "local")
	assert.Equal(t, string(consts.K8sK3s), "k3s")
	assert.Equal(t, string(consts.CloudAws), "aws")
	assert.Equal(t, string(consts.K8sKubeadm), "kubeadm")
	assert.Equal(t, string(consts.StoreLocal), "store-local")
	assert.Equal(t, string(consts.StoreK8s), "store-kubernetes")
	assert.Equal(t, string(consts.StoreExtMongo), "external-store-mongodb")
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
	if ca, etcd, key, err := GenerateCerts(
		context.WithValue(
			context.TODO(),
			consts.KsctlModuleNameKey,
			"demo"),
		log, []string{"192.168.1.1"}); err != nil {
		t.Fatalf("it shouldn't fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
	}

	if ca, etcd, key, err := GenerateCerts(
		context.WithValue(context.TODO(), consts.KsctlModuleNameKey, "demo"), log, []string{"192,168.1.1"}); err == nil ||
		len(ca) != 0 ||
		len(etcd) != 0 ||
		len(key) != 0 {
		t.Fatalf("it should fail, ca: %v, etcd: %v, key: %v, err: %v\n", ca, etcd, key, err)
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
	err := CreateSSHKeyPair(dummyCtx, log, mainStateDoc)
	if err != nil {
		t.Fatal(err)
	}
	dump.Println(mainStateDoc.SSHKeyPair)
}

func TestIsValidClusterName(t *testing.T) {
	assert.Check(t, nil == IsValidName(dummyCtx, log, "demo"), "Returns false for valid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "Dem-o234")
			return err != nil && ksctlErrors.ErrInvalidResourceName.Is(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(t, nil == IsValidName(dummyCtx, log, "d-234"), "Returns false for valid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "234")
			return err != nil && ksctlErrors.ErrInvalidResourceName.Is(err)
		}(),
		"Returns true for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "-2342")
			return err != nil && ksctlErrors.ErrInvalidResourceName.Is(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "demo-")
			return err != nil && ksctlErrors.ErrInvalidResourceName.Is(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "dscdscsd-#$#$#")
			return err != nil && ksctlErrors.ErrInvalidResourceName.Is(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "dds@#$#$#ds@#$#$#ds@#$#$#ds@#$#$#ds@#$#$#s@#$#$wefe#")
			return err != nil && ksctlErrors.ErrInvalidResourceName.Is(err)
		}(),
		"Returns True for invalid cluster name")
}

func TestSSHExecute(t *testing.T) {

	var sshTest SSHCollection = &SSHPayload{
		ctx: dummyCtx,
		log: log,
	}
	testSimulator := NewScriptCollection()
	testSimulator.Append(types.Script{
		Name:           "test",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
cat /etc/os-releases
`,
	})
	testSimulator.Append(types.Script{
		Name:           "testhaving retry",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
suao apt install ...
`,
	})
	sshTest.Username("fake")
	sshTest.PrivateKey(mainStateDoc.SSHKeyPair.PrivateKey)
	sshT := sshTest.Flag(consts.UtilExecWithoutOutput).Script(testSimulator).
		IPv4("A.A.A.A").
		FastMode(true).SSHExecute()
	assert.Assert(t, sshT == nil, fmt.Sprintf("ssh should fail, got: %v, exepected ! nil", sshT))

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

}

func TestIsValidVersion(t *testing.T) {
	testCases := map[string]bool{
		"1.1.1":            true,
		"latest":           true,
		"v1":               true,
		"v1.1":             true,
		"v1.1.1":           true,
		"1.1":              true,
		"1":                true,
		"v":                false,
		"stable":           true,
		"enhancement-2342": true,
		"enhancement":      true,
		"feature-2342":     true,
		"feature":          true,
		"feat":             true,
		"feat234":          true,
		"fix234":           true,
		"f14cd9094b2160c40ef8734e90141df81c22999e": true,
	}

	for ver, expected := range testCases {
		err := IsValidKsctlComponentVersion(dummyCtx, log, ver)
		var got bool = err == nil
		assert.Equal(t, got, expected, fmt.Sprintf("Ver: %s, got: %v, expected: %v", ver, got, expected))
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
		datas := []types.Script{
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

		expected := &types.Script{
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
				expectedFlannelVXLan,
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
				expectedFlannelVXLan,
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

func TestBackOffRun_SuccessOnFirstAttempt(t *testing.T) {
	ctx := context.Background()

	executeFunc := func() error {
		return nil
	}

	isSuccessful := func() bool {
		return true
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewBackOff(1*time.Second, 1, 3)

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err == nil)
}

func TestBackOffRun_RetryOnFailure(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	executeFunc := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("execute error")
		}
		return nil
	}

	isSuccessful := func() bool {
		return callCount == 3
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewBackOff(1*time.Second, 1, 3)

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err == nil)

	assert.Equal(t, 3, callCount)
}

func TestBackOffRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	executeFunc := func() error {
		return errors.New("execute error")
	}

	isSuccessful := func() bool {
		return false
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewBackOff(1*time.Second, 1, 3)

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err != nil && ksctlErrors.ErrContextCancelled.Is(err))

	assert.Equal(t, context.Canceled, ctx.Err())
}

func TestBackOffRun_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()

	executeFunc := func() error {
		return errors.New("execute error")
	}

	isSuccessful := func() bool {
		return false
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewBackOff(1*time.Second, 1, 3)

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err != nil && ksctlErrors.ErrTimeOut.Is(err))
}
