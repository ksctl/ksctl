// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package civo

import (
	"fmt"
	"github.com/ksctl/ksctl/pkg/providers"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/waiter"
	"time"

	"github.com/civo/civogo"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"

	"github.com/ksctl/ksctl/pkg/consts"
)

func (p *Provider) foundStateVM(idx int, creationMode bool, role consts.KsctlRole, name string) error {

	var instID string = ""
	var pubIP string = ""
	var pvIP string = ""
	switch role {
	case consts.RoleCp:

		instID = p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs[idx]
		pubIP = p.state.CloudInfra.Civo.InfoControlPlanes.PublicIPs[idx]
		pvIP = p.state.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[idx]
	case consts.RoleWp:
		instID = p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[idx]
		pubIP = p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[idx]
		pvIP = p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[idx]
	case consts.RoleDs:
		instID = p.state.CloudInfra.Civo.InfoDatabase.VMIDs[idx]
		pubIP = p.state.CloudInfra.Civo.InfoDatabase.PublicIPs[idx]
		pvIP = p.state.CloudInfra.Civo.InfoDatabase.PrivateIPs[idx]
	case consts.RoleLb:
		instID = p.state.CloudInfra.Civo.InfoLoadBalancer.VMID
		pubIP = p.state.CloudInfra.Civo.InfoLoadBalancer.PublicIP
		pvIP = p.state.CloudInfra.Civo.InfoLoadBalancer.PrivateIP
	}

	if creationMode {
		// creation mode
		if len(instID) != 0 {
			// instance id present
			if len(pubIP) != 0 && len(pvIP) != 0 {
				p.l.Print(p.ctx, "skipped vm found", "id", instID)
				return nil
			} else {
				// either one or > 1 info are absent
				err := p.watchInstance(instID, idx, role, name)
				return err
			}
		}
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			p.l.NewError(p.ctx, "vm not found"),
		)

	} else {
		// deletion mode
		if len(instID) != 0 {
			// need to delete
			p.l.Print(p.ctx, "Deleting the VM")
			return nil
		}
		// already deleted
		return ksctlErrors.NewError(ksctlErrors.ErrNoMatchingRecordsFound)
	}
}

func (p *Provider) NewVM(index int) error {

	name := <-p.chResName
	indexNo := index
	role := <-p.chRole
	vmtype := <-p.chVMType

	p.l.Debug(p.ctx, "Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	if err := p.foundStateVM(indexNo, true, role, name); err == nil {
		return nil
	} else {
		if !ksctlErrors.IsNoMatchingRecordsFound(err) {
			return err
		}
	}

	publicIP := "create"
	if !p.public {
		publicIP = "none"
	}

	diskImg, err := p.client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return err
	}

	firewallID := ""

	switch role {
	case consts.RoleCp:
		firewallID = p.state.CloudInfra.Civo.FirewallIDControlPlanes
	case consts.RoleWp:
		firewallID = p.state.CloudInfra.Civo.FirewallIDWorkerNodes
	case consts.RoleDs:
		firewallID = p.state.CloudInfra.Civo.FirewallIDDatabaseNodes
	case consts.RoleLb:
		firewallID = p.state.CloudInfra.Civo.FirewallIDLoadBalancer
	}

	networkID := p.state.CloudInfra.Civo.NetworkID

	initScript, err := providers.CloudInitScript(name)
	if err != nil {
		return err
	}
	p.l.Debug(p.ctx, "initscript", "script", initScript)

	instanceConfig := &civogo.InstanceConfig{
		Hostname:         name,
		InitialUser:      p.state.CloudInfra.Civo.B.SSHUser,
		Region:           p.Region,
		FirewallID:       firewallID,
		Size:             vmtype,
		TemplateID:       diskImg.ID,
		NetworkID:        networkID,
		SSHKeyID:         p.state.CloudInfra.Civo.B.SSHID,
		PublicIPRequired: publicIP,
		Script:           initScript,
	}

	p.l.Debug(p.ctx, "Printing", "instanceConfig", instanceConfig)
	p.l.Print(p.ctx, "Creating vm", "name", name)

	var inst *civogo.Instance
	inst, err = p.client.CreateInstance(instanceConfig)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	var errCreateVM error

	go func() {
		p.mu.Lock()

		switch role {
		case consts.RoleCp:
			p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs[indexNo] = inst.ID
			p.state.CloudInfra.Civo.InfoControlPlanes.VMSizes[indexNo] = vmtype
		case consts.RoleWp:
			p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[indexNo] = inst.ID
			p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes[indexNo] = vmtype
		case consts.RoleDs:
			p.state.CloudInfra.Civo.InfoDatabase.VMIDs[indexNo] = inst.ID
			p.state.CloudInfra.Civo.InfoDatabase.VMSizes[indexNo] = vmtype
		case consts.RoleLb:
			p.state.CloudInfra.Civo.InfoLoadBalancer.VMID = inst.ID
			p.state.CloudInfra.Civo.InfoLoadBalancer.VMSize = vmtype
		}

		if err := p.store.Write(p.state); err != nil {
			errCreateVM = err
			p.mu.Unlock()
			close(done)
			return
		}
		p.mu.Unlock()

		if err := p.watchInstance(inst.ID, indexNo, role, name); err != nil {
			errCreateVM = err
			close(done)
			return
		}

		p.l.Success(p.ctx, "Created vm", "vmName", name)

		close(done)
	}()

	<-done

	return errCreateVM
}

