package civo

import (
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/pkg/resources"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
)

func (obj *CivoProvider) foundStateVM(storage resources.StorageFactory, idx int, creationMode bool, role consts.KsctlRole, name string) error {

	var instID string = ""
	var pubIP string = ""
	var pvIP string = ""
	switch role {
	case consts.RoleCp:
		instID = civoCloudState.InstanceIDs.ControlNodes[idx]
		pubIP = civoCloudState.IPv4.IPControlplane[idx]
		pvIP = civoCloudState.IPv4.PrivateIPControlplane[idx]
	case consts.RoleWp:
		instID = civoCloudState.InstanceIDs.WorkerNodes[idx]
		pubIP = civoCloudState.IPv4.IPWorkerPlane[idx]
		pvIP = civoCloudState.IPv4.PrivateIPWorkerPlane[idx]
	case consts.RoleDs:
		instID = civoCloudState.InstanceIDs.DatabaseNode[idx]
		pubIP = civoCloudState.IPv4.IPDataStore[idx]
		pvIP = civoCloudState.IPv4.PrivateIPDataStore[idx]
	case consts.RoleLb:
		instID = civoCloudState.InstanceIDs.LoadBalancerNode
		pubIP = civoCloudState.IPv4.IPLoadbalancer
		pvIP = civoCloudState.IPv4.PrivateIPLoadbalancer
	}

	if len(instID) != 0 {
		// instance id present
		if len(pubIP) != 0 && len(pvIP) != 0 {
			// all info present
			if creationMode {
				log.Print("skipped vm found", "id", instID)
			}
			return nil
		} else {
			// either one or > 1 info are absent
			err := watchInstance(obj, storage, instID, idx, role, name)
			return err
		}
	}
	if creationMode {
		return log.NewError("vm not found")
	}
	return log.NewError("skipped already deleted vm having role: %s", role)
}

// NewVM implements resources.CloudFactory.
func (obj *CivoProvider) NewVM(storage resources.StorageFactory, index int) error {

	name := obj.metadata.resName
	indexNo := index
	role := obj.metadata.role
	vmtype := obj.metadata.vmType
	obj.mxRole.Unlock()
	obj.mxName.Unlock()
	obj.mxVMType.Unlock()

	log.Debug("Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	if role == consts.RoleDs && indexNo > 0 {
		log.Note("skipped currently multiple datastore not supported", "vmName", name)
		return nil
	}

	err := obj.foundStateVM(storage, indexNo, true, role, name)
	if err == nil {
		return log.NewError(err.Error())
	}

	publicIP := "create"
	if !obj.metadata.public {
		publicIP = "none"
	}

	diskImg, err := obj.client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return log.NewError(err.Error())
	}

	firewallID := ""

	switch role {
	case consts.RoleCp:
		firewallID = civoCloudState.NetworkIDs.FirewallIDControlPlaneNode
	case consts.RoleWp:
		firewallID = civoCloudState.NetworkIDs.FirewallIDWorkerNode
	case consts.RoleDs:
		firewallID = civoCloudState.NetworkIDs.FirewallIDDatabaseNode
	case consts.RoleLb:
		firewallID = civoCloudState.NetworkIDs.FirewallIDLoadBalancerNode
	}

	networkID := civoCloudState.NetworkIDs.NetworkID

	instanceConfig := &civogo.InstanceConfig{
		Hostname:         name,
		InitialUser:      civoCloudState.SSHUser,
		Region:           obj.region,
		FirewallID:       firewallID,
		Size:             vmtype,
		TemplateID:       diskImg.ID,
		NetworkID:        networkID,
		SSHKeyID:         civoCloudState.SSHID,
		PublicIPRequired: publicIP,
		// Script:           initializationScript,  // TODO: add the os updates and other non necessary things before we try to configure in kubernetes may be security fixes
	}

	log.Debug("Printing", "instanceConfig", instanceConfig)
	log.Print("Creating vm", "name", name)

	var inst *civogo.Instance
	inst, err = obj.client.CreateInstance(instanceConfig)
	if err != nil {
		return log.NewError(err.Error())
	}

	done := make(chan struct{})
	var errCreateVM error

	go func() {
		obj.mxState.Lock()

		switch role {
		case consts.RoleCp:
			civoCloudState.InstanceIDs.ControlNodes[indexNo] = inst.ID
		case consts.RoleWp:
			civoCloudState.InstanceIDs.WorkerNodes[indexNo] = inst.ID
		case consts.RoleDs:
			civoCloudState.InstanceIDs.DatabaseNode[indexNo] = inst.ID
		case consts.RoleLb:
			civoCloudState.InstanceIDs.LoadBalancerNode = inst.ID
		}

		path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

		if err := saveStateHelper(storage, path); err != nil {
			errCreateVM = err
			obj.mxState.Unlock()
			close(done)
			return
		}
		obj.mxState.Unlock()

		if err := watchInstance(obj, storage, inst.ID, indexNo, role, name); err != nil {
			errCreateVM = err
			close(done)
			return
		}

		log.Success("Created vm", "vmName", name)

		close(done)
	}()

	<-done

	return errCreateVM
}

