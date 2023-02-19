package azure

import "fmt"

func scriptLB() string {
	return `#!/bin/bash
apt update
apt install haproxy -y
systemctl start haproxy && systemctl enable haproxy
`
}

func configLBscript(controlPlaneIPs []string) string {
	script := `cat <<EOF > /etc/haproxy/haproxy.cfg
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
