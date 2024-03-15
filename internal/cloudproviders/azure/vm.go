package azure

import (
	"context"
	"encoding/base64"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// Sequence
// creation
// 1. PublicIP
// 2. Network Interface
// 3. Disk
// 4. VM

// DelVM implements resources.CloudFactory.
func (obj *AzureProvider) DelVM(storage resources.StorageFactory, index int) error {
	role := <-obj.chRole
	indexNo := index

	log.Debug("Printing", "role", role, "indexNo", indexNo)

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
		log.Print("skipped vm already deleted")
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
			log.Print("deleting vm...", "name", vmName)

			_, err = obj.client.PollUntilDoneDelVM(ctx, pollerResponse, nil)
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
			case consts.RoleCp:
				mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Hostnames[indexNo] = ""
			case consts.RoleLb:
				mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name = ""
				mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.HostName = ""
			case consts.RoleDs:
				mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[indexNo] = ""
				mainStateDocument.CloudInfra.Azure.InfoDatabase.Hostnames[indexNo] = ""
			}

			if err := storage.Write(mainStateDocument); err != nil {
				errDel = err
				return
			}

		}()
		<-donePoll
		if errDel != nil {
			return log.NewError(errDel.Error())
		}
		log.Success("Deleted the vm", "name", vmName)

	}

	if err := obj.DeleteDisk(ctx, storage, indexNo, role); err != nil {
		return log.NewError(err.Error())
	}

	if err := obj.DeleteNetworkInterface(ctx, storage, indexNo, role); err != nil {
		return log.NewError(err.Error())
	}

	if err := obj.DeletePublicIP(ctx, storage, indexNo, role); err != nil {
		return log.NewError(err.Error())
	}

	log.Debug("Printing", "mainStateDocument", mainStateDocument)

	return nil
}

// NewVM implements resources.CloudFactory.
func (obj *AzureProvider) NewVM(storage resources.StorageFactory, index int) error {
	name := <-obj.chResName
	indexNo := index
	role := <-obj.chRole
	vmtype := <-obj.chVMType

	log.Debug("Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	//if role == consts.RoleDs && indexNo > 0 {
	//	log.Print("skipped currently multiple datastore not supported")
	//	return nil
	//}
	pubIPName := name + "-pub"
	nicName := name + "-nic"
	diskName := name + "-disk"
	log.Debug("Printing", "pubIPName", pubIPName, "NICName", nicName, "diskName", diskName)

	if err := obj.CreatePublicIP(ctx, storage, pubIPName, indexNo, role); err != nil {
		return log.NewError(err.Error())
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

	log.Debug("Printing", "PubIP_id", pubIPID, "NsgID", nsgID)

	if err := obj.CreateNetworkInterface(ctx, storage, nicName, mainStateDocument.CloudInfra.Azure.SubnetID, pubIPID, nsgID, indexNo, role); err != nil {
		return log.NewError(err.Error())
	}

	// NOTE: check if the VM is already created
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
		log.Print("skipped vm already created", "name", vmName)
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
	log.Debug("Printing", "netInterfaceID", netInterfaceID)

	initScript, err := helpers.GenerateInitScriptForVM(name)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Debug("initscript", "script", initScript)

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(obj.region),
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     to.Ptr("0001-com-ubuntu-server-jammy"),
					Publisher: to.Ptr("Canonical"),
					SKU:       to.Ptr("22_04-lts-gen2"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(diskName),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS), // OSDisk type Standard/Premium HDD/SSD
					},
					//DiskSizeGB: to.Ptr[int32](100), // default 127G
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(vmtype)), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  to.Ptr(name),
				AdminUsername: to.Ptr(mainStateDocument.CloudInfra.Azure.B.SSHUser),
				CustomData:    to.Ptr(base64.StdEncoding.EncodeToString([]byte(initScript))),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr("/home/azureuser/.ssh/authorized_keys"),
								KeyData: to.Ptr(mainStateDocument.SSHKeyPair.PublicKey),
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.Ptr(netInterfaceID),
					},
				},
			},
		},
	}
	log.Debug("Printing", "VMConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateVM(name, parameters, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	// NOTE: Add the entry for name before polling starts so that state is present
	done := make(chan struct{})
	var errCreateVM error
	go func() {
		defer close(done)
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[indexNo] = name
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names[indexNo] = name
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.Name = name
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Azure.InfoDatabase.Names[indexNo] = name
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreateVM = err
			return
		}

	}()

	<-done
	if errCreateVM != nil {
		return log.NewError(errCreateVM.Error())
	}
	log.Print("creating vm...", "name", name)

	errCreateVM = nil //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		defer close(donePoll)

		resp, err := obj.client.PollUntilDoneCreateVM(ctx, pollerResponse, nil)
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
		return log.NewError(errCreateVM.Error())
	}
	log.Debug("Printing", "mainStateDocument", mainStateDocument)

	log.Success("Created virtual machine", "name", name)
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {
	diskName := ""
	// pass the role
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
		log.Print("skipped disk already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteDisk(diskName, nil)
	if err != nil {
		return err
	}
	log.Print("Deleting the disk..", "name", diskName)

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

	log.Debug("Printing", "mainStateDocument", mainStateDocument)
	log.Success("Deleted disk", "name", diskName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int, role consts.KsctlRole) error {

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
		log.Print("skipped pub ip already created", "name", publicIP)
		return nil
	}

	parameters := armnetwork.PublicIPAddress{
		Location: to.Ptr(obj.region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

	log.Debug("Printing", "PublicIPConfig", parameters)

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
	log.Print("creating the pubip..", "name", publicIPName)

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

	log.Debug("Printing", "mainStateDocument", mainStateDocument)
	log.Success("Created public IP address", "name", publicIPName)
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {

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
		log.Print("skipped pub ip already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeletePubIP(publicIP, nil)
	if err != nil {
		return err
	}
	log.Print("Deleting the pubip..", "name", publicIP)

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

	log.Debug("Printing", "mainStateDocument", mainStateDocument)
	log.Success("Deleted the pub IP", "name", publicIP)
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory,
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
		log.Print("skipped network interface already created", "name", interfaceName)
		return nil
	}

	parameters := armnetwork.Interface{
		Location: to.Ptr(obj.region),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr(mainStateDocument.CloudInfra.Azure.ResourceGroupName),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: to.Ptr(subnetID),
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: to.Ptr(publicIPID),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(networkSecurityGroupID),
			},
		},
	}

	log.Debug("Printing", "netInterfaceConfig", parameters)

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
	log.Print("Creating the network interface...", "name", nicName)

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
	log.Debug("Printing", "mainStateDocument", mainStateDocument)

	log.Success("Created network interface", "name", nicName)
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {
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
		log.Print("skipped network interface already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteNIC(interfaceName, nil)
	if err != nil {
		return err
	}
	log.Print("Deleting the network interface...", "name", interfaceName)

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
	log.Debug("Printing", "mainStateDocument", mainStateDocument)

	log.Success("Deleted the network interface", "name", interfaceName)

	return nil
}
