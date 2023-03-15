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
)

func scriptWP(privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
curl -sfL https://get.k3s.io | sh -s - agent --token=$SECRET --server https://%s:6443
`, token, privateIPlb)
}

func (obj *HAType) CreateWorkerNode(logging log.Logger, number int, privateIPlb, token string) (*civogo.Instance, error) {

	name := fmt.Sprintf("%s-ksctl-wp", obj.ClusterName)

	if len(obj.WPFirewallID) == 0 {
		firewall, err := obj.CreateFirewall(name)
		if err != nil {
			return nil, err
		}
		// TODO: More restrictive firewalls
		obj.WPFirewallID = firewall.ID
		err = obj.Configuration.ConfigWriterFirewallWorkerNodes(logging, firewall.ID)
		if err != nil {
			return nil, nil
		}
	}
	name += fmt.Sprint(number)

	instance, err := obj.CreateInstance(name, obj.WPFirewallID, obj.NodeSize, scriptWP(privateIPlb, token), false)
	if err != nil {
		return nil, err
	}

	err = obj.Configuration.ConfigWriterInstanceWorkerNodes(logging, instance.ID)
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
			logging.Info("💻 Booted Instance", name)
			return retObject, nil
		}
		logging.Info("🚧 Instance", name)
		time.Sleep(10 * time.Second)
	}

}
