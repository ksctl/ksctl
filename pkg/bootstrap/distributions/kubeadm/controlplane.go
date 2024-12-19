package kubeadm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func configureCP_1(storage types.StorageFactory, kubeadm *Kubeadm, sshExecutor helpers.SSHCollection) error {

	installKubeadmTools := scriptTransferEtcdCerts(
		scriptInstallKubeadmAndOtherTools(mainStateDocument.K8sBootstrap.Kubeadm.KubeadmVersion),
		mainStateDocument.K8sBootstrap.B.CACert,
		mainStateDocument.K8sBootstrap.B.EtcdCert,
		mainStateDocument.K8sBootstrap.B.EtcdKey)

	log.Print(kubeadmCtx, "Installing Kubeadm and copying etcd certificates")
	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(installKubeadmTools).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		FastMode(true).
		SSHExecute(); err != nil {
		return err
	}

	log.Print(kubeadmCtx, "Fetching Kubeadm Bootstrap Certificate Key")

	if err := sshExecutor.Flag(consts.UtilExecWithOutput).
		Script(scriptGetCertificateKey()).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute(); err != nil {
		return err
	}

	mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey = strings.Trim(sshExecutor.GetOutput()[0], "\n")

	log.Debug(kubeadmCtx, "Printing", "CertificateKey", mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey)

	log.Print(kubeadmCtx, "Generating Kubeadm Bootstrap Token")

	if err := sshExecutor.Flag(consts.UtilExecWithOutput).
		Script(scriptToGenerateBootStrapToken()).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute(); err != nil {
		return err
	}
	mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken = strings.Trim(sshExecutor.GetOutput()[0], "\n")
	mainStateDocument.K8sBootstrap.Kubeadm.BootstrapTokenExpireTimeUtc = time.Now().UTC().Add(20 * time.Minute)

	log.Print(kubeadmCtx, "Configuring K8s cluster")

	configureControlPlane0 := scriptAddKubeadmControlplane0(
		mainStateDocument.K8sBootstrap.Kubeadm.KubeadmVersion,
		mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
		mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey,
		mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer,
		mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer,
		mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores)

	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(configureControlPlane0).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		FastMode(true).
		SSHExecute(); err != nil {
		return err
	}

	log.Print(kubeadmCtx, "Fetching Discovery Token CA Cert Hash")

	if err := sshExecutor.Flag(consts.UtilExecWithOutput).
		Script(scriptDiscoveryTokenCACertHash()).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
		SSHExecute(); err != nil {
		return err
	}

	mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash = strings.Trim(sshExecutor.GetOutput()[0], "\n")

	log.Debug(kubeadmCtx, "Printing", "DiscoveryTokenCACertHash", mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash)

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	return nil
}

func (p *Kubeadm) ConfigureControlPlane(noOfCP int, storage types.StorageFactory) error {
	p.mu.Lock()
	idx := noOfCP
	sshExecutor := helpers.NewSSHExecutor(kubeadmCtx, log, mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	log.Note(kubeadmCtx, "configuring ControlPlane", "number", strconv.Itoa(idx))
	if idx == 0 {
		err := configureCP_1(storage, p, sshExecutor)
		if err != nil {
			return err
		}
	} else {

		installKubeadmTools := scriptTransferEtcdCerts(
			scriptInstallKubeadmAndOtherTools(mainStateDocument.K8sBootstrap.Kubeadm.KubeadmVersion),
			mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey)

		log.Print(kubeadmCtx, "Installing Kubeadm and copying etcd certificates")

		if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
			Script(installKubeadmTools).
			IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[idx]).
			FastMode(true).
			SSHExecute(); err != nil {
			return err
		}

		log.Print(kubeadmCtx, "Joining controlplane to existing cluster")
		if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
			Script(scriptJoinControlplane(
				mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer,
				mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
				mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash,
				mainStateDocument.K8sBootstrap.Kubeadm.CertificateKey,
			)).
			IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[idx]).
			FastMode(true).
			SSHExecute(); err != nil {
			return err
		}

		if idx+1 == len(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes) {

			log.Print(kubeadmCtx, "Fetching Kubeconfig")

			if err := sshExecutor.Flag(consts.UtilExecWithOutput).
				Script(scriptGetKubeconfig()).
				IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes[0]).
				FastMode(true).
				SSHExecute(); err != nil {
				return err
			}

			kubeconfig := sshExecutor.GetOutput()[0]
			contextName := mainStateDocument.ClusterName + "-" + mainStateDocument.Region + "-" + string(mainStateDocument.ClusterType) + "-" + string(mainStateDocument.InfraProvider) + "-ksctl"

			kubeconfig = strings.Replace(kubeconfig, "kubernetes-admin@kubernetes", contextName, -1)

			kubeconfig = strings.Replace(kubeconfig, mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer, mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer, 1)

			mainStateDocument.ClusterKubeConfig = kubeconfig
			mainStateDocument.ClusterKubeConfigContext = contextName

			if err := storage.Write(mainStateDocument); err != nil {
				return err
			}
		}
	}
	log.Success(kubeadmCtx, "configured ControlPlane", "number", strconv.Itoa(idx))

	return nil
}

