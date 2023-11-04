package azure

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// Sequence
// creation
// 1. PublicIP
// 2. Network Interface
// 3. Disk
// 4. VM

// DelVM implements resources.CloudFactory.
func (obj *AzureProvider) DelVM(storage resources.StorageFactory, index int) error {
	role := obj.metadata.role
	indexNo := index
	obj.mxRole.Unlock()

	log.Debug("Printing", "role", role, "indexNo", indexNo)

	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = azureCloudState.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = azureCloudState.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = azureCloudState.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = azureCloudState.InfoWorkerPlanes.Names[indexNo]
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
			obj.mxState.Lock()
			defer obj.mxState.Unlock()

			switch role {
			case consts.RoleWp:
				azureCloudState.InfoWorkerPlanes.Names[indexNo] = ""
				azureCloudState.InfoWorkerPlanes.Hostnames[indexNo] = ""
			case consts.RoleCp:
				azureCloudState.InfoControlPlanes.Names[indexNo] = ""
				azureCloudState.InfoControlPlanes.Hostnames[indexNo] = ""
			case consts.RoleLb:
				azureCloudState.InfoLoadBalancer.Name = ""
				azureCloudState.InfoLoadBalancer.HostName = ""
			case consts.RoleDs:
				azureCloudState.InfoDatabase.Names[indexNo] = ""
				azureCloudState.InfoDatabase.Hostnames[indexNo] = ""
			}

			if err := saveStateHelper(storage); err != nil {
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

	log.Debug("Printing", "azureCloudState", azureCloudState)

	return nil
}

// NewVM implements resources.CloudFactory.
func (obj *AzureProvider) NewVM(storage resources.StorageFactory, index int) error {
	name := obj.metadata.resName
	indexNo := index
	role := obj.metadata.role
	vmtype := obj.metadata.vmType
	obj.mxRole.Unlock()
	obj.mxName.Unlock()
	obj.mxVMType.Unlock()

	log.Debug("Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	if role == consts.RoleDs && indexNo > 0 {
		log.Print("skipped currently multiple datastore not supported")
		return nil
	}
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
		pubIPID = azureCloudState.InfoWorkerPlanes.PublicIPIDs[indexNo]
		nsgID = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID
	case consts.RoleCp:
		nsgID = azureCloudState.InfoControlPlanes.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoControlPlanes.PublicIPIDs[indexNo]
	case consts.RoleLb:
		nsgID = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoLoadBalancer.PublicIPID
	case consts.RoleDs:
		nsgID = azureCloudState.InfoDatabase.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoDatabase.PublicIPIDs[indexNo]
	}

	log.Debug("Printing", "PubIP_id", pubIPID, "NsgID", nsgID)

	if err := obj.CreateNetworkInterface(ctx, storage, nicName, azureCloudState.SubnetID, pubIPID, nsgID, indexNo, role); err != nil {
		return log.NewError(err.Error())
	}

	// NOTE: check if the VM is already created
	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = azureCloudState.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = azureCloudState.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = azureCloudState.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = azureCloudState.InfoWorkerPlanes.Names[indexNo]
	}
	if len(vmName) != 0 {
		log.Print("skipped vm already created", "name", vmName)
		return nil
	}

	sshPublicKeyPath := utils.GetPath(consts.UtilOtherPath, consts.CloudAzure, clusterType, clusterDirName, "keypair.pub")
	log.Debug("Printing", "sshKeyPath", sshPublicKeyPath)

	var sshBytes []byte
	_, err := os.Stat(sshPublicKeyPath)
	if err != nil {
		return log.NewError(err.Error())
	}
	sshBytes, err = storage.Path(sshPublicKeyPath).Load()
	if err != nil {
		return log.NewError(err.Error())
	}

	netInterfaceID := ""
	switch role {
	case consts.RoleWp:
		netInterfaceID = azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[indexNo]
	case consts.RoleCp:
		netInterfaceID = azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[indexNo]
	case consts.RoleLb:
		netInterfaceID = azureCloudState.InfoLoadBalancer.NetworkInterfaceID
	case consts.RoleDs:
		netInterfaceID = azureCloudState.InfoDatabase.NetworkInterfaceIDs[indexNo]
	}
	log.Debug("Printing", "netInterfaceID", netInterfaceID)

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
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr(name),
				AdminUsername: to.Ptr(azureCloudState.SSHUser),
				//CustomData:    to.Ptr(base64.StdEncoding.EncodeToString([]byte(script))),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr(string("/home/azureuser/.ssh/authorized_keys")),
								KeyData: to.Ptr(string(sshBytes)),
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
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.Names[indexNo] = name
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.Names[indexNo] = name
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.Name = name
		case consts.RoleDs:
			azureCloudState.InfoDatabase.Names[indexNo] = name
		}
		if err := saveStateHelper(storage); err != nil {
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
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.DiskNames[indexNo] = diskName
			azureCloudState.InfoWorkerPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName

			if len(azureCloudState.InfoWorkerPlanes.Names) == indexNo+1 {
				azureCloudState.IsCompleted = true
			}

		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.DiskNames[indexNo] = diskName
			azureCloudState.InfoControlPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
			if len(azureCloudState.InfoControlPlanes.Names) == indexNo+1 && len(azureCloudState.InfoWorkerPlanes.Names) == 0 {
				// when its the last resource to be created and we are done with the last controlplane creation
				azureCloudState.IsCompleted = true
			}
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.DiskName = diskName
			azureCloudState.InfoLoadBalancer.HostName = *resp.Properties.OSProfile.ComputerName
		case consts.RoleDs:
			azureCloudState.InfoDatabase.DiskNames[indexNo] = diskName
			azureCloudState.InfoDatabase.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
		}

		if err := saveStateHelper(storage); err != nil {
			errCreateVM = err
			return
		}

	}()
	<-donePoll
	if errCreateVM != nil {
		return log.NewError(errCreateVM.Error())
	}
	log.Debug("Printing", "azureCloudState", azureCloudState)

	log.Success("Created virtual machine", "name", name)
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {
	diskName := ""
	// pass the role
	switch role {
	case consts.RoleWp:
		diskName = azureCloudState.InfoWorkerPlanes.DiskNames[index]
	case consts.RoleCp:
		diskName = azureCloudState.InfoControlPlanes.DiskNames[index]
	case consts.RoleLb:
		diskName = azureCloudState.InfoLoadBalancer.DiskName
	case consts.RoleDs:
		diskName = azureCloudState.InfoDatabase.DiskNames[index]
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
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.DiskNames[index] = ""
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.DiskNames[index] = ""
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.DiskName = ""
		case consts.RoleDs:
			azureCloudState.InfoDatabase.DiskNames[index] = ""
		}
		if err := saveStateHelper(storage); err != nil {
			errDelete = err
			return
		}
	}()
	<-donePoll
	if errDelete != nil {
		return errDelete
	}

	log.Debug("Printing", "azureCloudState", azureCloudState)
	log.Success("Deleted disk", "name", diskName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int, role consts.KsctlRole) error {

	publicIP := ""
	switch role {
	case consts.RoleWp:
		publicIP = azureCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case consts.RoleCp:
		publicIP = azureCloudState.InfoControlPlanes.PublicIPNames[index]
	case consts.RoleLb:
		publicIP = azureCloudState.InfoLoadBalancer.PublicIPName
	case consts.RoleDs:
		publicIP = azureCloudState.InfoDatabase.PublicIPNames[index]
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
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.PublicIPNames[index] = publicIPName
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.PublicIPNames[index] = publicIPName
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.PublicIPName = publicIPName
		case consts.RoleDs:
			azureCloudState.InfoDatabase.PublicIPNames[index] = publicIPName
		}
		if err := saveStateHelper(storage); err != nil {
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

		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.PublicIPIDs[index] = *resp.ID
			azureCloudState.InfoWorkerPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.PublicIPIDs[index] = *resp.ID
			azureCloudState.InfoControlPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.PublicIPID = *resp.ID
			azureCloudState.InfoLoadBalancer.PublicIP = *resp.Properties.IPAddress
		case consts.RoleDs:
			azureCloudState.InfoDatabase.PublicIPIDs[index] = *resp.ID
			azureCloudState.InfoDatabase.PublicIPs[index] = *resp.Properties.IPAddress
		}

		if err := saveStateHelper(storage); err != nil {
			errCreatePub = err
			return
		}
	}()
	<-donePoll
	if errCreatePub != nil {
		return errCreatePub
	}

	log.Debug("Printing", "azureCloudState", azureCloudState)
	log.Success("Created public IP address", "name", publicIPName)
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {

	publicIP := ""
	switch role {
	case consts.RoleWp:
		publicIP = azureCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case consts.RoleCp:
		publicIP = azureCloudState.InfoControlPlanes.PublicIPNames[index]
	case consts.RoleLb:
		publicIP = azureCloudState.InfoLoadBalancer.PublicIPName
	case consts.RoleDs:
		publicIP = azureCloudState.InfoDatabase.PublicIPNames[index]
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

		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.PublicIPNames[index] = ""
			azureCloudState.InfoWorkerPlanes.PublicIPIDs[index] = ""
			azureCloudState.InfoWorkerPlanes.PublicIPs[index] = ""
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.PublicIPNames[index] = ""
			azureCloudState.InfoControlPlanes.PublicIPIDs[index] = ""
			azureCloudState.InfoControlPlanes.PublicIPs[index] = ""
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.PublicIPID = ""
			azureCloudState.InfoLoadBalancer.PublicIPName = ""
			azureCloudState.InfoLoadBalancer.PublicIP = ""
		case consts.RoleDs:
			azureCloudState.InfoDatabase.PublicIPNames[index] = ""
			azureCloudState.InfoDatabase.PublicIPIDs[index] = ""
			azureCloudState.InfoDatabase.PublicIPs[index] = ""
		}

		if err := saveStateHelper(storage); err != nil {
			errDelPub = err
			return
		}
	}()
	<-donePoll
	if errDelPub != nil {
		return errDelPub
	}

	log.Debug("Printing", "azureCloudState", azureCloudState)
	log.Success("Deleted the pub IP", "name", publicIP)
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory,
	nicName string, subnetID string, publicIPID string, networkSecurityGroupID string, index int, role consts.KsctlRole) error {

	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case consts.RoleCp:
		interfaceName = azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index]
	case consts.RoleLb:
		interfaceName = azureCloudState.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = azureCloudState.InfoDatabase.NetworkInterfaceNames[index]
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
					Name: to.Ptr(azureCloudState.ResourceGroupName),
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
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = nicName
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = nicName

		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.NetworkInterfaceName = nicName
		case consts.RoleDs:
			azureCloudState.InfoDatabase.NetworkInterfaceNames[index] = nicName
		}
		if err := saveStateHelper(storage); err != nil {
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

		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *resp.ID
			azureCloudState.InfoWorkerPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = *resp.ID
			azureCloudState.InfoControlPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress

		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.NetworkInterfaceID = *resp.ID
			azureCloudState.InfoLoadBalancer.PrivateIP = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case consts.RoleDs:
			azureCloudState.InfoDatabase.NetworkInterfaceIDs[index] = *resp.ID
			azureCloudState.InfoDatabase.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		}

		if err := saveStateHelper(storage); err != nil {
			errCreatenic = err
			return
		}
	}()
	<-donePoll
	if errCreatenic != nil {
		return errCreatenic
	}
	log.Debug("Printing", "azureCloudState", azureCloudState)

	log.Success("Created network interface", "name", nicName)
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {
	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case consts.RoleCp:
		interfaceName = azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index]
	case consts.RoleLb:
		interfaceName = azureCloudState.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = azureCloudState.InfoDatabase.NetworkInterfaceNames[index]
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

		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = ""
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
			azureCloudState.InfoWorkerPlanes.PrivateIPs[index] = ""
		case consts.RoleCp:
			azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = ""
			azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
			azureCloudState.InfoControlPlanes.PrivateIPs[index] = ""
		case consts.RoleLb:
			azureCloudState.InfoLoadBalancer.NetworkInterfaceName = ""
			azureCloudState.InfoLoadBalancer.NetworkInterfaceID = ""
			azureCloudState.InfoLoadBalancer.PrivateIP = ""
		case consts.RoleDs:
			azureCloudState.InfoDatabase.NetworkInterfaceNames[index] = ""
			azureCloudState.InfoDatabase.NetworkInterfaceIDs[index] = ""
			azureCloudState.InfoDatabase.PrivateIPs[index] = ""
		}
		if err := saveStateHelper(storage); err != nil {
			errDelnic = err
			return
		}
	}()
	<-donePoll
	if errDelnic != nil {
		return errDelnic
	}
	log.Debug("Printing", "azureCloudState", azureCloudState)

	log.Success("Deleted the network interface", "name", interfaceName)

	return nil
}
