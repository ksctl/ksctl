package kubeadm

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

func (p *Kubeadm) JoinWorkerplane(noOfWP int, storage resources.StorageFactory) error {
	p.mu.Lock()
	idx := noOfWP
	sshExecutor := helpers.NewSSHExecutor(mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	log.Print("configuring Workerplane", "number", strconv.Itoa(idx))

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	script := fmt.Sprintf("%s\n%s",
		scriptInstallKubeadmAndOtherTools(p.KubeadmVer),
		scriptJoinWorkerplane(
			mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer,
			mainStateDocument.K8sBootstrap.Kubeadm.BootstrapToken,
			mainStateDocument.K8sBootstrap.Kubeadm.DiscoveryTokenCACertHash,
		))
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

func scriptJoinWorkerplane(pubIPLb, token, cacertSHA string) string {

	return fmt.Sprintf(`
echo "To be used with kubeadm install"

sudo kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash sha256:%s
`, pubIPLb, token, cacertSHA)

}
