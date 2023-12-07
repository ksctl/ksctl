package k3s

import (
	"fmt"
	"strconv"

	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

// JoinWorkerplane implements resources.DistroFactory.
func (k3s *K3sDistro) JoinWorkerplane(idx int, storage resources.StorageFactory) error {

	log.Print("configuring Workerplane", "number", strconv.Itoa(idx))

	path := helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err := saveStateHelper(storage, path)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(
		scriptWP(k3s.K3sVer, k8sState.PrivateIPs.Loadbalancer, k8sState.K3sToken)).
		IPv4(k8sState.PublicIPs.WorkerPlanes[idx]).
		//<<<<<<< HEAD
		FastMode(true).SSHExecute(storage, log, k8sState.Provider)
	//=======
	//		FastMode(true).SSHExecute(storage, log)
	//>>>>>>> upstream/main
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("configured WorkerPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptWP(ver string, privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
cat <<EOF > worker-setup.sh
#!/bin/bash
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - agent --token %s --server https://%s:6443
EOF

sudo chmod +x worker-setup.sh
sudo ./worker-setup.sh
`, ver, token, privateIPlb)
}
