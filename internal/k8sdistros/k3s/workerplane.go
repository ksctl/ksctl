package k3s

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// JoinWorkerplane implements resources.DistroFactory.
func (k3s *K3s) JoinWorkerplane(idx int, storage resources.StorageFactory) error {

	log.Print("configuring Workerplane", "number", strconv.Itoa(idx))

	err := storage.Write(mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		scriptWP(k3s.K3sVer, mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer, mainStateDocument.K8sBootstrap.K3s.K3sToken)).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).SSHExecute(log)
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