func generateExternalEtcdConfig(ips []string) string {
	var ret strings.Builder
	for _, ip := range ips {
		ret.WriteString(fmt.Sprintf("    - https://%s:2379\n", ip))
	}
	return ret.String()
}

func scriptToGenerateBootStrapToken() types.ScriptCollection {
	collection := helpers.NewScriptCollection()
	collection.Append(
		types.Script{
			Name:           "generate bootstrap token",
			CanRetry:       false,
			ScriptExecutor: consts.LinuxBash,
			ShellScript: `
kubeadm token generate
`,
		},
	)

	return collection
}

func scriptToRenewBootStrapToken() types.ScriptCollection {
	collection := helpers.NewScriptCollection()
	collection.Append(
		types.Script{
			Name:           "renew bootstrap token",
			CanRetry:       false,
			ScriptExecutor: consts.LinuxBash,
			ShellScript: `
kubeadm token create --ttl 20m --description "ksctl bootstrap token"
`,
		},
	)

	return collection
}

func scriptAddKubeadmControlplane0(ver string, bootstrapToken, certificateKey, publicIPLb string, privateIpLb string, privateIPDs []string) types.ScriptCollection {

	etcdConf := generateExternalEtcdConfig(privateIPDs)

	collection := helpers.NewScriptCollection()

	collection.Append(types.Script{
		Name:       "store configuration for Controlplane0",
		CanRetry:   true,
		MaxRetries: 3,
		ShellScript: fmt.Sprintf(`
cat <<EOF > kubeadm-config.yml
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: %s
  ttl: 20m
  description: "ksctl bootstrap token"
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
  podSubnet: 10.244.0.0/16
scheduler: {}
EOF

`, bootstrapToken, certificateKey, publicIPLb, privateIpLb, etcdConf, ver, publicIPLb),
	})

	collection.Append(types.Script{
		Name:       "kubeadm init",
		CanRetry:   true,
		MaxRetries: 3,
		ShellScript: `
sudo kubeadm init --config kubeadm-config.yml --upload-certs  &>> ksctl.log
#### Adding the below for the kubeconfig to be set so that otken renew can work
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
`,
	})

	return collection
}

func scriptGetKubeconfig() types.ScriptCollection {

	collection := helpers.NewScriptCollection()
	collection.Append(types.Script{
		Name:     "fetch kubeconfig",
		CanRetry: false,
		ShellScript: `
sudo cat /etc/kubernetes/admin.conf
`,
	})
	return collection
}

func scriptDiscoveryTokenCACertHash() types.ScriptCollection {
	collection := helpers.NewScriptCollection()
	collection.Append(types.Script{
		Name:     "fetch discovery token ca cert hash",
		CanRetry: false,
		ShellScript: `
sudo openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -pubkey | openssl rsa -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1
`,
	})
	return collection
}

func scriptGetCertificateKey() types.ScriptCollection {

	collection := helpers.NewScriptCollection()
	collection.Append(types.Script{
		Name:     "fetch bootstrap certificate key",
		CanRetry: false,
		ShellScript: `
sudo kubeadm certs certificate-key
`,
	})
	return collection
}

func scriptTransferEtcdCerts(collection types.ScriptCollection, ca, etcd, key string) types.ScriptCollection {
	collection.Append(types.Script{
		Name:           "save etcd certificate",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
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
`, ca, etcd, key),
	})

	return collection
}

func scriptJoinControlplane(privateIPLb, token, cacertSHA, certKey string) types.ScriptCollection {

	collection := helpers.NewScriptCollection()
	collection.Append(types.Script{
		Name:           "Join Controlplane [1..N]",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s --control-plane --certificate-key %s  &>> ksctl.log
`, privateIPLb, token, cacertSHA, certKey),
	})
	return collection
}
