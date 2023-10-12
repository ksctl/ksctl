package k3s

import (
	"fmt"
	"strconv"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// JoinWorkerplane implements resources.DistroFactory.
func (k3s *K3sDistro) JoinWorkerplane(idx int, storage resources.StorageFactory) error {

	storage.Logger().Print("[k3s] configuring Workerplane", strconv.Itoa(idx))

	path := utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err := saveStateHelper(storage, path)
	if err != nil {
		return err
	}

	err = k3s.SSHInfo.Flag(EXEC_WITHOUT_OUTPUT).Script(
		scriptWP(k3s.K3sVer, k8sState.PrivateIPs.Loadbalancer, k8sState.K3sToken)).
		IPv4(k8sState.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return fmt.Errorf("[k3s] workerplane %v", err)
	}

	storage.Logger().Success("[k3s] configured WorkerPlane", strconv.Itoa(idx))

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
