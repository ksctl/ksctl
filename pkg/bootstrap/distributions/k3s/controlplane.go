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
	"strconv"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/ssh"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

// NOTE: configureCP_1 is not meant for concurrency
func (p *K3s) configureCP_1(sshExecutor ssh.RemoteConnection) error {

	var script ssh.ExecutionPipeline

	if consts.KsctlValidCNIPlugin(p.Cni) != consts.CNINone {
		p.state.ProvisionerAddons.Cni = statefile.SlimProvisionerAddon{
			Name: p.Cni,
			For:  consts.K8sK3s,
		}
	}

	if consts.KsctlValidCNIPlugin(p.Cni) == consts.CNINone {
		script = scriptCP_1WithoutCNI(
			p.state.K8sBootstrap.B.CACert,
			p.state.K8sBootstrap.B.EtcdCert,
			p.state.K8sBootstrap.B.EtcdKey,
			*p.state.Versions.K3s,
			p.state.K8sBootstrap.B.PrivateIPs.DataStores,
			p.state.K8sBootstrap.B.PublicIPs.LoadBalancer,
			p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer)
	} else {
		script = scriptCP_1(
			p.state.K8sBootstrap.B.CACert,
			p.state.K8sBootstrap.B.EtcdCert,
			p.state.K8sBootstrap.B.EtcdKey,
			*p.state.Versions.K3s,
			p.state.K8sBootstrap.B.PrivateIPs.DataStores,
			p.state.K8sBootstrap.B.PublicIPs.LoadBalancer,
			p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer)
	}

	err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(script).
		IPv4(p.state.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		FastMode(true).SSHExecute()
	if err != nil {
		return err
	}

	// K3stoken
	err = sshExecutor.Flag(consts.UtilExecWithOutput).Script(scriptForK3sToken()).
		IPv4(p.state.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute()
	if err != nil {
		return err
	}

	p.l.Debug(p.ctx, "fetching k3s token")

	p.state.K8sBootstrap.K3s.K3sToken = strings.Trim(sshExecutor.GetOutput()[0], "\n")

	err = p.store.Write(p.state)
	if err != nil {
		return err
	}
	return nil
}

func (p *K3s) ConfigureControlPlane(noOfCP int) error {
	p.mu.Lock()
	idx := noOfCP
	sshExecutor := ssh.NewSSHExecutor(p.ctx, p.l, p.state) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	p.l.Note(p.ctx, "configuring ControlPlane", "number", strconv.Itoa(idx))
	if idx == 0 {
		err := p.configureCP_1(sshExecutor)
		if err != nil {
			return err
		}
	} else {

		var script ssh.ExecutionPipeline

		if consts.KsctlValidCNIPlugin(p.Cni) == consts.CNINone {
			script = scriptCP_NWithoutCNI(
				p.state.K8sBootstrap.B.CACert,
				p.state.K8sBootstrap.B.EtcdCert,
				p.state.K8sBootstrap.B.EtcdKey,
				*p.state.Versions.K3s,
				p.state.K8sBootstrap.B.PrivateIPs.DataStores,
				p.state.K8sBootstrap.B.PublicIPs.LoadBalancer,
				p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer,
				p.state.K8sBootstrap.K3s.K3sToken)
		} else {
			script = scriptCP_N(
				p.state.K8sBootstrap.B.CACert,
				p.state.K8sBootstrap.B.EtcdCert,
				p.state.K8sBootstrap.B.EtcdKey,
				*p.state.Versions.K3s,
				p.state.K8sBootstrap.B.PrivateIPs.DataStores,
				p.state.K8sBootstrap.B.PublicIPs.LoadBalancer,
				p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer,
				p.state.K8sBootstrap.K3s.K3sToken)
		}

		err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(script).
			IPv4(p.state.K8sBootstrap.B.PublicIPs.ControlPlanes[idx]).
			FastMode(true).SSHExecute()
		if err != nil {
			return err
		}

		if idx+1 == len(p.state.K8sBootstrap.B.PublicIPs.ControlPlanes) {

			p.l.Debug(p.ctx, "fetching kubeconfig")
			err = sshExecutor.Flag(consts.UtilExecWithOutput).Script(scriptKUBECONFIG()).
				IPv4(p.state.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
				FastMode(true).SSHExecute()
			if err != nil {
				return err
			}
			// as only a single case where it is getting invoked we actually don't need locks

			contextName := p.state.ClusterName + "-" + p.state.Region + "-" + string(p.state.ClusterType) + "-" + string(p.state.InfraProvider) + "-ksctl"
			kubeconfig := sshExecutor.GetOutput()[0]
			kubeconfig = strings.Replace(kubeconfig, "127.0.0.1", p.state.K8sBootstrap.B.PublicIPs.LoadBalancer, 1)
			kubeconfig = strings.Replace(kubeconfig, "default", contextName, -1)

			p.state.ClusterKubeConfig = kubeconfig
			p.state.ClusterKubeConfigContext = contextName

			err = p.store.Write(p.state)
			if err != nil {
				return err
			}
		}

	}
	p.l.Success(p.ctx, "configured ControlPlane", "number", strconv.Itoa(idx))

	return nil
}

func getScriptForEtcdCerts(ca, etcd, key string) ssh.Script {
	return ssh.Script{
		Name:           "store etcd certificates",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo mkdir -p /var/lib/etcd

cat <<EOF > ca.pem
%s
EOF

cat <<EOF > etcd.pem
%s
EOF

cat <<EOF > etcd-key.pem
%s
EOF

sudo mv -v ca.pem etcd.pem etcd-key.pem /var/lib/etcd
`, ca, etcd, key),
	}
}

func scriptCP_1WithoutCNI(ca, etcd, key, ver string, privateEtcdIps []string,
	pubIPlb, privIplb string) ssh.ExecutionPipeline {

	collection := ssh.NewExecutionPipeline()

	collection.Append(getScriptForEtcdCerts(ca, etcd, key))

	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)

	collection.Append(ssh.Script{
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
`, ver, dbEndpoint, pubIPlb, privIplb),
	})

	return collection
}

func scriptCP_1(ca, etcd, key, ver string, privateEtcdIps []string, pubIPlb,
	privateIPLb string) ssh.ExecutionPipeline {

	collection := ssh.NewExecutionPipeline()

	collection.Append(getScriptForEtcdCerts(ca, etcd, key))

	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)

	collection.Append(ssh.Script{
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
`, ver, dbEndpoint, pubIPlb, privateIPLb),
	})

	return collection
}

func scriptForK3sToken() ssh.ExecutionPipeline {

	collection := ssh.NewExecutionPipeline()
	collection.Append(ssh.Script{
		Name:           "Get k3s server token",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo cat /var/lib/rancher/k3s/server/token
`,
	})

	return collection
}

func scriptCP_N(ca, etcd, key, ver string, privateEtcdIps []string,
	pubIplb, privateIPlb, token string) ssh.ExecutionPipeline {

	collection := ssh.NewExecutionPipeline()

	collection.Append(getScriptForEtcdCerts(ca, etcd, key))

	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)

	collection.Append(ssh.Script{
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
`, ver, token, dbEndpoint, privateIPlb, pubIplb, privateIPlb),
	})

	return collection
}

func scriptCP_NWithoutCNI(ca, etcd, key, ver string, privateEtcdIps []string,
	pubIplb, privateIPlb, token string) ssh.ExecutionPipeline {

	collection := ssh.NewExecutionPipeline()

	collection.Append(getScriptForEtcdCerts(ca, etcd, key))

	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)

	collection.Append(ssh.Script{
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
`, ver, token, dbEndpoint, privateIPlb, pubIplb, privateIPlb),
	})

	return collection
}
