package k3s

import (
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// ConfigureLoadbalancer implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureLoadbalancer(storage resources.StorageFactory) error {

	var controlPlaneIPs = make([]string, len(k8sState.PublicIPs.ControlPlanes))
	for i := 0; i < len(k8sState.PublicIPs.ControlPlanes); i++ {
		controlPlaneIPs[i] = k8sState.PrivateIPs.ControlPlanes[i] + ":6443"
	}

	err := k3s.SSHInfo.Flag(utils.EXEC_WITHOUT_OUTPUT).Script(
		configLBscript(controlPlaneIPs)).
		IPv4(k8sState.PublicIPs.Loadbalancer).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return fmt.Errorf("[k3s] loadbalancer%v", err)
	}

	storage.Logger().Success("[k3s] configured LoadBalancer")

	return nil
}

func configLBscript(controlPlaneIPs []string) string {
	script := `#!/bin/bash
apt update
apt install haproxy -y
systemctl start haproxy && systemctl enable haproxy

cat <<EOF > /etc/haproxy/haproxy.cfg
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

systemctl restart haproxy
`
	return script
}
