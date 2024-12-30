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

package azure

import (
	"encoding/base64"

	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/provider"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func (p *Provider) DelVM(index int) error {
	role := <-p.chRole
	indexNo := index

	p.l.Debug(p.ctx, "Printing", "role", role, "indexNo", indexNo)

	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = p.state.CloudInfra.Azure.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = p.state.CloudInfra.Azure.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = p.state.CloudInfra.Azure.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = p.state.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo]
	}

	if len(vmName) == 0 {
		p.l.Print(p.ctx, "skipped vm already deleted")
	} else {

		var errDel error //just to make sure its nil
		donePoll := make(chan struct{})
		go func() {
			defer close(donePoll)
			pollerResponse, err := p.client.BeginDeleteVM(vmName, nil)
			if err != nil {
				errDel = err
				return
			}
			p.l.Print(p.ctx, "deleting vm...", "name", vmName)

			_, err = p.client.PollUntilDoneDelVM(p.ctx, pollerResponse, nil)
			if err != nil {
				errDel = err
				return
			}
			p.mu.Lock()
			defer p.mu.Unlock()

			switch role {
			case consts.RoleWp:
				p.state.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo] = ""
				p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[indexNo] = ""
				p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes[indexNo] = ""
			case consts.RoleCp:
				p.state.CloudInfra.Azure.InfoControlPlanes.Names[indexNo] = ""
				p.state.CloudInfra.Azure.InfoControlPlanes.Hostnames[indexNo] = ""
				p.state.CloudInfra.Azure.InfoControlPlanes.VMSizes[indexNo] = ""
			case consts.RoleLb:
				p.state.CloudInfra.Azure.InfoLoadBalancer.Name = ""
				p.state.CloudInfra.Azure.InfoLoadBalancer.HostName = ""
				p.state.CloudInfra.Azure.InfoLoadBalancer.VMSize = ""
			case consts.RoleDs:
				p.state.CloudInfra.Azure.InfoDatabase.Names[indexNo] = ""
				p.state.CloudInfra.Azure.InfoDatabase.Hostnames[indexNo] = ""
				p.state.CloudInfra.Azure.InfoDatabase.VMSizes[indexNo] = ""
			}

			if err := p.store.Write(p.state); err != nil {
				errDel = err
				return
			}

		}()
		<-donePoll
		if errDel != nil {
			return errDel
		}
		p.l.Success(p.ctx, "Deleted the vm", "name", vmName)

	}

	if err := p.DeleteDisk(indexNo, role); err != nil {
		return err
	}

	if err := p.DeleteNetworkInterface(indexNo, role); err != nil {
		return err
	}

	if err := p.DeletePublicIP(indexNo, role); err != nil {
		return err
	}

	return nil
}