// DelVM implements resources.CloudFactory.
func (obj *CivoProvider) DelVM(storage resources.StorageFactory, index int) error {

	indexNo := index
	role := obj.metadata.role
	obj.mxRole.Unlock()

	log.Debug("Printing", "role", role, "indexNo", indexNo)

	err := obj.foundStateVM(storage, indexNo, false, role, "")
	if err != nil {
		log.Success(err.Error())
		return nil
	}

	instID := ""
	done := make(chan struct{})
	var errCreateVM error

	switch role {
	case consts.RoleCp:
		instID = civoCloudState.InstanceIDs.ControlNodes[indexNo]
		log.Debug("Printing", "instID", instID)

		go func() {
			defer close(done)
			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}

			obj.mxState.Lock()
			defer obj.mxState.Unlock()

			civoCloudState.InstanceIDs.ControlNodes[indexNo] = ""
			civoCloudState.IPv4.IPControlplane[indexNo] = ""
			civoCloudState.IPv4.PrivateIPControlplane[indexNo] = ""
			civoCloudState.HostNames.ControlNodes[indexNo] = ""

			path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

			if err := saveStateHelper(storage, path); err != nil {
				errCreateVM = err
				return
			}

			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success("Deleted vm", "id", instID)
		}()

		<-done

	case consts.RoleWp:
		go func() {
			defer close(done)
			instID = civoCloudState.InstanceIDs.WorkerNodes[indexNo]
			log.Debug("Printing", "instID", instID)

			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			obj.mxState.Lock()
			defer obj.mxState.Unlock()
			civoCloudState.InstanceIDs.WorkerNodes[indexNo] = ""
			civoCloudState.IPv4.IPWorkerPlane[indexNo] = ""
			civoCloudState.IPv4.PrivateIPWorkerPlane[indexNo] = ""
			civoCloudState.HostNames.WorkerNodes[indexNo] = ""
			path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

			if err := saveStateHelper(storage, path); err != nil {
				errCreateVM = err
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success("Deleted vm", "id", instID)
		}()
		<-done

	case consts.RoleDs:
		go func() {
			defer close(done)
			instID = civoCloudState.InstanceIDs.DatabaseNode[indexNo]
			log.Debug("Printing", "instID", instID)

			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			obj.mxState.Lock()
			defer obj.mxState.Unlock()
			civoCloudState.InstanceIDs.DatabaseNode[indexNo] = ""
			civoCloudState.IPv4.IPDataStore[indexNo] = ""
			civoCloudState.IPv4.PrivateIPDataStore[indexNo] = ""
			civoCloudState.HostNames.DatabaseNode[indexNo] = ""
			path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

			if err := saveStateHelper(storage, path); err != nil {
				errCreateVM = err
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success("Deleted vm", "id", instID)
		}()
		<-done

	case consts.RoleLb:
		go func() {
			defer close(done)
			instID = civoCloudState.InstanceIDs.LoadBalancerNode
			log.Debug("Printing", "instID", instID)

			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			obj.mxState.Lock()
			defer obj.mxState.Unlock()
			civoCloudState.InstanceIDs.LoadBalancerNode = ""
			civoCloudState.IPv4.IPLoadbalancer = ""
			civoCloudState.IPv4.PrivateIPLoadbalancer = ""
			civoCloudState.HostNames.LoadBalancerNode = ""
			path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

			if err := saveStateHelper(storage, path); err != nil {
				errCreateVM = err
				close(done)
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success("Deleted vm", "id", instID)
		}()
		<-done
	}
	log.Debug("Printing", "cloudState", civoCloudState)

	return errCreateVM
}

func watchInstance(obj *CivoProvider, storage resources.StorageFactory, instID string, idx int, role consts.KsctlRole, name string) error {
	for {
		// NOTE: this is prone to network failure

		currRetryCounter := consts.KsctlCounterConsts(0)
		var getInst *civogo.Instance
		for currRetryCounter < consts.CounterMaxWatchRetryCount {
			var err error

			getInst, err = obj.client.GetInstance(instID)
			if err != nil {
				currRetryCounter++
				log.Warn("RETRYING", err)
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == consts.CounterMaxWatchRetryCount {
			return log.NewError("failed to get the state of vm")
		}

		if getInst.Status == "ACTIVE" {

			pubIP := getInst.PublicIP
			pvIP := getInst.PrivateIP
			hostNam := getInst.Hostname

			obj.mxState.Lock()
			defer obj.mxState.Unlock()
			// critical section
			switch role {
			case consts.RoleCp:
				civoCloudState.IPv4.IPControlplane[idx] = pubIP
				civoCloudState.IPv4.PrivateIPControlplane[idx] = pvIP
				civoCloudState.HostNames.ControlNodes[idx] = hostNam
				if len(civoCloudState.InstanceIDs.ControlNodes) == idx+1 && len(civoCloudState.InstanceIDs.WorkerNodes) == 0 {
					// no wp set so it is the final cloud provisioning
					civoCloudState.IsCompleted = true
				}
			case consts.RoleWp:
				civoCloudState.IPv4.IPWorkerPlane[idx] = pubIP
				civoCloudState.IPv4.PrivateIPWorkerPlane[idx] = pvIP
				civoCloudState.HostNames.WorkerNodes[idx] = hostNam

				// make it isComplete when the workernode [idx -1] == len of it
				if len(civoCloudState.InstanceIDs.WorkerNodes) == idx+1 {
					civoCloudState.IsCompleted = true
				}
			case consts.RoleDs:
				civoCloudState.IPv4.IPDataStore[idx] = pubIP
				civoCloudState.IPv4.PrivateIPDataStore[idx] = pvIP
				civoCloudState.HostNames.DatabaseNode[idx] = hostNam
			case consts.RoleLb:
				civoCloudState.IPv4.IPLoadbalancer = pubIP
				civoCloudState.IPv4.PrivateIPLoadbalancer = pvIP
				civoCloudState.HostNames.LoadBalancerNode = hostNam
			}

			path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

			if err := saveStateHelper(storage, path); err != nil {
				return err
			}

			return nil
		}
		log.Debug("waiting for vm to be ready..", "vmName", name, "Status", getInst.Status)
		time.Sleep(10 * time.Second)
	}
}
