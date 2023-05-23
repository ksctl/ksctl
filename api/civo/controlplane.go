/*
Kubesimplify
Credit to @civo
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package civo

import (
	"fmt"
	"time"

	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/civo/civogo"
	util "github.com/kubesimplify/ksctl/api/utils"
)

// scriptWithoutCP_1 script used to configure the control-plane-1 with no need of output inital
// params
//
//	dbEndpoint: database-Endpoint
//	privateIPlb: private IP of loadbalancer
func scriptWithoutCP_1(dbEndpoint, privateIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
export K3S_DATASTORE_ENDPOINT='%s'
export INSTALL_K3S_EXEC='--tls-san %s 0.0.0.0/0'
curl -sfL https://get.k3s.io | sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute
`, dbEndpoint, privateIPlb)

	// NOTE: Feature to add other CNI like Cilium
	// Add these tags for having different CNI
	// also check out the default loadbalancer available

	//	return fmt.Sprintf(`#!/bin/bash
	// export K3S_DATASTORE_ENDPOINT='%s'
	//	curl -sfL https://get.k3s.io | sh -s - server \
	//		--flannel-backend=none \
	//		--disable-network-policy \
	//		--node-taint CriticalAddonsOnly=true:NoExecute \
	//		--tls-san %s
	//
	// `, dbEndpoint, privateIPlb)
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
export INSTALL_K3S_EXEC='--tls-san %s 0.0.0.0/0'
curl -sfL https://get.k3s.io | sh -s - server \
	--token=$SECRET \
	--node-taint CriticalAddonsOnly=true:NoExecute
`, token, dbEndpoint, privateIPlb)

	// NOTE: Feature to add other CNI like Cilium
	// Add these tags for having different CNI
	// also check out the default loadbalancer available

	//	return fmt.Sprintf(`#!/bin/bash
	// export SECRET='%s'
	// export K3S_DATASTORE_ENDPOINT='%s'
	//
	//	curl -sfL https://get.k3s.io | sh -s - server \
	//		--token=$SECRET \
	//		--node-taint CriticalAddonsOnly=true:NoExecute \
	//		--flannel-backend=none \
	//		--disable-network-policy \
	//		--tls-san %s
	//
	// `, token, dbEndpoint, privateIPlb)
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
cat /etc/rancher/k3s/k3s.yaml`
}

func (obj *HAType) FetchKUBECONFIG(logging log.Logger, instanceCP *civogo.Instance) (string, error) {
	obj.SSH_Payload.PublicIP = instanceCP.PublicIP
	obj.SSH_Payload.Output = ""
	err := obj.SSH_Payload.SSHExecute(logging, util.EXEC_WITH_OUTPUT, scriptKUBECONFIG(), true)

	if err != nil {
		return "", nil
	}

	return obj.SSH_Payload.Output, nil
}

func (obj *HAType) CreateControlPlane(logging log.Logger, number int) (*civogo.Instance, error) {

	name := fmt.Sprintf("%s-ksctl-cp", obj.ClusterName)

	if len(obj.CPFirewallID) == 0 {
		firewall, err := obj.CreateFirewall(name)
		if err != nil {
			return nil, err
		}
		obj.CPFirewallID = firewall.ID
		// TODO: Add better firewall rules
		err = obj.Configuration.ConfigWriterFirewallControlPlaneNodes(logging, firewall.ID)
		if err != nil {
			return nil, nil
		}
	}
	name += fmt.Sprint(number)

	instance, err := obj.CreateInstance(name, obj.CPFirewallID, obj.NodeSize, "", true)
	if err != nil {
		return nil, err
	}

	err = obj.Configuration.ConfigWriterInstanceControlPlaneNodes(logging, instance.ID)
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
			logging.Info("💻 Booted Instance ", name)
			return retObject, nil
		}
		logging.Info("🚧 Instance ", name)
		time.Sleep(10 * time.Second)
	}

}

// GetTokenFromCP_1 used to extract the K3S_TOKEN from the first Controlplane node
func (obj *HAType) GetTokenFromCP_1(logging log.Logger, instance *civogo.Instance) string {
	obj.SSH_Payload.PublicIP = instance.PublicIP
	obj.SSH_Payload.Output = ""
	err := obj.SSH_Payload.SSHExecute(logging, util.EXEC_WITH_OUTPUT, scriptWithCP_1(), true)
	if err != nil {
		return ""
	}
	token := obj.SSH_Payload.Output
	obj.SSH_Payload.Output = ""
	err = obj.Configuration.ConfigWriterServerToken(logging, token)

	if err != nil {
		return ""
	}
	return token
}
