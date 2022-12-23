package ha_civo

import (
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
)

// HAPROXY LOADBALANCER
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
		script += fmt.Sprintf(`  server k3sserver-%d %s
`, index+1, controlPlaneIP)
	}

	script += `EOF

systemctl restart haproxy
`
	return script
}

func ConfigLoadBalancer(instance *civogo.Instance, CPIPs []string) error {
	getScript := configLBscript(CPIPs)
	err := ExecWithoutOutput(instance.PublicIP, instance.InitialPassword, getScript, true)
	if err == nil {
		log.Println("✅ 🔧 LoadBalancer")
	}
	return err
}

func (obj *HAType) CreateLoadbalancer() (*civogo.Instance, error) {

	name := obj.ClusterName + "-ksctl-lb"
	firewall, err := obj.CreateFirewall(name)
	if err != nil {
		return nil, err
	}
	obj.LBFirewallID = firewall.ID

	err = obj.Configuration.ConfigWriterFirewallLoadBalancerNodes(firewall.ID)
	if err != nil {
		return nil, nil
	}

	instance, err := obj.CreateInstance(name, firewall.ID, "g3.medium", scriptLB(), true)
	if err != nil {
		return nil, err
	}

	err = obj.Configuration.ConfigWriterInstanceLoadBalancer(instance.ID)
	if err != nil {
		return nil, nil
	}

	var retObject *civogo.Instance

	for {
		getInstance, err := obj.GetInstance(instance.ID)
		if err != nil {
			return nil, err
		}

		if getInstance.Status == "ACTIVE" {
			retObject = getInstance

			log.Println("✅ 🚀 Instance " + name)
			return retObject, nil
		}
		log.Println("🚧 Instance " + name)
		time.Sleep(10 * time.Second)
	}
}
