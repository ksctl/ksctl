// Copyright 2024 ksctl
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
	"context"
	"encoding/base64"

	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
)

// NOTE: here we might need to define another ctx var for each function
// make sure that is passed instead of azureCtx

func (obj *AzureProvider) DelVM(storage types.StorageFactory, index int) error {
	role := <-obj.chRole
	indexNo := index

	log.Debug(azureCtx, "Printing", "role", role, "indexNo", indexNo)

	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo]
	}

	if len(vmName) == 0 {
		log.Print(azureCtx, "skipped vm already deleted")
	} else {

		var errDel error //just to make sure its nil
		donePoll := make(chan struct{})
		go func() {
			defer close(donePoll)
			pollerResponse, err := obj.client.BeginDeleteVM(vmName, nil)
			if err != nil {
				errDel = err
				return
			}
			log.Print(azureCtx, "deleting vm...", "name", vmName)

			_, err = obj.client.PollUntilDoneDelVM(azureCtx, pollerResponse, nil)
			if err != nil {
				errDel = err
				return
			}
			obj.mu.Lock()
			defer obj.mu.Unlock()

			switch role {
			case consts.RoleWp:
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes[indexNo] = ""
			case consts.RoleCp:
				mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Hostnames[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoControlPlanes.VMSizes[indexNo] = ""
			case consts.RoleLb:
				mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name = ""
				mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.HostName = ""
				mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.VMSize = ""
			case consts.RoleDs:
				mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoDatabase.Hostnames[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoDatabase.VMSizes[indexNo] = ""
			}

			if err := storage.Write(mainStateDocument); err != nil {
				errDel = err
				return
			}

		}()
		<-donePoll
		if errDel != nil {
			return errDel
		}
		log.Success(azureCtx, "Deleted the vm", "name", vmName)

	}

	if err := obj.DeleteDisk(azureCtx, storage, indexNo, role); err != nil {
		return err
	}

	if err := obj.DeleteNetworkInterface(azureCtx, storage, indexNo, role); err != nil {
		return err
	}

	if err := obj.DeletePublicIP(azureCtx, storage, indexNo, role); err != nil {
		return err
	}

	return nil
}

func (obj *AzureProvider) NewVM(storage types.StorageFactory, index int) error {
	name := <-obj.chResName
	indexNo := index
	role := <-obj.chRole
	vmtype := <-obj.chVMType

	log.Debug(azureCtx, "Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	pubIPName := name + "-pub"
	nicName := name + "-nic"
	diskName := name + "-disk"
	log.Debug(azureCtx, "Printing", "pubIPName", pubIPName, "NICName", nicName, "diskName", diskName)

	if err := obj.CreatePublicIP(azureCtx, storage, pubIPName, indexNo, role); err != nil {
		return err
	}

	pubIPID := ""
	nsgID := ""

	switch role {
	case consts.RoleWp:
		pubIPID = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[indexNo]
		nsgID = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID
	case consts.RoleCp:
		nsgID = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID
		pubIPID = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[indexNo]
	case consts.RoleLb:
		nsgID = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID
		pubIPID = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPID
	case consts.RoleDs:
		nsgID = mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID
		pubIPID = mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPIDs[indexNo]
	}

	log.Debug(azureCtx, "Printing", "PubIP_id", pubIPID, "NsgID", nsgID)

	if err := obj.CreateNetworkInterface(azureCtx, storage, nicName, mainStateDocument.CloudInfra.Azure.SubnetID, pubIPID, nsgID, indexNo, role); err != nil {
		return err
	}

	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo]
	}
	if len(vmName) != 0 {
		log.Print(azureCtx, "skipped vm already created", "name", vmName)
		return nil
	}

	netInterfaceID := ""
	switch role {
	case consts.RoleWp:
		netInterfaceID = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[indexNo]
	case consts.RoleCp:
		netInterfaceID = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[indexNo]
	case consts.RoleLb:
		netInterfaceID = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID
	case consts.RoleDs:
		netInterfaceID = mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[indexNo]
	}
	log.Debug(azureCtx, "Printing", "netInterfaceID", netInterfaceID)

	initScript, err := helpers.GenerateInitScriptForVM(name)
	if err != nil {
		return err
	}

	log.Debug(azureCtx, "initscript", "script", initScript)

	parameters := armcompute.VirtualMachine{
		Location: utilities.Ptr(obj.region),
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
				AdminUsername: utilities.Ptr(mainStateDocument.CloudInfra.Azure.B.SSHUser),
				CustomData:    utilities.Ptr(base64.StdEncoding.EncodeToString([]byte(initScript))),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: utilities.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    utilities.Ptr("/home/azureuser/.ssh/authorized_keys"),
								KeyData: utilities.Ptr(mainStateDocument.SSHKeyPair.PublicKey),
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
	log.Debug(azureCtx, "Printing", "VMConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateVM(name, parameters, nil)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	var errCreateVM error
	go func() {
		defer close(done)
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo] = name
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes[indexNo] = name
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[indexNo] = name
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.VMSizes[indexNo] = name
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name = name
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.VMSize = name
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[indexNo] = name
			mainStateDocument.CloudInfra.Azure.InfoDatabase.VMSizes[indexNo] = name
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreateVM = err
			return
		}

	}()

	<-done
	if errCreateVM != nil {
		return errCreateVM
	}
	log.Print(azureCtx, "creating vm...", "name", name)

	errCreateVM = nil //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)

		resp, err := obj.client.PollUntilDoneCreateVM(azureCtx, pollerResponse, nil)
		if err != nil {
			errCreateVM = err
			return
		}
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[indexNo] = diskName
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName

			if len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names) == indexNo+1 {
				mainStateDocument.CloudInfra.Azure.B.IsCompleted = true
			}

		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.DiskNames[indexNo] = diskName
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
			if len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names) == indexNo+1 && len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names) == 0 {
				// when its the last resource to be created and we are done with the last controlplane creation
				mainStateDocument.CloudInfra.Azure.B.IsCompleted = true
			}
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.DiskName = diskName
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.HostName = *resp.Properties.OSProfile.ComputerName
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.DiskNames[indexNo] = diskName
			mainStateDocument.CloudInfra.Azure.InfoDatabase.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
		}

		if err := storage.Write(mainStateDocument); err != nil {
			errCreateVM = err
			return
		}

	}()
	<-donePoll
	if errCreateVM != nil {
		return errCreateVM
	}

	log.Success(azureCtx, "Created virtual machine", "name", name)
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, storage types.StorageFactory, index int, role consts.KsctlRole) error {
	diskName := ""
	switch role {
	case consts.RoleWp:
		diskName = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[index]
	case consts.RoleCp:
		diskName = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.DiskNames[index]
	case consts.RoleLb:
		diskName = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.DiskName
	case consts.RoleDs:
		diskName = mainStateDocument.CloudInfra.Azure.InfoDatabase.DiskNames[index]
	}
	if len(diskName) == 0 {
		log.Print(azureCtx, "skipped disk already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteDisk(diskName, nil)
	if err != nil {
		return err
	}
	log.Print(azureCtx, "Deleting the disk..", "name", diskName)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelete error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		_, err = obj.client.PollUntilDoneDelDisk(ctx, pollerResponse, nil)
		if err != nil {
			errDelete = err
			return
		}
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[index] = ""
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.DiskNames[index] = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.DiskName = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.DiskNames[index] = ""
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errDelete = err
			return
		}
	}()
	<-donePoll
	if errDelete != nil {
		return errDelete
	}

	log.Success(azureCtx, "Deleted disk", "name", diskName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, storage types.StorageFactory, publicIPName string, index int, role consts.KsctlRole) error {

	publicIP := ""
	switch role {
	case consts.RoleWp:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index]
	case consts.RoleCp:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index]
	case consts.RoleLb:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPName
	case consts.RoleDs:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) != 0 {
		log.Print(azureCtx, "skipped pub ip already created", "name", publicIP)
		return nil
	}

	parameters := armnetwork.PublicIPAddress{
		Location: utilities.Ptr(obj.region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: utilities.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

	log.Debug(azureCtx, "Printing", "PublicIPConfig", parameters)

	pollerResponse, err := obj.client.BeginCreatePubIP(publicIPName, parameters, nil)
	if err != nil {
		return err
	}

	// NOTE: Add the entry for name before polling starts so that state is present
	done := make(chan struct{})
	var errCreate error
	go func() {
		defer close(done)
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index] = publicIPName
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index] = publicIPName
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPName = publicIPName
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames[index] = publicIPName
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreate = err
			return
		}
	}()
	<-done
	if errCreate != nil {
		return errCreate
	}
	log.Print(azureCtx, "creating the pubip..", "name", publicIPName)

	var errCreatePub error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		resp, err := obj.client.PollUntilDoneCreatePubIP(ctx, pollerResponse, nil)
		if err != nil {
			errCreatePub = err
			return
		}

		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[index] = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[index] = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPID = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIP = *resp.Properties.IPAddress
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPIDs[index] = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs[index] = *resp.Properties.IPAddress
		}

		if err := storage.Write(mainStateDocument); err != nil {
			errCreatePub = err
			return
		}
	}()
	<-donePoll
	if errCreatePub != nil {
		return errCreatePub
	}

	log.Success(azureCtx, "Created public IP address", "name", publicIPName)
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, storage types.StorageFactory, index int, role consts.KsctlRole) error {

	publicIP := ""
	switch role {
	case consts.RoleWp:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index]
	case consts.RoleCp:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index]
	case consts.RoleLb:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPName
	case consts.RoleDs:
		publicIP = mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) == 0 {
		log.Print(azureCtx, "skipped pub ip already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeletePubIP(publicIP, nil)
	if err != nil {
		return err
	}
	log.Print(azureCtx, "Deleting the pubip..", "name", publicIP)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelPub error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		_, err = obj.client.PollUntilDoneDelPubIP(ctx, pollerResponse, nil)
		if err != nil {
			errDelPub = err
			return
		}

		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[index] = ""
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs[index] = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPID = ""
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIPName = ""
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIP = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPIDs[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs[index] = ""
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errDelPub = err
			return
		}
	}()
	<-donePoll
	if errDelPub != nil {
		return errDelPub
	}

	log.Success(azureCtx, "Deleted the pub IP", "name", publicIP)
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, storage types.StorageFactory,
	nicName string, subnetID string, publicIPID string, networkSecurityGroupID string, index int, role consts.KsctlRole) error {

	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case consts.RoleCp:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index]
	case consts.RoleLb:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) != 0 {
		log.Print(azureCtx, "skipped network interface already created", "name", interfaceName)
		return nil
	}

	parameters := armnetwork.Interface{
		Location: utilities.Ptr(obj.region),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: utilities.Ptr(mainStateDocument.CloudInfra.Azure.ResourceGroupName),
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

	log.Debug(azureCtx, "Printing", "netInterfaceConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateNIC(nicName, parameters, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present
	done := make(chan struct{})
	var errCreate error
	go func() {
		defer close(done)
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index] = nicName
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index] = nicName

		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName = nicName
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index] = nicName
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreate = err
			return
		}
	}()
	<-done
	if errCreate != nil {
		return errCreate
	}
	log.Print(azureCtx, "Creating the network interface...", "name", nicName)

	var errCreatenic error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		resp, err := obj.client.PollUntilDoneCreateNetInterface(ctx, pollerResponse, nil)
		if err != nil {
			errCreatenic = err
			return
		}

		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[index] = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress

		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PrivateIP = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[index] = *resp.ID
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		}

		if err := storage.Write(mainStateDocument); err != nil {
			errCreatenic = err
			return
		}
	}()
	<-donePoll
	if errCreatenic != nil {
		return errCreatenic
	}

	log.Success(azureCtx, "Created network interface", "name", nicName)
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, storage types.StorageFactory, index int, role consts.KsctlRole) error {
	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case consts.RoleCp:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index]
	case consts.RoleLb:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) == 0 {
		log.Print(azureCtx, "skipped network interface already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteNIC(interfaceName, nil)
	if err != nil {
		return err
	}
	log.Print(azureCtx, "Deleting the network interface...", "name", interfaceName)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelnic error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)
		_, err = obj.client.PollUntilDoneDelNetInterface(ctx, pollerResponse, nil)
		if err != nil {
			errDelnic = err
			return
		}

		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[index] = ""
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs[index] = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceName = ""
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.NetworkInterfaceID = ""
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PrivateIP = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs[index] = ""
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs[index] = ""
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errDelnic = err
			return
		}
	}()
	<-donePoll
	if errDelnic != nil {
		return errDelnic
	}

	log.Success(azureCtx, "Deleted the network interface", "name", interfaceName)

	return nil
}
