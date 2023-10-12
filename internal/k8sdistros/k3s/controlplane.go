package k3s

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func configureCP_1(storage resources.StorageFactory, k3s *K3sDistro) error {

	var script string
	if k3s.Cni == "flannel" {
		script = scriptCP_1(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer)
	} else if k3s.Cni == "cilium" {
		script = scriptCP_1WithoutCNI(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer)
	} else {
		return errors.New("[k3s] unsupported cni")
	}

	err := k3s.SSHInfo.Flag(EXEC_WITHOUT_OUTPUT).Script(script).
		IPv4(k8sState.PublicIPs.ControlPlanes[0]).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return err
	}

	// K3stoken
	err = k3s.SSHInfo.Flag(EXEC_WITH_OUTPUT).Script(scriptForK3sToken()).
		IPv4(k8sState.PublicIPs.ControlPlanes[0]).
		SSHExecute(storage)
	if err != nil {
		return err
	}

	storage.Logger().Note("[k3s] fetching k3s token")
	k8sState.K3sToken = strings.Trim(k3s.SSHInfo.GetOutput(), "\n")

	path := utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err = saveStateHelper(storage, path)
	if err != nil {
		return err
	}
	return nil
}

// ConfigureControlPlane implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureControlPlane(noOfCP int, storage resources.StorageFactory) error {
	storage.Logger().Print("[k3s] configuring ControlPlane", strconv.Itoa(noOfCP))
	if noOfCP == 0 {
		err := configureCP_1(storage, k3s)
		if err != nil {
			return err
		}
	} else {

		var script string
		if k3s.Cni == "flannel" {
			script = scriptCP_N(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer, k8sState.K3sToken)
		} else if k3s.Cni == "cilium" {
			script = scriptCP_NWithoutCNI(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer, k8sState.K3sToken)
		} else {
			return errors.New("[k3s] unsupported cni")
		}

		err := k3s.SSHInfo.Flag(EXEC_WITHOUT_OUTPUT).Script(script).
			IPv4(k8sState.PublicIPs.ControlPlanes[noOfCP]).
			FastMode(true).SSHExecute(storage)
		if err != nil {
			return err
		}
		path := utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
		err = saveStateHelper(storage, path)
		if err != nil {
			return err
		}

		if noOfCP+1 == len(k8sState.PublicIPs.ControlPlanes) {

			storage.Logger().Note("[k3s] fetching kubeconfig")
			err = k3s.SSHInfo.Flag(EXEC_WITH_OUTPUT).Script(scriptKUBECONFIG()).
				IPv4(k8sState.PublicIPs.ControlPlanes[0]).
				FastMode(true).SSHExecute(storage)
			if err != nil {
				return err
			}

			kubeconfig := k3s.SSHInfo.GetOutput()
			kubeconfig = strings.Replace(kubeconfig, "127.0.0.1", k8sState.PublicIPs.Loadbalancer, 1)
			kubeconfig = strings.Replace(kubeconfig, "default", k8sState.ClusterName+"-"+k8sState.Region+"-"+string(k8sState.ClusterType)+"-"+string(k8sState.Provider)+"-ksctl", -1)

			// modify
			path = utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)
			err = saveKubeconfigHelper(storage, path, kubeconfig)
			if err != nil {
				return err
			}
			printKubeconfig(storage, OPERATION_STATE_CREATE)
		}

	}
	storage.Logger().Success("[k3s] configured ControlPlane", strconv.Itoa(noOfCP))

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
		return "", "", fmt.Errorf("[k3s] status is not correct")
	}

	path = utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)

	var raw []byte
	raw, err = storage.Path(path).Load()

	data = string(raw)

	return
}
