package civo

import (
	"time"

	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/types"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func (obj *CivoProvider) foundStateVM(storage types.StorageFactory, idx int, creationMode bool, role consts.KsctlRole, name string) error {

	var instID string = ""
	var pubIP string = ""
	var pvIP string = ""
	switch role {
	case consts.RoleCp:

		instID = mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs[idx]
		pubIP = mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs[idx]
		pvIP = mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[idx]
	case consts.RoleWp:
		instID = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[idx]
		pubIP = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[idx]
		pvIP = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[idx]
	case consts.RoleDs:
		instID = mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs[idx]
		pubIP = mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs[idx]
		pvIP = mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs[idx]
	case consts.RoleLb:
		instID = mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.VMID
		pubIP = mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PublicIP
		pvIP = mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PrivateIP
	}

	if len(instID) != 0 {
		// instance id present
		if len(pubIP) != 0 && len(pvIP) != 0 {
			// all info present
			if creationMode {
				log.Print(civoCtx, "skipped vm found", "id", instID)
			}
			return nil
		} else {
			// either one or > 1 info are absent
			err := watchInstance(obj, storage, instID, idx, role, name)
			return err
		}
	}
	if creationMode {
		return log.NewError(civoCtx, "vm not found")
	}
	return log.NewError(civoCtx, "skipped already deleted vm", "role", role)
}

// NewVM implements types.CloudFactory.
func (obj *CivoProvider) NewVM(storage types.StorageFactory, index int) error {

	name := <-obj.chResName
	indexNo := index
	role := <-obj.chRole
	vmtype := <-obj.chVMType

	log.Debug(civoCtx, "Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	err := obj.foundStateVM(storage, indexNo, true, role, name)
	if err == nil {
		return nil
	}

	publicIP := "create"
	if !obj.metadata.public {
		publicIP = "none"
	}

	diskImg, err := obj.client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return err
	}

	firewallID := ""

	switch role {
	case consts.RoleCp:
		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDControlPlanes
	case consts.RoleWp:
		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDWorkerNodes
	case consts.RoleDs:
		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDDatabaseNodes
	case consts.RoleLb:
		firewallID = mainStateDocument.CloudInfra.Civo.FirewallIDLoadBalancer
	}

	networkID := mainStateDocument.CloudInfra.Civo.NetworkID

	initScript, err := helpers.GenerateInitScriptForVM(name)
	if err != nil {
		return log.NewError(civoCtx, "failed gen init script for vm", "Reason", err)
	}
	log.Debug(civoCtx, "initscript", "script", initScript)

	instanceConfig := &civogo.InstanceConfig{
		Hostname:         name,
		InitialUser:      mainStateDocument.CloudInfra.Civo.B.SSHUser,
		Region:           obj.region,
		FirewallID:       firewallID,
		Size:             vmtype,
		TemplateID:       diskImg.ID,
		NetworkID:        networkID,
		SSHKeyID:         mainStateDocument.CloudInfra.Civo.B.SSHID,
		PublicIPRequired: publicIP,
		Script:           initScript,
	}

	log.Debug(civoCtx, "Printing", "instanceConfig", instanceConfig)
	log.Print(civoCtx, "Creating vm", "name", name)

	var inst *civogo.Instance
	inst, err = obj.client.CreateInstance(instanceConfig)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	var errCreateVM error

	go func() {
		obj.mu.Lock()

		switch role {
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs[indexNo] = inst.ID
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[indexNo] = inst.ID
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs[indexNo] = inst.ID
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.VMID = inst.ID
		}

		if err := storage.Write(mainStateDocument); err != nil {
			errCreateVM = err
			obj.mu.Unlock()
			close(done)
			return
		}
		obj.mu.Unlock()

		if err := watchInstance(obj, storage, inst.ID, indexNo, role, name); err != nil {
			errCreateVM = err
			close(done)
			return
		}

		log.Success(civoCtx, "Created vm", "vmName", name)

		close(done)
	}()

	<-done

	return errCreateVM
}

