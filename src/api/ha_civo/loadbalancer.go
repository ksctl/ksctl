package ha_civo

import (
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
)

// NGINX LOADBALANCER (DEPRICATED)
// func scriptLB() string {
// 	return `#!/bin/bash
// apt update
// apt install nginx -y
// systemctl start nginx && systemctl enable nginx
// `
// }

// func configLBscript(controlPlaneIPs []string) string {
// 	script := `cat <<EOF > /etc/nginx/nginx.conf
// user www-data;
// worker_processes auto;
// pid /run/nginx.pid;
// include /etc/nginx/modules-enabled/*.conf;

// events {}
// stream {
//   upstream k3s_servers {
// `

// 	for _, controlPlaneIP := range controlPlaneIPs {
// 		script += fmt.Sprintf(`    server %s;
// `, controlPlaneIP)
// 	}

// 	script += `  }
//   server {
//     listen 6443;
//     proxy_pass k3s_servers;
//   }
// }

// EOF

// systemctl restart nginx
// `
// 	return script
// }

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
		log.Println("[CONFIGURED] LoadBalancer")
	}
	return err
}

func CreateLoadbalancer(client *civogo.Client, clusterName string) (*civogo.Instance, error) {

	var networkID string
	network, err := GetNetwork(client, clusterName+"-ksctl")
	if err != nil {
		return nil, err
	}
	networkID = network.ID

	diskImg, err := client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return nil, err
	}

	name := clusterName + "-ksctl-lb"
	firewall, err := CreateFirewall(client, name, networkID)
	if err != nil {
		return nil, err
	}

	err = ConfigWriterFirewall(firewall, clusterName, client.Region)
	if err != nil {
		return nil, nil
	}

	instance, err := CreateInstance(client, name, firewall.ID, diskImg.ID, "g3.medium", networkID, "")
	if err != nil {
		return nil, err
	}

	err = ConfigWriterInstance(instance, clusterName, client.Region)
	if err != nil {
		return nil, nil
	}

	var retObject *civogo.Instance

	for {
		getInstance, err := GetInstance(client, instance.ID)
		if err != nil {
			return nil, err
		}

		if getInstance.Status == "ACTIVE" {
			retObject = getInstance
			err = ExecWithoutOutput(getInstance.PublicIP, getInstance.InitialPassword, scriptLB(), false)
			if err != nil {
				return nil, err
			}

			log.Println("[ CREATED ] Instance " + name)
			return retObject, nil
		}
		log.Println(getInstance.Status)
		time.Sleep(10 * time.Second)
	}
}
