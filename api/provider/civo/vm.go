package civo

import (
	"fmt"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func (obj *CivoProvider) foundStateVM(storage resources.StorageFactory, idx int, creationMode bool) error {

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
			if creationMode {
				storage.Logger().Success("[skip] vm found", instID)
			}
			return nil
		} else {
			// either one or > 1 info are absent
			err := watchInstance(obj, storage, instID, idx)
			return err
		}
	}
	if creationMode {
		return fmt.Errorf("[civo] vm not found")
	}
	return fmt.Errorf("[skip] already deleted vm having role:", obj.Metadata.Role)

}

// NewVM implements resources.CloudFactory.
func (obj *CivoProvider) NewVM(storage resources.StorageFactory, indexNo int) error {

	if obj.Metadata.Role == utils.ROLE_DS && indexNo > 0 {
		storage.Logger().Note("[skip] currently multiple datastore not supported")
		return nil
	}

	err := obj.foundStateVM(storage, indexNo, true)
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
		// Script:           initializationScript,  // TODO: add the os updates and other non necessary things before we try to configure in kubernetes may be security fixes
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
	err := obj.foundStateVM(storage, indexNo, false)
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
		civoCloudState.HostNames.ControlNodes[indexNo] = ""

	case utils.ROLE_WP:
		instID = civoCloudState.InstanceIDs.WorkerNodes[indexNo]
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.WorkerNodes[indexNo] = ""
		civoCloudState.IPv4.IPWorkerPlane[indexNo] = ""
		civoCloudState.IPv4.PrivateIPWorkerPlane[indexNo] = ""
		civoCloudState.HostNames.WorkerNodes[indexNo] = ""

	case utils.ROLE_DS:
		instID = civoCloudState.InstanceIDs.DatabaseNode[indexNo]
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.DatabaseNode[indexNo] = ""
		civoCloudState.IPv4.IPDataStore[indexNo] = ""
		civoCloudState.IPv4.PrivateIPDataStore[indexNo] = ""
		civoCloudState.HostNames.DatabaseNode[indexNo] = ""

	case utils.ROLE_LB:
		instID = civoCloudState.InstanceIDs.LoadBalancerNode
		_, err := civoClient.DeleteInstance(instID)
		if err != nil {
			return err
		}
		civoCloudState.InstanceIDs.LoadBalancerNode = ""
		civoCloudState.IPv4.IPLoadbalancer = ""
		civoCloudState.IPv4.PrivateIPLoadbalancer = ""
		civoCloudState.HostNames.LoadBalancerNode = ""
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
		//NOTE: this is prone to network failure

		currRetryCounter := 0
		var getInst *civogo.Instance
		for currRetryCounter < utils.MAX_WATCH_RETRY_COUNT {
			var err error
			getInst, err = civoClient.GetInstance(instID)
			if err != nil {
				currRetryCounter++
				storage.Logger().Err(fmt.Sprintln("RETRYING", err))
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == utils.MAX_WATCH_RETRY_COUNT {
			return fmt.Errorf("[civo] failed to get the state of vm")
		}

		if getInst.Status == "ACTIVE" {

			pubIP := getInst.PublicIP
			pvIP := getInst.PrivateIP
			hostNam := getInst.Hostname

			switch obj.Metadata.Role {
			case utils.ROLE_CP:
				civoCloudState.IPv4.IPControlplane[idx] = pubIP
				civoCloudState.IPv4.PrivateIPControlplane[idx] = pvIP
				civoCloudState.HostNames.ControlNodes[idx] = hostNam
				if len(civoCloudState.InstanceIDs.ControlNodes) == idx+1 && len(civoCloudState.InstanceIDs.WorkerNodes) == 0 {
					// no wp set so it is the final cloud provisioning
					civoCloudState.IsCompleted = true
				}
			case utils.ROLE_WP:
				civoCloudState.IPv4.IPWorkerPlane[idx] = pubIP
				civoCloudState.IPv4.PrivateIPWorkerPlane[idx] = pvIP
				civoCloudState.HostNames.WorkerNodes[idx] = hostNam

				// make it isComplete when the workernode [idx -1] == len of it
				if len(civoCloudState.InstanceIDs.WorkerNodes) == idx+1 {
					civoCloudState.IsCompleted = true
				}
			case utils.ROLE_DS:
				civoCloudState.IPv4.IPDataStore[idx] = pubIP
				civoCloudState.IPv4.PrivateIPDataStore[idx] = pvIP
				civoCloudState.HostNames.DatabaseNode[idx] = hostNam
			case utils.ROLE_LB:
				civoCloudState.IPv4.IPLoadbalancer = pubIP
				civoCloudState.IPv4.PrivateIPLoadbalancer = pvIP
				civoCloudState.HostNames.LoadBalancerNode = hostNam
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