func (p *Provider) NewVM(index int) error {
	name := <-p.chResName
	indexNo := index
	role := <-p.chRole
	vmtype := <-p.chVMType

	p.l.Debug(p.ctx, "Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	pubIPName := name + "-pub"
	nicName := name + "-nic"
	diskName := name + "-disk"
	p.l.Debug(p.ctx, "Printing", "pubIPName", pubIPName, "NICName", nicName, "diskName", diskName)

	if err := p.CreatePublicIP(pubIPName, indexNo, role); err != nil {
		return err
	}

	pubIPID := ""
	nsgID := ""

	switch role {
	case consts.RoleWp:
		pubIPID = p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[indexNo]
		nsgID = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID
	case consts.RoleCp:
		nsgID = p.state.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID
		pubIPID = p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[indexNo]
	case consts.RoleLb:
		nsgID = p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID
		pubIPID = p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPID
	case consts.RoleDs:
		nsgID = p.state.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID
		pubIPID = p.state.CloudInfra.Azure.InfoDatabase.PublicIPIDs[indexNo]
	}

	p.l.Debug(p.ctx, "Printing", "PubIP_id", pubIPID, "NsgID", nsgID)

	if err := p.CreateNetworkInterface(nicName, p.state.CloudInfra.Azure.SubnetID, pubIPID, nsgID, indexNo, role); err != nil {
		return err
	}

	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = p.state.CloudInfra.Azure.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = p.state.CloudInfra.Azure.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = p.state.CloudInfra.Azure.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = p.state.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo]
	}
	if len(vmName) != 0 {
		p.l.Print(p.ctx, "skipped vm already created", "name", vmName)
		return nil
	}

	netInterfaceID := ""
	switch role {
	case consts.RoleWp:
		netInterfaceID = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[indexNo]
	case consts.RoleCp:
		netInterfaceID = p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[indexNo]
	case consts.RoleLb:
		netInterfaceID = p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID
	case consts.RoleDs:
		netInterfaceID = p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[indexNo]
	}
	p.l.Debug(p.ctx, "Printing", "netInterfaceID", netInterfaceID)

	initScript, err := provider.CloudInitScript(name)
	if err != nil {
		return err
	}

	p.l.Debug(p.ctx, "initscript", "script", initScript)

	parameters := armcompute.VirtualMachine{
		Location: utilities.Ptr(p.Region),
		Identity: &armcompute.VirtualMachineIdentity{
			Type: utilities.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     utilities.Ptr("0001-com-ubuntu-server-jammy"),
					Publisher: utilities.Ptr("Canonical"),
					SKU:       utilities.Ptr("22_04-lts-gen2"),
					Version:   utilities.Ptr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         utilities.Ptr(diskName),
					CreateOption: utilities.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					Caching:      utilities.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: utilities.Ptr(armcompute.StorageAccountTypesStandardLRS), // OSDisk type Standard/Premium HDD/SSD
					},
					//DiskSizeGB: utilities.Ptr[int32](100), // default 127G
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: utilities.Ptr(armcompute.VirtualMachineSizeTypes(vmtype)), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  utilities.Ptr(name),
				AdminUsername: utilities.Ptr(p.state.CloudInfra.Azure.B.SSHUser),
				CustomData:    utilities.Ptr(base64.StdEncoding.EncodeToString([]byte(initScript))),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: utilities.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    utilities.Ptr("/home/azureuser/.ssh/authorized_keys"),
								KeyData: utilities.Ptr(p.state.SSHKeyPair.PublicKey),
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: utilities.Ptr(netInterfaceID),
					},
				},
			},
		},
	}
	p.l.Debug(p.ctx, "Printing", "VMConfig", parameters)

	pollerResponse, err := p.client.BeginCreateVM(name, parameters, nil)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	var errCreateVM error
	go func() {
		defer close(done)
		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo] = name
			p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes[indexNo] = name
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.Names[indexNo] = name
			p.state.CloudInfra.Azure.InfoControlPlanes.VMSizes[indexNo] = name
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.Name = name
			p.state.CloudInfra.Azure.InfoLoadBalancer.VMSize = name
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.Names[indexNo] = name
			p.state.CloudInfra.Azure.InfoDatabase.VMSizes[indexNo] = name
		}
		if err := p.store.Write(p.state); err != nil {
			errCreateVM = err
			return
		}

	}()

	<-done
	if errCreateVM != nil {
		return errCreateVM
	}
	p.l.Print(p.ctx, "creating vm...", "name", name)

	errCreateVM = nil //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)

		resp, err := p.client.PollUntilDoneCreateVM(p.ctx, pollerResponse, nil)
		if err != nil {
			errCreateVM = err
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[indexNo] = diskName
			p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName

			if len(p.state.CloudInfra.Azure.InfoWorkerPlanes.Names) == indexNo+1 {
				p.state.CloudInfra.Azure.B.IsCompleted = true
			}

		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.DiskNames[indexNo] = diskName
			p.state.CloudInfra.Azure.InfoControlPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
			if len(p.state.CloudInfra.Azure.InfoControlPlanes.Names) == indexNo+1 && len(p.state.CloudInfra.Azure.InfoWorkerPlanes.Names) == 0 {
				// when its the last resource to be created and we are done with the last controlplane creation
				p.state.CloudInfra.Azure.B.IsCompleted = true
			}
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.DiskName = diskName
			p.state.CloudInfra.Azure.InfoLoadBalancer.HostName = *resp.Properties.OSProfile.ComputerName
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.DiskNames[indexNo] = diskName
			p.state.CloudInfra.Azure.InfoDatabase.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
		}

		if err := p.store.Write(p.state); err != nil {
			errCreateVM = err
			return
		}

	}()
	<-donePoll
	if errCreateVM != nil {
		return errCreateVM
	}

	p.l.Success(p.ctx, "Created virtual machine", "name", name)
	return nil
}