func (p *Provider) DelVM(index int) error {

	indexNo := index
	role := <-p.chRole

	p.l.Debug(p.ctx, "Printing", "role", role, "indexNo", indexNo)

	if err := p.foundStateVM(indexNo, false, role, ""); err != nil {
		if ksctlErrors.IsNoMatchingRecordsFound(err) {
			p.l.Success(p.ctx, "skipped already deleted vm")
		}
	}

	instID := ""
	done := make(chan struct{})
	var errCreateVM error

	switch role {
	case consts.RoleCp, consts.RoleWp, consts.RoleDs:
		var vmState *statefile.CivoStateVMs
		switch role {
		case consts.RoleCp:
			vmState = &p.state.CloudInfra.Civo.InfoControlPlanes
		case consts.RoleWp:
			vmState = &p.state.CloudInfra.Civo.InfoWorkerPlanes
		case consts.RoleDs:
			vmState = &p.state.CloudInfra.Civo.InfoDatabase
		}
		instID = vmState.VMIDs[indexNo]
		p.l.Debug(p.ctx, "Printing", "instID", instID)

		go func() {
			defer close(done)
			_, err := p.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}

			p.mu.Lock()
			defer p.mu.Unlock()

			vmState.VMIDs[indexNo] = ""
			vmState.PublicIPs[indexNo] = ""
			vmState.PrivateIPs[indexNo] = ""
			vmState.Hostnames[indexNo] = ""
			vmState.VMSizes[indexNo] = ""

			if err := p.store.Write(p.state); err != nil {
				errCreateVM = err
				return
			}

			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			p.l.Success(p.ctx, "Deleted vm", "id", instID)
		}()

		<-done

	case consts.RoleLb:
		go func() {
			defer close(done)
			instID = p.state.CloudInfra.Civo.InfoLoadBalancer.VMID
			p.l.Debug(p.ctx, "Printing", "instID", instID)

			_, err := p.client.DeleteInstance(instID)
			if err != nil {
				errCreateVM = err
				return
			}
			p.mu.Lock()
			defer p.mu.Unlock()
			p.state.CloudInfra.Civo.InfoLoadBalancer.VMID = ""
			p.state.CloudInfra.Civo.InfoLoadBalancer.PublicIP = ""
			p.state.CloudInfra.Civo.InfoLoadBalancer.PrivateIP = ""
			p.state.CloudInfra.Civo.InfoLoadBalancer.HostName = ""
			p.state.CloudInfra.Civo.InfoLoadBalancer.VMSize = ""

			if err := p.store.Write(p.state); err != nil {
				errCreateVM = err
				close(done)
				return
			}
			time.Sleep(2 * time.Second) // NOTE: to make sure the instances gets time to be deleted
			p.l.Success(p.ctx, "Deleted vm", "id", instID)
		}()
		<-done
	}

	return errCreateVM
}

func (p *Provider) watchInstance(instID string, idx int, role consts.KsctlRole, name string) error {

	expoBackoff := waiter.NewWaiter(
		10*time.Second,
		2,
		2*int(consts.CounterMaxWatchRetryCount),
	)

	var getInst *civogo.Instance
	_err := expoBackoff.Run(
		p.ctx,
		p.l,
		func() (err error) {
			getInst, err = p.client.GetInstance(instID)
			return err
		},
		func() bool {
			return getInst.Status == "ACTIVE"
		},
		nil,
		func() error {
			pubIP := getInst.PublicIP
			pvIP := getInst.PrivateIP
			hostNam := getInst.Hostname

			p.mu.Lock()
			defer p.mu.Unlock()
			// critical section
			switch role {
			case consts.RoleCp:
				p.state.CloudInfra.Civo.InfoControlPlanes.PublicIPs[idx] = pubIP
				p.state.CloudInfra.Civo.InfoControlPlanes.PrivateIPs[idx] = pvIP
				p.state.CloudInfra.Civo.InfoControlPlanes.Hostnames[idx] = hostNam
				if len(p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs) == idx+1 && len(p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs) == 0 {
					// no wp set so it is the final cloud provisioning
					p.state.CloudInfra.Civo.B.IsCompleted = true
				}
			case consts.RoleWp:
				p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[idx] = pubIP
				p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[idx] = pvIP
				p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[idx] = hostNam

				// make it isComplete when the workernode [idx -1] == len of it
				if len(p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs) == idx+1 {
					p.state.CloudInfra.Civo.B.IsCompleted = true
				}
			case consts.RoleDs:
				p.state.CloudInfra.Civo.InfoDatabase.PublicIPs[idx] = pubIP
				p.state.CloudInfra.Civo.InfoDatabase.PrivateIPs[idx] = pvIP
				p.state.CloudInfra.Civo.InfoDatabase.Hostnames[idx] = hostNam
			case consts.RoleLb:
				p.state.CloudInfra.Civo.InfoLoadBalancer.PublicIP = pubIP
				p.state.CloudInfra.Civo.InfoLoadBalancer.PrivateIP = pvIP
				p.state.CloudInfra.Civo.InfoLoadBalancer.HostName = hostNam
			}

			return p.store.Write(p.state)
		},
		fmt.Sprintf("waiting for vm %s to be ready", name),
	)
	if _err != nil {
		return _err
	}

	return nil
}
