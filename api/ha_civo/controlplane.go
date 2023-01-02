/*
Kubesimplify
Credit to @civo
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package ha_civo

import (
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
)

// scriptWithoutCP_1 script used to configure the control-plane-1 with no need of output inital
// params
//
//	dbEndpoint: database-Endpoint
//	privateIPlb: private IP of loadbalancer
func scriptWithoutCP_1(dbEndpoint, privateIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
export K3S_DATASTORE_ENDPOINT='%s'
curl -sfL https://get.k3s.io | sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--tls-san %s
`, dbEndpoint, privateIPlb)

	// Add these tags for having different CNI
	// 	--flannel-backend=none \
	// --disable-network-policy \
	// also check out the default loadbalancer available
}

func scriptWithCP_1() string {
	return `#!/bin/bash
cat /var/lib/rancher/k3s/server/token
`
}

func scriptCP_n(dbEndpoint, privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
export K3S_DATASTORE_ENDPOINT='%s'
curl -sfL https://get.k3s.io | sh -s - server --token=$SECRET --node-taint CriticalAddonsOnly=true:NoExecute --tls-san %s
`, token, dbEndpoint, privateIPlb)
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
cat /etc/rancher/k3s/k3s.yaml`
}

func FetchKUBECONFIG(instanceCP *civogo.Instance) (string, error) {
	kubeconfig, err := ExecWithOutput(instanceCP.PublicIP, instanceCP.InitialPassword, scriptKUBECONFIG(), true)
	if err != nil {
		return "", nil
	}
	return kubeconfig, nil
}

func (obj *HAType) CreateControlPlane(number int) (*civogo.Instance, error) {

	name := fmt.Sprintf("%s-ksctl-cp", obj.ClusterName)

	if len(obj.CPFirewallID) == 0 {
		firewall, err := obj.CreateFirewall(name)
		if err != nil {
			return nil, err
		}
		obj.CPFirewallID = firewall.ID
		err = obj.Configuration.ConfigWriterFirewallControlPlaneNodes(firewall.ID)
		if err != nil {
			return nil, nil
		}
	}
	name += fmt.Sprint(number)

	instance, err := obj.CreateInstance(name, obj.CPFirewallID, obj.NodeSize, "", true)
	if err != nil {
		return nil, err
	}

	err = obj.Configuration.ConfigWriterInstanceControlPlaneNodes(instance.ID)
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
			log.Println("💻 Booted Instance " + name)
			return retObject, nil
		}
		log.Println("🚧 Instance " + name)
		time.Sleep(10 * time.Second)
	}

}

func (obj *HAType) GetTokenFromCP_1(instance *civogo.Instance) string {
	token, err := ExecWithOutput(instance.PublicIP, instance.InitialPassword, scriptWithCP_1(), true)
	if err != nil {
		return ""
	}
	err = obj.Configuration.ConfigWriterServerToken(token)
	if err != nil {
		return ""
	}
	return token
}
