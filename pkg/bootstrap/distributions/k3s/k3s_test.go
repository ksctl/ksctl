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

package k3s

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ksctl/ksctl/pkg/certs"
	"github.com/ksctl/ksctl/pkg/ssh"
	testHelper "github.com/ksctl/ksctl/pkg/ssh"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/provider"
	"github.com/ksctl/ksctl/pkg/statefile"
	"gotest.tools/v3/assert"
)

func NewClientHelper(x provider.CloudResourceState, state *statefile.StorageDocument) *K3s {

	p := &K3s{mu: &sync.Mutex{}}

	p.ctx = parentCtx
	p.l = parentLogger

	p.state = state
	p.state.K8sBootstrap = &statefile.KubernetesBootstrapState{}
	var err error
	p.state.K8sBootstrap.B.CACert, p.state.K8sBootstrap.B.EtcdCert, p.state.K8sBootstrap.B.EtcdKey, err = certs.GenerateCerts(parentCtx, parentLogger, x.PrivateIPv4DataStores)
	if err != nil {
		return nil
	}

	p.state.K8sBootstrap.B.PublicIPs.ControlPlanes = x.IPv4ControlPlanes
	p.state.K8sBootstrap.B.PrivateIPs.ControlPlanes = x.PrivateIPv4ControlPlanes

	p.state.K8sBootstrap.B.PublicIPs.DataStores = x.IPv4DataStores
	p.state.K8sBootstrap.B.PrivateIPs.DataStores = x.PrivateIPv4DataStores

	p.state.K8sBootstrap.B.PublicIPs.WorkerPlanes = x.IPv4WorkerPlanes

	p.state.K8sBootstrap.B.PublicIPs.LoadBalancer = x.IPv4LoadBalancer
	p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer = x.PrivateIPv4LoadBalancer
	p.state.K8sBootstrap.B.SSHInfo = statefile.SSHInfo{
		UserName:   x.SSHUserName,
		PrivateKey: x.SSHPrivateKey,
	}

	return p
}

func TestK3sDistro_Version(t *testing.T) {
	forTesting := map[string]bool{
		"":       true,
		"1.30.3": true,
		"1.27":   false,
	}
	for ver, expected := range forTesting {
		v, err := fakeClient.isValidK3sVersion(ver)
		got := err == nil

		if got != expected {
			t.Fatalf("Expected for %s as %v but got %v", ver, expected, got)
		} else {
			if err == nil {
				if len(ver) == 0 {
					assert.Equal(t, v, "v1.30.3+k3s1", "it should be equal")
				} else {
					assert.Equal(t, v, convertK3sVersion(ver), "it should be equal")
				}
			}
		}
	}
}

