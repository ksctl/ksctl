package k3s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// configureCP_1 is not meant for concurrency
func configureCP_1(storage resources.StorageFactory, k3s *K3s, sshExecutor helpers.SSHCollection) error {

	var script string

	if consts.KsctlValidCNIPlugin(k3s.Cni) == consts.CNINone {
		script = scriptCP_1WithoutCNI(
			mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey,
			k3s.K3sVer,
			mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores,
			mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer)
	} else {
		script = scriptCP_1(
			mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey,
			k3s.K3sVer,
			mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores,
			mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer)
	}

	err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(script).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}

	// K3stoken
	err = sshExecutor.Flag(consts.UtilExecWithOutput).Script(scriptForK3sToken()).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Debug("fetching k3s token")

	mainStateDocument.K8sBootstrap.K3s.K3sToken = strings.Trim(sshExecutor.GetOutput(), "\n")

	log.Debug("Printing", "k3sToken", mainStateDocument.K8sBootstrap.K3s.K3sToken)

	err = storage.Write(mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

// ConfigureControlPlane implements resources.DistroFactory.
func (k3s *K3s) ConfigureControlPlane(noOfCP int, storage resources.StorageFactory) error {
	k3s.mu.Lock()
	idx := noOfCP
	sshExecutor := helpers.NewSSHExecutor(mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	k3s.mu.Unlock()

	log.Print("configuring ControlPlane", "number", strconv.Itoa(idx))
	if idx == 0 {
		err := configureCP_1(storage, k3s, sshExecutor)
		if err != nil {
			return log.NewError(err.Error())
		}
	} else {

		var script string

		if consts.KsctlValidCNIPlugin(k3s.Cni) == consts.CNINone {
			script = scriptCP_NWithoutCNI(
				mainStateDocument.K8sBootstrap.B.CACert,
				mainStateDocument.K8sBootstrap.B.EtcdCert,
				mainStateDocument.K8sBootstrap.B.EtcdKey,
				k3s.K3sVer,
				mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores,
				mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer,
				mainStateDocument.K8sBootstrap.K3s.K3sToken)
		} else {
			script = scriptCP_N(
				mainStateDocument.K8sBootstrap.B.CACert,
				mainStateDocument.K8sBootstrap.B.EtcdCert,
				mainStateDocument.K8sBootstrap.B.EtcdKey,
				k3s.K3sVer,
				mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores,
				mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer,
				mainStateDocument.K8sBootstrap.K3s.K3sToken)
		}

		err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(script).
			IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[idx]).
			FastMode(true).SSHExecute(log)
		if err != nil {
			return log.NewError(err.Error())
		}

		// err = storage.Write(mainStateDocument)
		// if err != nil {
		// 	return log.NewError(err.Error())
		// }

		if idx+1 == len(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes) {

			log.Debug("fetching kubeconfig")
			err = sshExecutor.Flag(consts.UtilExecWithOutput).Script(scriptKUBECONFIG()).
				IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
				FastMode(true).SSHExecute(log)
			if err != nil {
				return log.NewError(err.Error())
			}
			// as only a single case where it is getting invoked we actually don't need locks

			kubeconfig := sshExecutor.GetOutput()
			kubeconfig = strings.Replace(kubeconfig, "127.0.0.1", mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer, 1)
			kubeconfig = strings.Replace(kubeconfig, "default", mainStateDocument.ClusterName+"-"+mainStateDocument.Region+"-"+string(mainStateDocument.ClusterType)+"-"+string(mainStateDocument.InfraProvider)+"-ksctl", -1)
			// k3s.mu.Lock()
			// k3s.mu.Unlock()
			mainStateDocument.ClusterKubeConfig = kubeconfig
			log.Debug("Printing", "kubeconfig", kubeconfig)
			// modify
			err = storage.Write(mainStateDocument)
			if err != nil {
				return log.NewError(err.Error())
			}
		}

	}
	log.Success("configured ControlPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptCP_1WithoutCNI(ca, etcd, key, ver string, privateEtcdIps []string, pubIPlb string) string {
	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)

	return fmt.Sprintf(`#!/bin/bash
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


cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, ca, etcd, key, ver, dbEndpoint, pubIPlb)
}

func scriptCP_NWithoutCNI(ca, etcd, key, ver string, privateEtcdIps []string, pubIPlb, token string) string {
	//INSTALL_K3S_CHANNEL="v1.24.6+k3s1"   missing the usage of k3s version for ha
	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)
	return fmt.Sprintf(`#!/bin/bash
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


cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--token %s \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ca, etcd, key, ver, token, dbEndpoint, pubIPlb)
}

// scriptCP_1 script used to configure the control-plane-1 with no need of output inital
func scriptCP_1(ca, etcd, key, ver string, privateEtcdIps []string, pubIPlb string) string {
	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)
	return fmt.Sprintf(`#!/bin/bash
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


cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, ca, etcd, key, ver, dbEndpoint, pubIPlb)
}

func scriptForK3sToken() string {
	return `#!/bin/bash
sudo cat /var/lib/rancher/k3s/server/token
`
}

func scriptCP_N(ca, etcd, key, ver string, privateEtcdIps []string, pubIPlb, token string) string {
	//INSTALL_K3S_CHANNEL="v1.24.6+k3s1"   missing the usage of k3s version for ha
	dbEndpoint := getEtcdMemberIPFieldForControlplane(privateEtcdIps)
	return fmt.Sprintf(`#!/bin/bash
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


cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--token %s \
	--datastore-endpoint "%s" \
	--datastore-cafile=/var/lib/etcd/ca.pem \
	--datastore-keyfile=/var/lib/etcd/etcd-key.pem \
	--datastore-certfile=/var/lib/etcd/etcd.pem \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ca, etcd, key, ver, token, dbEndpoint, pubIPlb)
}
