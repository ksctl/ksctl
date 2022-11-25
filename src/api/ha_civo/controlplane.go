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
curl -sfL https://get.k3s.io | sh -s - server --node-taint CriticalAddonsOnly=true:NoExecute --tls-san %s
`, dbEndpoint, privateIPlb)
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

func CreateControlPlane(client *civogo.Client, number int, clusterName, nodeSize string) (*civogo.Instance, error) {

	network, err := GetNetwork(client, clusterName+"-ksctl")
	if err != nil {
		return nil, err
	}

	diskImg, err := client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s-ksctl-cp-%d", clusterName, number)

	firewall, err := CreateFirewall(client, name, network.ID)
	if err != nil {
		return nil, err
	}

	err = ConfigWriterFirewall(firewall, clusterName, client.Region)
	if err != nil {
		return nil, nil
	}

	instance, err := CreateInstance(client, name, firewall.ID, diskImg.ID, nodeSize, network.ID, "")
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
			log.Println("[ CREATED ] Instance " + name)
			return retObject, nil
		}
		log.Println(getInstance.Status)
		time.Sleep(10 * time.Second)
	}

}

func GetTokenFromCP_1(instance *civogo.Instance) string {
	token, err := ExecWithOutput(instance.PublicIP, instance.InitialPassword, scriptWithCP_1(), true)
	if err != nil {
		return ""
	}
	return token
}