// DelVM implements types.CloudFactory.
func (obj *CivoProvider) DelVM(storage types.StorageFactory, index int) error {

	indexNo := index
	role := <-obj.chRole

	log.Debug(civoCtx, "Printing", "role", role, "indexNo", indexNo)

	err := obj.foundStateVM(storage, indexNo, false, role, "")
	if err != nil {
		log.Success(civoCtx, err.Error()) // Try to make it better
		return nil
	}

	instID := ""
	done := make(chan struct{})
	var errCreateVM error

	switch role {
	case consts.RoleCp:
		instID = mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs[indexNo]
		log.Debug(civoCtx, "Printing", "instID", instID)

		go func() {
			defer close(done)
			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}

			obj.mu.Lock()
			defer obj.mu.Unlock()

			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames[indexNo] = ""

			if err := storage.Write(mainStateDocument); err != nil {
				errCreateVM = err
				return
			}

			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success(civoCtx, "Deleted vm", "id", instID)
		}()

		<-done

	case consts.RoleWp:
		go func() {
			defer close(done)
			instID = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[indexNo]
			log.Debug(civoCtx, "Printing", "instID", instID)

			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			obj.mu.Lock()
			defer obj.mu.Unlock()
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[indexNo] = ""

			if err := storage.Write(mainStateDocument); err != nil {
				errCreateVM = err
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success(civoCtx, "Deleted vm", "id", instID)
		}()
		<-done

	case consts.RoleDs:
		go func() {
			defer close(done)
			instID = mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs[indexNo]
			log.Debug(civoCtx, "Printing", "instID", instID)

			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			obj.mu.Lock()
			defer obj.mu.Unlock()
			mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs[indexNo] = ""
			mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames[indexNo] = ""

			if err := storage.Write(mainStateDocument); err != nil {
				errCreateVM = err
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success(civoCtx, "Deleted vm", "id", instID)
		}()
		<-done

	case consts.RoleLb:
		go func() {
			defer close(done)
			instID = mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.VMID
			log.Debug(civoCtx, "Printing", "instID", instID)

			_, err := obj.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			obj.mu.Lock()
			defer obj.mu.Unlock()
			mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.VMID = ""
			mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PublicIP = ""
			mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PrivateIP = ""
			mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.HostName = ""

			if err := storage.Write(mainStateDocument); err != nil {
				errCreateVM = err
				close(done)
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			log.Success(civoCtx, "Deleted vm", "id", instID)
		}()
		<-done
	}
	log.Debug(civoCtx, "Printing", "cloudState", mainStateDocument)

	return errCreateVM
}

func watchInstance(obj *CivoProvider, storage types.StorageFactory, instID string, idx int, role consts.KsctlRole, name string) error {
	for {
		// NOTE: this is prone to network failure

		currRetryCounter := consts.KsctlCounterConsts(0)
		var getInst *civogo.Instance
		for currRetryCounter < consts.CounterMaxWatchRetryCount {
			var err error

			getInst, err = obj.client.GetInstance(instID)
			if err != nil {
				currRetryCounter++
				log.Warn(civoCtx, "retrying", "err", err)
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == consts.CounterMaxWatchRetryCount {
			return log.NewError(civoCtx, "maximum retry reached failed to get the state of vm")
		}

		if getInst.Status == "ACTIVE" {

			pubIP := getInst.PublicIP
			pvIP := getInst.PrivateIP
			hostNam := getInst.Hostname

			obj.mu.Lock()
			defer obj.mu.Unlock()
			// critical section
			switch role {
			case consts.RoleCp:
				mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs[idx] = pubIP
				mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[idx] = pvIP
				mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames[idx] = hostNam
				if len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs) == idx+1 && len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs) == 0 {
					// no wp set so it is the final cloud provisioning
					mainStateDocument.CloudInfra.Civo.B.IsCompleted = true
				}
			case consts.RoleWp:
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[idx] = pubIP
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[idx] = pvIP
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[idx] = hostNam

				// make it isComplete when the workernode [idx -1] == len of it
				if len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs) == idx+1 {
					mainStateDocument.CloudInfra.Civo.B.IsCompleted = true
				}
			case consts.RoleDs:
				mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs[idx] = pubIP
				mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs[idx] = pvIP
				mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames[idx] = hostNam
			case consts.RoleLb:
				mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PublicIP = pubIP
				mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PrivateIP = pvIP
				mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.HostName = hostNam
			}

			if err := storage.Write(mainStateDocument); err != nil {
				return err
			}

			return nil
		}
		log.Debug(civoCtx, "waiting for vm to be ready..", "vmName", name, "Status", getInst.Status)
		time.Sleep(10 * time.Second)
	}
}