func TestScriptsControlplane(t *testing.T) {

	ver := []string{"1.26.1", "1.27"}
	privIP := []string{"9.9.9.9", "1.1.1.1"}
	dbEndpoint := getEtcdMemberIPFieldForControlplane(privIP)
	pubIPLb := []string{"192.16.9.2", "23.34.4.1"}
	privateIPLb := []string{"192.16.3.2", "1.1.1.1"}
	ca, etcd, key := "-- CA_CERT --", "-- ETCD_CERT --", "-- ETCD_KEY --"

	sampleToken := "k3ssdcdsXXXYYYZZZ"

	t.Run("script for controlplane 0", func(t *testing.T) {

		t.Run("script without cni", func(t *testing.T) {

			for i := 0; i < len(ver); i++ {

				testHelper.HelperTestTemplate(
					t,
					[]ssh.Script{
						getScriptForEtcdCerts(ca, etcd, key),
						{
							Name:           "Start K3s Controlplane-[0] without CNI",
							MaxRetries:     9,
							CanRetry:       true,
							ScriptExecutor: consts.LinuxBash,
							ShellScript: fmt.Sprintf(`
cat <<EOF > control-setup.sh
#!/bin/bash

/bin/bash /usr/local/bin/k3s-uninstall.sh || echo "already deleted"

curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh &>> ksctl.log
`, ver[i], dbEndpoint, pubIPLb[i], privateIPLb[i]),
						},
					},
					func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
						return scriptCP_1WithoutCNI(ca, etcd, key, ver[i], privIP, pubIPLb[i], privateIPLb[i])
					},
				)

			}
		})
		t.Run("script with cni", func(t *testing.T) {

			for i := 0; i < len(ver); i++ {
				testHelper.HelperTestTemplate(
					t,
					[]ssh.Script{
						getScriptForEtcdCerts(ca, etcd, key),
						{
							Name:           "Start K3s Controlplane-[0] with CNI",
							MaxRetries:     9,
							CanRetry:       true,
							ScriptExecutor: consts.LinuxBash,
							ShellScript: fmt.Sprintf(`
cat <<EOF > control-setup.sh
#!/bin/bash
/bin/bash /usr/local/bin/k3s-uninstall.sh || echo "already deleted"
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--tls-san %s \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh &>> ksctl.log
`, ver[i], dbEndpoint, pubIPLb[i], privateIPLb[i]),
						},
					},
					func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
						return scriptCP_1(ca, etcd, key, ver[i], privIP, pubIPLb[i], privateIPLb[i])
					},
				)
			}
		})
	})

	t.Run("script for controlplane 1..N", func(t *testing.T) {

		t.Run("script without cni", func(t *testing.T) {

			for i := 0; i < len(ver); i++ {

				testHelper.HelperTestTemplate(
					t,
					[]ssh.Script{
						getScriptForEtcdCerts(ca, etcd, key),
						{
							Name:           "Start K3s Controlplane-[1..N] without CNI",
							MaxRetries:     9,
							CanRetry:       true,
							ScriptExecutor: consts.LinuxBash,
							ShellScript: fmt.Sprintf(`
cat <<EOF > control-setupN.sh
#!/bin/bash
/bin/bash /usr/local/bin/k3s-uninstall.sh || echo "already deleted"
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--token %s \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--flannel-backend=none \
	--disable-network-policy \
	--server https://%s:6443 \
    --tls-san %s \
    --tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh &>> ksctl.log
`, ver[i], sampleToken, dbEndpoint, privateIPLb[i], pubIPLb[i], privateIPLb[i]),
						},
					},
					func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
						return scriptCP_NWithoutCNI(ca, etcd, key, ver[i], privIP, pubIPLb[i], privateIPLb[i], sampleToken)
					},
				)
			}
		})
		t.Run("script with cni", func(t *testing.T) {

			for i := 0; i < len(ver); i++ {

				testHelper.HelperTestTemplate(
					t,
					[]ssh.Script{
						getScriptForEtcdCerts(ca, etcd, key),
						{
							Name:           "Start K3s Controlplane-[1..N] with CNI",
							MaxRetries:     9,
							CanRetry:       true,
							ScriptExecutor: consts.LinuxBash,
							ShellScript: fmt.Sprintf(`
cat <<EOF > control-setupN.sh
#!/bin/bash
/bin/bash /usr/local/bin/k3s-uninstall.sh || echo "already deleted"
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--token %s \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--server https://%s:6443 \
    --tls-san %s \
    --tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh &>> ksctl.log
`, ver[i], sampleToken, dbEndpoint, privateIPLb[i], pubIPLb[i], privateIPLb[i]),
						},
					},
					func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
						return scriptCP_N(ca, etcd, key, ver[i], privIP, pubIPLb[i], privateIPLb[i], sampleToken)
					},
				)
			}
		})
	})

	t.Run("get k3s token", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "Get k3s server token",
					CanRetry:       false,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: `
sudo cat /var/lib/rancher/k3s/server/token
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptForK3sToken()
			},
		)
	})

	t.Run("get kubeconfig", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "k3s kubeconfig",
					CanRetry:       false,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: `
sudo cat /etc/rancher/k3s/k3s.yaml
`,
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptKUBECONFIG()
			},
		)
	})

}

func TestSciprWorkerplane(t *testing.T) {
	ver := "1.18"
	token := "K#Sde43rew34"
	private := "192.20.3.3"

	t.Run("get kubeconfig", func(t *testing.T) {
		testHelper.HelperTestTemplate(
			t,
			[]ssh.Script{
				{
					Name:           "Join the workerplane-[0..M]",
					CanRetry:       true,
					MaxRetries:     3,
					ScriptExecutor: consts.LinuxBash,
					ShellScript: fmt.Sprintf(`
cat <<EOF > worker-setup.sh
#!/bin/bash
/bin/bash /usr/local/bin/k3s-agent-uninstall.sh || echo "already deleted"
export K3S_DEBUG=true
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - agent --token %s --server https://%s:6443
EOF

sudo chmod +x worker-setup.sh
sudo ./worker-setup.sh &>> ksctl.log
`, ver, token, private),
				},
			},
			func() ssh.ExecutionPipeline { // Adjust the signature to match your needs
				return scriptWP(ver, private, token)
			},
		)
	})
}

func checkCurrentStateFile(t *testing.T) {

	raw, err := storeHA.Read()
	if err != nil {
		t.Fatalf("Unable to access statefile")
	}

	assert.DeepEqual(t, fakeClient.state, raw)
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(consts.OperationCreate), nil, "should be initlize the state")

	fakeClient.K8sVersion("")

	noCP := len(fakeStateFromCloud.IPv4ControlPlanes)
	noWP := len(fakeStateFromCloud.IPv4WorkerPlanes)

	fakeClient.CNI("flannel")
	for no := 0; no < noCP; no++ {
		err := fakeClient.ConfigureControlPlane(no)
		if err != nil {
			t.Fatalf("Configure Controlplane unable to operate %v", err)
		}
	}
	checkCurrentStateFile(t)

	assert.Equal(t, fakeClient.state.K8sBootstrap.K3s.K3sVersion, "v1.30.3+k3s1", "should be equal")

	for no := 0; no < noWP; no++ {
		err := fakeClient.JoinWorkerplane(no)
		if err != nil {
			t.Fatalf("Configure Workerplane unable to operate %v", err)
		}
	}

}

func TestCNI(t *testing.T) {
	testCases := map[string]bool{
		string(consts.CNIFlannel): false,
		string(consts.CNICilium):  true,
	}

	for k, v := range testCases {
		got := fakeClient.CNI(k)
		assert.Equal(t, got, v, "missmatch")
	}
}

func TestGetEtcdMemberIPFieldForControlplane(t *testing.T) {
	ips := []string{"9.9.9.9", "1.1.1.1"}
	res1 := "https://9.9.9.9:2379,https://1.1.1.1:2379"
	assert.Equal(t, res1, getEtcdMemberIPFieldForControlplane(ips), "it should be equal")

	assert.Equal(t, "", getEtcdMemberIPFieldForControlplane([]string{}), "it should be equal")
}
