package kubeadm

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func (p *Kubeadm) JoinWorkerplane(noOfWP int, storage types.StorageFactory) error {
	p.mu.Lock()
	idx := noOfWP
	sshExecutor := helpers.NewSSHExecutor(mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	log.Print("configuring Workerplane", "number", strconv.Itoa(idx))

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	script := scriptJoinWorkerplane(
		scriptInstallKubeadmAndOtherTools(p.KubeadmVer),
		mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer,
		mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
		mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash,
	)
	log.Print("Installing Kubeadm and Joining WorkerNode to existing cluster")

	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(script).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).
		SSHExecute(log); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("configured WorkerPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptJoinWorkerplane(collection types.ScriptCollection, privateIPLb, token, cacertSHA string) types.ScriptCollection {

	collection.Append(types.Script{
		Name:           "Join K3s workerplane",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s &>> ksctl.log
`, privateIPLb, token, cacertSHA),
	})

	return collection
}
