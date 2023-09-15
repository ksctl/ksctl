package k3s

import (
	"fmt"
	"strings"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func configureCP_1(storage resources.StorageFactory, k3s *K3sDistro) error {

	err := k3s.SSHInfo.Flag(utils.EXEC_WITHOUT_OUTPUT).Script(
		scriptWithoutCP_1(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.ControlPlanes[0])).
		IPv4(k8sState.PublicIPs.ControlPlanes[0]).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return err
	}

	// K3stoken
	err = k3s.SSHInfo.Flag(utils.EXEC_WITH_OUTPUT).Script(scriptForK3sToken()).
		IPv4(k8sState.PublicIPs.ControlPlanes[0]).
		SSHExecute(storage)
	if err != nil {
		return err
	}

	storage.Logger().Note("[k3s] fetching k3s token")
	k8sState.K3sToken = strings.Trim(k3s.SSHInfo.GetOutput(), "\n")

	path := utils.GetPath(utils.CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err = saveStateHelper(storage, path)
	if err != nil {
		return err
	}
	return nil
}

// ConfigureControlPlane implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureControlPlane(noOfCP int, storage resources.StorageFactory) error {

	if noOfCP == 0 {
		err := configureCP_1(storage, k3s)
		if err != nil {
			return err
		}
	} else {

		err := k3s.SSHInfo.Flag(utils.EXEC_WITHOUT_OUTPUT).Script(
			scriptCP_N(k3s.K3sVer, k8sState.DataStoreEndPoint, k8sState.PublicIPs.Loadbalancer, k8sState.K3sToken)).
			IPv4(k8sState.PublicIPs.ControlPlanes[noOfCP]).
			FastMode(true).SSHExecute(storage)
		if err != nil {
			return err
		}
		path := utils.GetPath(utils.CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
		err = saveStateHelper(storage, path)
		if err != nil {
			return err
		}

		if noOfCP+1 == len(k8sState.PublicIPs.ControlPlanes) {

			storage.Logger().Note("[k3s] fetching kubeconfig")
			err = k3s.SSHInfo.Flag(utils.EXEC_WITH_OUTPUT).Script(scriptKUBECONFIG()).
				IPv4(k8sState.PublicIPs.ControlPlanes[0]).
				FastMode(true).SSHExecute(storage)
			if err != nil {
				return err
			}

			kubeconfig := k3s.SSHInfo.GetOutput()
			kubeconfig = strings.Replace(kubeconfig, "127.0.0.1", k8sState.PublicIPs.Loadbalancer, 1)
			kubeconfig = strings.Replace(kubeconfig, "default", k8sState.ClusterName+"-"+k8sState.Region+"-"+k8sState.ClusterType+"-"+k8sState.Provider+"-ksctl", -1)

			// modify
			path = utils.GetPath(utils.CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)
			err = saveKubeconfigHelper(storage, path, kubeconfig)
			if err != nil {
				return err
			}
			printKubeconfig(storage, utils.OPERATION_STATE_CREATE)
		}

	}
	storage.Logger().Success("[k3s] configured ControlPlane", string(rune(noOfCP)))

	return nil
}

// scriptWithoutCP_1 script used to configure the control-plane-1 with no need of output inital
func scriptWithoutCP_1(ver string, dbEndpoint, pubIPlb string) string {

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

	// NOTE: Feature to add other CNI like Cilium
	// Add it when the installApplication() feature is added so that it could install cni plugins

	// Add these tags for having different CNI
	// also check out the default loadbalancer available

	//	return fmt.Sprintf(`#!/bin/bash
	// export K3S_DATASTORE_ENDPOINT='%s'
	//	curl -sfL https://get.k3s.io | sh -s - server \
	//		--flannel-backend=none \
	//		--disable-network-policy \
	//		--node-taint CriticalAddonsOnly=true:NoExecute \
	//		--tls-san %s
	//
	// `, dbEndpoint, privateIPlb)
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

	// NOTE: Feature to add other CNI like Cilium
	// Add these tags for having different CNI
	// also check out the default loadbalancer available

	//	return fmt.Sprintf(`#!/bin/bash
	// export SECRET='%s'
	// export K3S_DATASTORE_ENDPOINT='%s'
	//
	//	curl -sfL https://get.k3s.io | sh -s - server \
	//		--token=$SECRET \
	//		--node-taint CriticalAddonsOnly=true:NoExecute \
	//		--flannel-backend=none \
	//		--disable-network-policy \
	//		--tls-san %s
	//
	// `, token, dbEndpoint, privateIPlb)
}

func (k3s *K3sDistro) GetKubeConfig(storage resources.StorageFactory) (path string, data string, err error) {

	if len(k8sState.Provider) == 0 || len(k8sState.ClusterDir) == 0 || len(k8sState.ClusterDir) == 0 {
		return "", "", fmt.Errorf("[k3s] status is not correct")
	}

	path = utils.GetPath(utils.CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)

	var raw []byte
	raw, err = storage.Path(path).Load()

	data = string(raw)

	return
}
