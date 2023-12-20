package k3s

import (
	"fmt"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

// ConfigureLoadbalancer implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureLoadbalancer(storage resources.StorageFactory) error {

	log.Print("configuring Loadbalancer")

	var controlPlaneIPs = make([]string, len(k8sState.PublicIPs.ControlPlanes))
	for i := 0; i < len(k8sState.PublicIPs.ControlPlanes); i++ {
		controlPlaneIPs[i] = k8sState.PrivateIPs.ControlPlanes[i] + ":6443"
	}

	err := k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(
		configLBscript(controlPlaneIPs)).
		IPv4(k8sState.PublicIPs.Loadbalancer).
		FastMode(true).SSHExecute(storage, log, k8sState.Provider)
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
