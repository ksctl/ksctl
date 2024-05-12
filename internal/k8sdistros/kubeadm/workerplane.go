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
	sshExecutor := helpers.NewSSHExecutor(kubeadmCtx, log, mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	log.Print(kubeadmCtx, "configuring Workerplane", "number", strconv.Itoa(idx))

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	script := scriptJoinWorkerplane(
		scriptInstallKubeadmAndOtherTools(p.KubeadmVer),
		mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer,
		mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
		mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash,
	)
	log.Print(kubeadmCtx, "Installing Kubeadm and Joining WorkerNode to existing cluster")

	if err := sshExecutor.Flag(consts.UtilExecWithoutOutput).
		Script(script).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).
		SSHExecute(); err != nil {
		return err
	}

	log.Success(kubeadmCtx, "configured WorkerPlane", "number", strconv.Itoa(idx))

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
