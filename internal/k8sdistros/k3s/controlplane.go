package k3s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func configureCP_1(storage resources.StorageFactory, k3s *K3sDistro) error {

	var script string

	if consts.KsctlValidCNIPlugin(k3s.Cni) == consts.CNINone {
		script = scriptCP_1WithoutCNI(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer)
	} else {
		script = scriptCP_1(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer)
	}

	err := k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(script).
		IPv4(k8sState.PublicIPs.ControlPlanes[0]).
		FastMode(true).SSHExecute(storage, log, k8sState.Provider)
	if err != nil {
		return log.NewError(err.Error())
	}

	// K3stoken
	err = k3s.SSHInfo.Flag(consts.UtilExecWithOutput).Script(scriptForK3sToken()).
		IPv4(k8sState.PublicIPs.ControlPlanes[0]).
		SSHExecute(storage, log, k8sState.Provider)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Debug("fetching k3s token")

	k8sState.K3sToken = strings.Trim(k3s.SSHInfo.GetOutput(), "\n")

	log.Debug("Printing", "k3sToken", k8sState.K3sToken)

	path := helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err = saveStateHelper(storage, path)
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

// ConfigureControlPlane implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureControlPlane(noOfCP int, storage resources.StorageFactory) error {
	log.Print("configuring ControlPlane", "number", strconv.Itoa(noOfCP))
	if noOfCP == 0 {
		err := configureCP_1(storage, k3s)
		if err != nil {
			return log.NewError(err.Error())
		}
	} else {

		var script string

		if consts.KsctlValidCNIPlugin(k3s.Cni) == consts.CNINone {
			script = scriptCP_NWithoutCNI(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer, k8sState.K3sToken)
		} else {
			script = scriptCP_N(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer, k8sState.K3sToken)
		}

		err := k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(script).
			IPv4(k8sState.PublicIPs.ControlPlanes[noOfCP]).
			FastMode(true).SSHExecute(storage, log, k8sState.Provider)
		if err != nil {
			return log.NewError(err.Error())
		}
		path := helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
		err = saveStateHelper(storage, path)
		if err != nil {
			return log.NewError(err.Error())
		}

		if noOfCP+1 == len(k8sState.PublicIPs.ControlPlanes) {

			log.Debug("fetching kubeconfig")
			err = k3s.SSHInfo.Flag(consts.UtilExecWithOutput).Script(scriptKUBECONFIG()).
				IPv4(k8sState.PublicIPs.ControlPlanes[0]).
				FastMode(true).SSHExecute(storage, log, k8sState.Provider)
			if err != nil {
				return log.NewError(err.Error())
			}

			kubeconfig := k3s.SSHInfo.GetOutput()
			kubeconfig = strings.Replace(kubeconfig, "127.0.0.1", k8sState.PublicIPs.Loadbalancer, 1)
			kubeconfig = strings.Replace(kubeconfig, "default", k8sState.ClusterName+"-"+k8sState.Region+"-"+string(k8sState.ClusterType)+"-"+string(k8sState.Provider)+"-ksctl", -1)

			log.Debug("Printing", "kubeconfig", kubeconfig)
			// modify
			path = helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)
			err = saveKubeconfigHelper(storage, path, kubeconfig)
			if err != nil {
				return log.NewError(err.Error())
			}
			printKubeconfig(storage, consts.OperationStateCreate)
		}

	}
	log.Success("configured ControlPlane", "number", strconv.Itoa(noOfCP))

	return nil
}

func scriptCP_1WithoutCNI(ver string, dbEndpoint, pubIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, ver, dbEndpoint, pubIPlb)
}

func scriptCP_NWithoutCNI(ver string, dbEndpoint, pubIPlb, token string) string {
	//INSTALL_K3S_CHANNEL="v1.24.6+k3s1"   missing the usage of k3s version for ha
	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--token %s \
	--datastore-endpoint="%s" \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--flannel-backend=none \
	--disable-network-policy \
	--tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ver, token, dbEndpoint, pubIPlb)
}

// scriptCP_1 script used to configure the control-plane-1 with no need of output inital
func scriptCP_1(ver string, dbEndpoint, pubIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--datastore-endpoint "%s" \
	--tls-san %s
EOF

sudo chmod +x control-setup.sh
sudo ./control-setup.sh
`, ver, dbEndpoint, pubIPlb)
}

func scriptForK3sToken() string {
	return `#!/bin/bash
sudo cat /var/lib/rancher/k3s/server/token
`
}

func scriptCP_N(ver string, dbEndpoint, pubIPlb, token string) string {
	//INSTALL_K3S_CHANNEL="v1.24.6+k3s1"   missing the usage of k3s version for ha
	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > control-setupN.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - server --token %s --datastore-endpoint="%s" --node-taint CriticalAddonsOnly=true:NoExecute --tls-san %s
EOF

sudo chmod +x control-setupN.sh
sudo ./control-setupN.sh
`, ver, token, dbEndpoint, pubIPlb)
}

func (k3s *K3sDistro) GetKubeConfig(storage resources.StorageFactory) (path string, data string, err error) {

	if len(k8sState.Provider) == 0 || len(k8sState.ClusterDir) == 0 || len(k8sState.ClusterDir) == 0 {
		return "", "", log.NewError("status is not correct")
	}

	path = helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)

	var raw []byte
	raw, err = storage.Path(path).Load()

	data = string(raw)

	return
}
