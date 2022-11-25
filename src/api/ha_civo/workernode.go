package ha_civo

import (
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
)

func scriptWP(privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
curl -sfL https://get.k3s.io | sh -s - agent --token=$SECRET --server https://%s:6443
`, token, privateIPlb)
}

func CreateWorkerNode(client *civogo.Client, number int, clusterName, privateIPlb, token, nodeSize string) (*civogo.Instance, error) {

	network, err := GetNetwork(client, clusterName+"-ksctl")
	if err != nil {
		return nil, err
	}

	diskImg, err := client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s-ksctl-wp-%d", clusterName, number)

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
			err := ExecWithoutOutput(getInstance.PublicIP, getInstance.InitialPassword, scriptWP(privateIPlb, token), false)
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
