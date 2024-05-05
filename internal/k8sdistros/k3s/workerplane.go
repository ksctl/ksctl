package k3s

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// JoinWorkerplane implements resources.DistroFactory.
func (k3s *K3s) JoinWorkerplane(no int, _ resources.StorageFactory) error {
	k3s.mu.Lock()
	idx := no
	sshExecutor := helpers.NewSSHExecutor(mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	k3s.mu.Unlock()

	log.Print("configuring Workerplane", "number", strconv.Itoa(idx))

	// err := storage.Write(mainStateDocument)
	// if err != nil {
	// 	return log.NewError(err.Error())
	// }

	err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		scriptWP(k3s.K3sVer, mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer, mainStateDocument.K8sBootstrap.K3s.K3sToken)).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("configured WorkerPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptWP(ver string, privateIPlb, token string) resources.ScriptCollection {

	collection := helpers.NewScriptCollection()

	collection.Append(resources.Script{
		Name:           "Join the workerplane-[0..M]",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
cat <<EOF > worker-setup.sh
#!/bin/bash
export K3S_DEBUG=true
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - agent --token %s --server https://%s:6443
EOF

sudo chmod +x worker-setup.sh
sudo ./worker-setup.sh &>> ksctl.log
`, ver, token, privateIPlb),
	})

	return collection
}
