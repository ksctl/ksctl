package civo

import (
	"fmt"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func (obj *CivoProvider) foundStateVM(storage resources.StorageFactory, idx int) error {

	var instID string = ""
	var pubIP string = ""
	var pvIP string = ""
	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		instID = civoCloudState.InstanceIDs.ControlNodes[idx]
		pubIP = civoCloudState.IPv4.IPControlplane[idx]
		pvIP = civoCloudState.IPv4.PrivateIPControlplane[idx]
	case utils.ROLE_WP:
		instID = civoCloudState.InstanceIDs.WorkerNodes[idx]
		pubIP = civoCloudState.IPv4.IPWorkerPlane[idx]
		pvIP = civoCloudState.IPv4.PrivateIPWorkerPlane[idx]
	case utils.ROLE_DS:
		instID = civoCloudState.InstanceIDs.DatabaseNode[idx]
		pubIP = civoCloudState.IPv4.IPDataStore[idx]
		pvIP = civoCloudState.IPv4.PrivateIPDataStore[idx]
	case utils.ROLE_LB:
		instID = civoCloudState.InstanceIDs.LoadBalancerNode
		pubIP = civoCloudState.IPv4.IPLoadbalancer
		pvIP = civoCloudState.IPv4.PrivateIPLoadbalancer
	}

	if len(instID) != 0 {
		// instance id present
		if len(pubIP) != 0 && len(pvIP) != 0 {
			// all info present

			storage.Logger().Success("[skip] vm found", instID)
			return nil
		} else {
			// either one or > 1 info are absent
			err := watchInstance(obj, storage, instID, idx)
			return err
		}
	}
	return fmt.Errorf("[civo] not found or [skip] already deleted vm role:", obj.Metadata.Role)

}

// NewVM implements resources.CloudFactory.
func (obj *CivoProvider) NewVM(storage resources.StorageFactory, indexNo int) error {

	err := obj.foundStateVM(storage, indexNo)
	if err == nil {
		return nil
	}

	publicIP := "create"
	if !obj.Metadata.Public {
		publicIP = "none"
	}

	diskImg, err := civoClient.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return err
	}

	firewallID := ""

	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		firewallID = civoCloudState.NetworkIDs.FirewallIDControlPlaneNode
	case utils.ROLE_WP:
		firewallID = civoCloudState.NetworkIDs.FirewallIDWorkerNode
	case utils.ROLE_DS:
		firewallID = civoCloudState.NetworkIDs.FirewallIDDatabaseNode
	case utils.ROLE_LB:
		firewallID = civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode
	}

	networkID := civoCloudState.NetworkIDs.NetworkID

	instanceConfig := &civogo.InstanceConfig{
		Hostname:         obj.Metadata.ResName,
		InitialUser:      civoCloudState.SSHUser,
		Region:           obj.Region,
		FirewallID:       firewallID,
		Size:             obj.Metadata.VmType,
		TemplateID:       diskImg.ID,
		NetworkID:        networkID,
		SSHKeyID:         civoCloudState.SSHID,
		PublicIPRequired: publicIP,
		// Script:           initializationScript,
	}

	inst, err := civoClient.CreateInstance(instanceConfig)
	if err != nil {
		return err
	}

	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		civoCloudState.InstanceIDs.ControlNodes[indexNo] = inst.ID
	case utils.ROLE_WP:
		civoCloudState.InstanceIDs.WorkerNodes[indexNo] = inst.ID
	case utils.ROLE_DS:
		civoCloudState.InstanceIDs.DatabaseNode[indexNo] = inst.ID
	case utils.ROLE_LB:
		civoCloudState.InstanceIDs.LoadBalancerNode = inst.ID
	}

	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)

	if err := saveStateHelper(storage, path); err != nil {
		return err
	}

	if err := watchInstance(obj, storage, inst.ID, indexNo); err != nil {
		return err
	}

	storage.Logger().Success("[civo] Created vm", obj.Metadata.ResName)
	return nil
}

// DelVM implements resources.CloudFactory.
func (obj *CivoProvider) DelVM(storage resources.StorageFactory, indexNo int) error {
	err := obj.foundStateVM(storage, indexNo)
	if err != nil {
		storage.Logger().Success(err.Error())
		return nil
	}

	instID := ""

	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		instID = civoCloudState.InstanceIDs.ControlNodes[indexNo]
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.ControlNodes[indexNo] = ""
		civoCloudState.IPv4.IPControlplane[indexNo] = ""
		civoCloudState.IPv4.PrivateIPControlplane[indexNo] = ""

	case utils.ROLE_WP:
		instID = civoCloudState.InstanceIDs.WorkerNodes[indexNo]
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.WorkerNodes[indexNo] = ""
		civoCloudState.IPv4.IPWorkerPlane[indexNo] = ""
		civoCloudState.IPv4.PrivateIPWorkerPlane[indexNo] = ""

	case utils.ROLE_DS:
		instID = civoCloudState.InstanceIDs.DatabaseNode[indexNo]
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.DatabaseNode[indexNo] = ""
		civoCloudState.IPv4.IPDataStore[indexNo] = ""
		civoCloudState.IPv4.PrivateIPDataStore[indexNo] = ""

	case utils.ROLE_LB:
		instID = civoCloudState.InstanceIDs.LoadBalancerNode
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.LoadBalancerNode = ""
		civoCloudState.IPv4.IPLoadbalancer = ""
		civoCloudState.IPv4.PrivateIPLoadbalancer = ""
	}

	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)

	if err := saveStateHelper(storage, path); err != nil {
		return err
	}
	time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
	storage.Logger().Success("[civo] Deleted vm", instID)
	return nil
}

func watchInstance(obj *CivoProvider, storage resources.StorageFactory, instID string, idx int) error {
	for {
		getInstance, err := civoClient.GetInstance(instID)
		if err != nil {
			return err
		}

		if getInstance.Status == "ACTIVE" {

			pubIP := getInstance.PublicIP
			pvIP := getInstance.PrivateIP

			switch obj.Metadata.Role {
			case utils.ROLE_CP:
				civoCloudState.IPv4.IPControlplane[idx] = pubIP
				civoCloudState.IPv4.PrivateIPControlplane[idx] = pvIP
			case utils.ROLE_WP:
				civoCloudState.IPv4.IPWorkerPlane[idx] = pubIP
				civoCloudState.IPv4.PrivateIPWorkerPlane[idx] = pvIP

				// make it isComplete when the workernode [idx -1] == len of it
				if len(civoCloudState.InstanceIDs.WorkerNodes) == idx+1 {
					civoCloudState.IsCompleted = true
				}
			case utils.ROLE_DS:
				civoCloudState.IPv4.IPDataStore[idx] = pubIP
				civoCloudState.IPv4.PrivateIPDataStore[idx] = pvIP
			case utils.ROLE_LB:
				civoCloudState.IPv4.IPLoadbalancer = pubIP
				civoCloudState.IPv4.PrivateIPLoadbalancer = pvIP
			}

			path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)

			if err := saveStateHelper(storage, path); err != nil {
				return err
			}

			return nil
		}
		storage.Logger().Print("[civo] waiting for vm to be ready..", obj.Metadata.ResName)
		time.Sleep(10 * time.Second)
	}
}