func (p *Provider) DeleteDisk(index int, role consts.KsctlRole) error {
	diskName := ""
	switch role {
	case consts.RoleWp:
		diskName = p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[index]
	case consts.RoleCp:
		diskName = p.state.CloudInfra.Azure.InfoControlPlanes.DiskNames[index]
	case consts.RoleLb:
		diskName = p.state.CloudInfra.Azure.InfoLoadBalancer.DiskName
	case consts.RoleDs:
		diskName = p.state.CloudInfra.Azure.InfoDatabase.DiskNames[index]
	}
	if len(diskName) == 0 {
		p.l.Print(p.ctx, "skipped disk already deleted")
		return nil
	}

	pollerResponse, err := p.client.BeginDeleteDisk(diskName, nil)
	if err != nil {
		return err
	}
	p.l.Print(p.ctx, "Deleting the disk..", "name", diskName)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelete error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		_, err = p.client.PollUntilDoneDelDisk(p.ctx, pollerResponse, nil)
		if err != nil {
			errDelete = err
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[index] = ""
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.DiskNames[index] = ""
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.DiskName = ""
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.DiskNames[index] = ""
		}
		if err := p.store.Write(p.state); err != nil {
			errDelete = err
			return
		}
	}()
	<-donePoll
	if errDelete != nil {
		return errDelete
	}

	p.l.Success(p.ctx, "Deleted disk", "name", diskName)
	return nil
}

func (p *Provider) CreatePublicIP(publicIPName string, index int, role consts.KsctlRole) error {

	publicIP := ""
	switch role {
	case consts.RoleWp:
		publicIP = p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index]
	case consts.RoleCp:
		publicIP = p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index]
	case consts.RoleLb:
		publicIP = p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPName
	case consts.RoleDs:
		publicIP = p.state.CloudInfra.Azure.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) != 0 {
		p.l.Print(p.ctx, "skipped pub ip already created", "name", publicIP)
		return nil
	}

	parameters := armnetwork.PublicIPAddress{
		Location: utilities.Ptr(p.Region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: utilities.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

	p.l.Debug(p.ctx, "Printing", "PublicIPConfig", parameters)

	pollerResponse, err := p.client.BeginCreatePubIP(publicIPName, parameters, nil)
	if err != nil {
		return err
	}

	// NOTE: Add the entry for name before polling starts so that state is present
	done := make(chan struct{})
	var errCreate error
	go func() {
		defer close(done)
		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index] = publicIPName
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index] = publicIPName
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPName = publicIPName
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPNames[index] = publicIPName
		}
		if err := p.store.Write(p.state); err != nil {
			errCreate = err
			return
		}
	}()
	<-done
	if errCreate != nil {
		return errCreate
	}
	p.l.Print(p.ctx, "creating the pubip..", "name", publicIPName)

	var errCreatePub error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		resp, err := p.client.PollUntilDoneCreatePubIP(p.ctx, pollerResponse, nil)
		if err != nil {
			errCreatePub = err
			return
		}

		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[index] = *resp.ID
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[index] = *resp.ID
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPID = *resp.ID
			p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIP = *resp.Properties.IPAddress
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPIDs[index] = *resp.ID
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPs[index] = *resp.Properties.IPAddress
		}

		if err := p.store.Write(p.state); err != nil {
			errCreatePub = err
			return
		}
	}()
	<-donePoll
	if errCreatePub != nil {
		return errCreatePub
	}

	p.l.Success(p.ctx, "Created public IP address", "name", publicIPName)
	return nil
}

func (p *Provider) DeletePublicIP(index int, role consts.KsctlRole) error {

	publicIP := ""
	switch role {
	case consts.RoleWp:
		publicIP = p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index]
	case consts.RoleCp:
		publicIP = p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index]
	case consts.RoleLb:
		publicIP = p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPName
	case consts.RoleDs:
		publicIP = p.state.CloudInfra.Azure.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) == 0 {
		p.l.Print(p.ctx, "skipped pub ip already deleted")
		return nil
	}

	pollerResponse, err := p.client.BeginDeletePubIP(publicIP, nil)
	if err != nil {
		return err
	}
	p.l.Print(p.ctx, "Deleting the pubip..", "name", publicIP)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelPub error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		_, err = p.client.PollUntilDoneDelPubIP(p.ctx, pollerResponse, nil)
		if err != nil {
			errDelPub = err
			return
		}

		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index] = ""
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[index] = ""
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[index] = ""
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index] = ""
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[index] = ""
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPs[index] = ""
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPID = ""
			p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIPName = ""
			p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIP = ""
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPNames[index] = ""
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPIDs[index] = ""
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPs[index] = ""
		}
		if err := p.store.Write(p.state); err != nil {
			errDelPub = err
			return
		}
	}()
	<-donePoll
	if errDelPub != nil {
		return errDelPub
	}

	p.l.Success(p.ctx, "Deleted the pub IP", "name", publicIP)
	return nil
}

