package k8sdistros

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"

	"github.com/ksctl/ksctl/pkg/resources"
)

func (p *PreBootstrap) ConfigureLoadbalancer(_ resources.StorageFactory) error {
	log.Print("configuring Loadbalancer")
	p.mu.Lock()
	sshExecutor := helpers.NewSSHExecutor(mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	var controlPlaneIPs = make([]string, len(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes))

	for i := 0; i < len(mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes); i++ {
		controlPlaneIPs[i] = mainStateDocument.K8sBootstrap.B.PrivateIPs.ControlPlanes[i] + ":6443"
	}

	err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		configLBscript(controlPlaneIPs)).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("configured LoadBalancer")
	return nil
}

func configLBscript(controlPlaneIPs []string) string {
	script := `#!/bin/bash
sudo apt update
sudo apt install haproxy -y
sleep 2s
sudo systemctl start haproxy && sudo systemctl enable haproxy

cat <<EOF > haproxy.cfg
frontend kubernetes-frontend
  bind *:6443
  mode tcp
  option tcplog
  timeout client 10s
  default_backend kubernetes-backend

backend kubernetes-backend
  timeout connect 10s
  timeout server 10s
  mode tcp
  option tcp-check
  balance roundrobin
`

	for index, controlPlaneIP := range controlPlaneIPs {
		script += fmt.Sprintf(`  server k3sserver-%d %s check
`, index+1, controlPlaneIP)
	}

	script += `EOF

sudo mv haproxy.cfg /etc/haproxy/haproxy.cfg
sudo systemctl restart haproxy
`
	return script
}
