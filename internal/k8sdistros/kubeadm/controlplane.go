package kubeadm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

func configureCP_1(storage resources.StorageFactory, kubeadm *Kubeadm) error {

	installKubeadmTools := fmt.Sprintf("%s\n%s", scriptInstallKubeadmAndOtherTools(kubeadm.KubeadmVer), scriptTransferEtcdCerts(
		mainStateDocument.K8sBootstrap.B.CACert,
		mainStateDocument.K8sBootstrap.B.EtcdCert,
		mainStateDocument.K8sBootstrap.B.EtcdKey))

	log.Print("Installing Kubeadm and copying etcd certificates")
	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(installKubeadmTools).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		FastMode(true).
		SSHExecute(log); err != nil {
		return log.NewError(err.Error())
	}

	log.Print("Fetching Kubeadm Bootstrap Certificate Key")

	if err := sshExecutor.Flag(consts.UtilExecWithOutput).
		Script(scriptGetCertificateKey()).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute(log); err != nil {
		return log.NewError(err.Error())
	}

	mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey = strings.Trim(sshExecutor.GetOutput(), "\n")

	log.Debug("Printing", "CertificateKey", mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey)

	log.Print("Generating Kubeadm Bootstrap Token")

	if v, err := generatebootstrapToken(); err != nil {
		return log.NewError(err.Error())
	} else {
		mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken = v
		log.Debug("Printing", "BootstrapToken", mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken)
	}

	log.Print("Configuring K8s cluster")

	configureControlPlane0 := scriptAddKubeadmControlplane0(
		kubeadm.KubeadmVer,
		mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
		mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey,
		mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer,
		mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores)

	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(configureControlPlane0).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		FastMode(true).
		SSHExecute(log); err != nil {
		return log.NewError(err.Error())
	}

	log.Print("Fetching Discovery Token CA Cert Hash")

	if err := sshExecutor.Flag(consts.UtilExecWithOutput).
		Script(scriptDiscoveryTokenCACertHash()).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute(log); err != nil {
		return log.NewError(err.Error())
	}

	mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash = strings.Trim(sshExecutor.GetOutput(), "\n")

	log.Debug("Printing", "DiscoveryTokenCACertHash", mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash)

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (p *Kubeadm) ConfigureControlPlane(noOfCP int, storage resources.StorageFactory) error {
	idx := noOfCP
	log.Print("configuring ControlPlane", "number", strconv.Itoa(idx))
	if idx == 0 {
		err := configureCP_1(storage, p)
		if err != nil {
			return log.NewError(err.Error())
		}
	} else {

		installKubeadmTools := fmt.Sprintf(`%s\n%s`, scriptInstallKubeadmAndOtherTools(p.KubeadmVer), scriptTransferEtcdCerts(
			mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey))

		log.Print("Installing Kubeadm and copying etcd certificates")

		if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
			Script(installKubeadmTools).
			IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[idx]).
			FastMode(true).
			SSHExecute(log); err != nil {
			return log.NewError(err.Error())
		}

		log.Print("Joining controlplane to existing cluster")
		if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
			Script(scriptJoinControlplane(
				mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer,
				mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
				mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash,
				mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey,
			)).
			IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[idx]).
			FastMode(true).
			SSHExecute(log); err != nil {
			return log.NewError(err.Error())
		}

		if idx+1 == len(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes) {

			log.Print("Fetching Kubeconfig")

			if err := sshExecutor.Flag(consts.UtilExecWithOutput).
				Script(scriptGetKubeconfig()).
				IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
				FastMode(true).
				SSHExecute(log); err != nil {
				return log.NewError(err.Error())
			}

			kubeconfig := sshExecutor.GetOutput()
			kubeconfig = strings.Replace(kubeconfig, "kubernetes-admin@kubernetes", mainStateDocument.ClusterName+"-"+mainStateDocument.Region+"-"+string(mainStateDocument.ClusterType)+"-"+string(mainStateDocument.InfraProvider)+"-ksctl", -1)
			mainStateDocument.ClusterKubeConfig = kubeconfig
			log.Debug("Printing", "kubeconfig", kubeconfig)

			if err := storage.Write(mainStateDocument); err != nil {
				return log.NewError(err.Error())
			}
		}

	}
	log.Success("configured ControlPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptAddKubeadmControlplane0(ver string, bootstrapToken, certificateKey, publicIPLb string, privateIPDs []string) string {

	var generateExternalEtcdConfig func([]string) string = func(ips []string) string {
		var ret strings.Builder
		for _, ip := range ips {
			ret.WriteString(fmt.Sprintf("    - https://%s:2379\n", ip))
		}
		return ret.String()
	}

	etcdConf := generateExternalEtcdConfig(privateIPDs)

	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > kubeadm-config.yml
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: %s
  ttl: 24h0m0s
  usages:
  - signing
  - authentication

certificateKey: %s
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
  imagePullPolicy: IfNotPresent
  taints: null
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs:
    - "%s"
    - "127.0.0.1"
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns: {}
etcd:
  external:
    endpoints:
%s
    caFile: "/etcd/kubernetes/pki/etcd/ca.pem"
    certFile: "/etcd/kubernetes/pki/etcd/etcd.pem"
    keyFile: "/etcd/kubernetes/pki/etcd/etcd-key.pem"
imageRepository: registry.k8s.io
kubernetesVersion: %s.0
controlPlaneEndpoint: "%s:6443"
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
scheduler: {}
EOF

sudo kubeadm init --config kubeadm-config.yml --upload-certs
`, bootstrapToken, certificateKey, publicIPLb, etcdConf, ver, publicIPLb)
}

func scriptGetKubeconfig() string {
	return `#!/bin/bash
sudo cat /etc/kubernetes/admin.conf
`
}

func scriptDiscoveryTokenCACertHash() string {
	return `#!/bin/bash
sudo openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
`
}

func scriptGetCertificateKey() string {
	return `#!/bin/bash
sudo kubeadm certs certificate-key
`
}

func generatebootstrapToken() (string, error) {
	//form "\\A([a-z0-9]{6})\\.([a-z0-9]{16})\\z"
	prefix, err := helpers.GenRandomString(6)
	if err != nil {
		return "", err
	}

	postfix, err := helpers.GenRandomString(16)
	if err != nil {
		return "", err
	}

	prefix = strings.ToLower(prefix)
	postfix = strings.ToLower(postfix)
	return prefix + "." + postfix, nil
}

func scriptTransferEtcdCerts(ca, etcd, key string) string {
	return fmt.Sprintf(`
echo "This script needs to be used as in concatenation to install kubeadm tools"

sudo mkdir -vp /etcd/kubernetes/pki/etcd/

cat <<EOF > ca.pem
%s
EOF

cat <<EOF > etcd.pem
%s
EOF

cat <<EOF > etcd-key.pem
%s
EOF

sudo mv -v ca.pem etcd.pem etcd-key.pem /etcd/kubernetes/pki/etcd
`, ca, etcd, key)
}

func scriptJoinControlplane(pubIPLb, token, cacertSHA, certKey string) string {
	return fmt.Sprintf(`#!/bin/bash
sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s --control-plane --certificate-key %s
`, pubIPLb, token, cacertSHA, certKey)
}