func (p *Provider) CreateNetworkInterface(nicName string, subnetID string, publicIPID string, networkSecurityGroupID string, index int, role consts.KsctlRole) error {

	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case consts.RoleCp:
		interfaceName = p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index]
	case consts.RoleLb:
		interfaceName = p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) != 0 {
		p.l.Print(p.ctx, "skipped network interface already created", "name", interfaceName)
		return nil
	}

	parameters := armnetwork.Interface{
		Location: utilities.Ptr(p.Region),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: utilities.Ptr(p.state.CloudInfra.Azure.ResourceGroupName),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: utilities.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: utilities.Ptr(subnetID),
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: utilities.Ptr(publicIPID),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: utilities.Ptr(networkSecurityGroupID),
			},
		},
	}

	p.l.Debug(p.ctx, "Printing", "netInterfaceConfig", parameters)

	pollerResponse, err := p.client.BeginCreateNIC(nicName, parameters, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present
	done := make(chan struct{})
	var errCreate error
	go func() {
		defer close(done)
		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index] = nicName
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index] = nicName

		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName = nicName
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index] = nicName
		}
		if err := p.store.Write(p.state); err != nil {
			errCreate = err
			return
		}
	}()
	<-done
	if errCreate != nil {
		return errCreate
	}
	p.l.Print(p.ctx, "Creating the network interface...", "name", nicName)

	var errCreatenic error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		resp, err := p.client.PollUntilDoneCreateNetInterface(p.ctx, pollerResponse, nil)
		if err != nil {
			errCreatenic = err
			return
		}

		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *resp.ID
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[index] = *resp.ID
			p.state.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress

		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID = *resp.ID
			p.state.CloudInfra.Azure.InfoLoadBalancer.PrivateIP = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[index] = *resp.ID
			p.state.CloudInfra.Azure.InfoDatabase.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		}

		if err := p.store.Write(p.state); err != nil {
			errCreatenic = err
			return
		}
	}()
	<-donePoll
	if errCreatenic != nil {
		return errCreatenic
	}

	p.l.Success(p.ctx, "Created network interface", "name", nicName)
	return nil
}

func (p *Provider) DeleteNetworkInterface(index int, role consts.KsctlRole) error {
	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case consts.RoleCp:
		interfaceName = p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index]
	case consts.RoleLb:
		interfaceName = p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) == 0 {
		p.l.Print(p.ctx, "skipped network interface already deleted")
		return nil
	}

	pollerResponse, err := p.client.BeginDeleteNIC(interfaceName, nil)
	if err != nil {
		return err
	}
	p.l.Print(p.ctx, "Deleting the network interface...", "name", interfaceName)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelnic error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		_, err = p.client.PollUntilDoneDelNetInterface(p.ctx, pollerResponse, nil)
		if err != nil {
			errDelnic = err
			return
		}

		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index] = ""
			p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[index] = ""
		case consts.RoleCp:
			p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index] = ""
			p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
			p.state.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[index] = ""
		case consts.RoleLb:
			p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName = ""
			p.state.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID = ""
			p.state.CloudInfra.Azure.InfoLoadBalancer.PrivateIP = ""
		case consts.RoleDs:
			p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index] = ""
			p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[index] = ""
			p.state.CloudInfra.Azure.InfoDatabase.PrivateIPs[index] = ""
		}
		if err := p.store.Write(p.state); err != nil {
			errDelnic = err
			return
		}
	}()
	<-donePoll
	if errDelnic != nil {
		return errDelnic
	}

	p.l.Success(p.ctx, "Deleted the network interface", "name", interfaceName)

	return nil
}
