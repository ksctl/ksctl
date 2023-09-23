package azure

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
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

	vmName := ""
	switch role {
	case ROLE_CP:
		vmName = azureCloudState.InfoControlPlanes.Names[indexNo]
	case ROLE_DS:
		vmName = azureCloudState.InfoDatabase.Names[indexNo]
	case ROLE_LB:
		vmName = azureCloudState.InfoLoadBalancer.Name
	case ROLE_WP:
		vmName = azureCloudState.InfoWorkerPlanes.Names[indexNo]
	}

	if len(vmName) == 0 {
		storage.Logger().Success("[skip] vm already deleted")
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
			storage.Logger().Print("[azure] deleting vm...", vmName)

			_, err = obj.client.PollUntilDoneDelVM(ctx, pollerResponse, nil)
			if err != nil {
				errDel = err
				return
			}
			obj.mxState.Lock()
			defer obj.mxState.Unlock()

			switch role {
			case ROLE_WP:
				azureCloudState.InfoWorkerPlanes.Names[indexNo] = ""
				azureCloudState.InfoWorkerPlanes.Hostnames[indexNo] = ""
			case ROLE_CP:
				azureCloudState.InfoControlPlanes.Names[indexNo] = ""
				azureCloudState.InfoControlPlanes.Hostnames[indexNo] = ""
			case ROLE_LB:
				azureCloudState.InfoLoadBalancer.Name = ""
				azureCloudState.InfoLoadBalancer.HostName = ""
			case ROLE_DS:
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
			return errDel
		}
		storage.Logger().Success("[azure] Deleted the vm", vmName)

	}

	if err := obj.DeleteDisk(ctx, storage, indexNo, role); err != nil {
		return err
	}

	if err := obj.DeleteNetworkInterface(ctx, storage, indexNo, role); err != nil {
		return err
	}

	if err := obj.DeletePublicIP(ctx, storage, indexNo, role); err != nil {
		return err
	}

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

	if role == ROLE_DS && indexNo > 0 {
		storage.Logger().Note("[skip] currently multiple datastore not supported")
		return nil
	}
	pubIPName := name + "-pub"
	nicName := name + "-nic"
	diskName := name + "-disk"
	if err := obj.CreatePublicIP(ctx, storage, pubIPName, indexNo, role); err != nil {
		return err
	}

	pubIPID := ""
	nsgID := ""

	switch role {
	case ROLE_WP:
		pubIPID = azureCloudState.InfoWorkerPlanes.PublicIPIDs[indexNo]
		nsgID = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID
	case ROLE_CP:
		nsgID = azureCloudState.InfoControlPlanes.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoControlPlanes.PublicIPIDs[indexNo]
	case ROLE_LB:
		nsgID = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoLoadBalancer.PublicIPID
	case ROLE_DS:
		nsgID = azureCloudState.InfoDatabase.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoDatabase.PublicIPIDs[indexNo]
	}

	if err := obj.CreateNetworkInterface(ctx, storage, nicName, azureCloudState.SubnetID, pubIPID, nsgID, indexNo, role); err != nil {
		return err
	}

	// NOTE: check if the VM is already created
	vmName := ""
	switch role {
	case ROLE_CP:
		vmName = azureCloudState.InfoControlPlanes.Names[indexNo]
	case ROLE_DS:
		vmName = azureCloudState.InfoDatabase.Names[indexNo]
	case ROLE_LB:
		vmName = azureCloudState.InfoLoadBalancer.Name
	case ROLE_WP:
		vmName = azureCloudState.InfoWorkerPlanes.Names[indexNo]
	}
	if len(vmName) != 0 {
		storage.Logger().Success("[skip] vm already created", vmName)
		return nil
	}

	sshPublicKeyPath := utils.GetPath(OTHER_PATH, CLOUD_AZURE, clusterType, clusterDirName, "keypair.pub")
	var sshBytes []byte
	_, err := os.Stat(sshPublicKeyPath)
	if err != nil {
		return err
	}
	sshBytes, err = storage.Path(sshPublicKeyPath).Load()
	if err != nil {
		return err
	}

	netInterfaceID := ""
	switch role {
	case ROLE_WP:
		netInterfaceID = azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[indexNo]
	case ROLE_CP:
		netInterfaceID = azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[indexNo]
	case ROLE_LB:
		netInterfaceID = azureCloudState.InfoLoadBalancer.NetworkInterfaceID
	case ROLE_DS:
		netInterfaceID = azureCloudState.InfoDatabase.NetworkInterfaceIDs[indexNo]
	}

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
	pollerResponse, err := obj.client.BeginCreateVM(name, parameters, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present
	done := make(chan struct{})
	var errCreateVM error
	go func() {
		defer close(done)
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.Names[indexNo] = name
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.Names[indexNo] = name
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.Name = name
		case ROLE_DS:
			azureCloudState.InfoDatabase.Names[indexNo] = name
		}
		if err := saveStateHelper(storage); err != nil {
			errCreateVM = err
			return
		}

	}()

	<-done
	if errCreateVM != nil {
		return errCreateVM
	}
	storage.Logger().Print("[azure] creating vm...", name)

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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.DiskNames[indexNo] = diskName
			azureCloudState.InfoWorkerPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName

			if len(azureCloudState.InfoWorkerPlanes.Names) == indexNo+1 {
				azureCloudState.IsCompleted = true
			}
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.DiskNames[indexNo] = diskName
			azureCloudState.InfoControlPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
			if len(azureCloudState.InfoControlPlanes.Names) == indexNo+1 && len(azureCloudState.InfoWorkerPlanes.Names) == 0 {
				// when its the last resource to be created and we are done with the last controlplane creation
				azureCloudState.IsCompleted = true
			}
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.DiskName = diskName
			azureCloudState.InfoLoadBalancer.HostName = *resp.Properties.OSProfile.ComputerName
		case ROLE_DS:
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
		return errCreateVM
	}
	storage.Logger().Success("[azure] Created virtual machine", name)
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int, role KsctlRole) error {
	diskName := ""
	// pass the role
	switch role {
	case ROLE_WP:
		diskName = azureCloudState.InfoWorkerPlanes.DiskNames[index]
	case ROLE_CP:
		diskName = azureCloudState.InfoControlPlanes.DiskNames[index]
	case ROLE_LB:
		diskName = azureCloudState.InfoLoadBalancer.DiskName
	case ROLE_DS:
		diskName = azureCloudState.InfoDatabase.DiskNames[index]
	}
	if len(diskName) == 0 {
		storage.Logger().Success("[skip] disk already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteDisk(diskName, nil)
	if err != nil {
		return err
	}
	storage.Logger().Print("[azure] Deleting the disk..", diskName)

	// NOTE: Add the entry for name before polling starts so that state is present

	var errDelete error //just to make sure its nil
	donePoll := make(chan struct{})
	go func() {
		close(donePoll)
		_, err = obj.client.PollUntilDoneDelDisk(ctx, pollerResponse, nil)
		if err != nil {
			errDelete = err
			return
		}
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.DiskNames[index] = ""
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.DiskNames[index] = ""
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.DiskName = ""
		case ROLE_DS:
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
	storage.Logger().Success("[azure] Deleted disk", diskName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int, role KsctlRole) error {

	publicIP := ""
	switch role {
	case ROLE_WP:
		publicIP = azureCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case ROLE_CP:
		publicIP = azureCloudState.InfoControlPlanes.PublicIPNames[index]
	case ROLE_LB:
		publicIP = azureCloudState.InfoLoadBalancer.PublicIPName
	case ROLE_DS:
		publicIP = azureCloudState.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) != 0 {
		storage.Logger().Success("[skip] pub ip already created", publicIP)
		return nil
	}

	parameters := armnetwork.PublicIPAddress{
		Location: to.Ptr(obj.region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.PublicIPNames[index] = publicIPName
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.PublicIPNames[index] = publicIPName
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.PublicIPName = publicIPName
		case ROLE_DS:
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
	storage.Logger().Print("[azure] creating the pubip..", publicIPName)

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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.PublicIPIDs[index] = *resp.ID
			azureCloudState.InfoWorkerPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.PublicIPIDs[index] = *resp.ID
			azureCloudState.InfoControlPlanes.PublicIPs[index] = *resp.Properties.IPAddress
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.PublicIPID = *resp.ID
			azureCloudState.InfoLoadBalancer.PublicIP = *resp.Properties.IPAddress
		case ROLE_DS:
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
	storage.Logger().Success("[azure] Created public IP address", publicIPName)
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int, role KsctlRole) error {

	publicIP := ""
	switch role {
	case ROLE_WP:
		publicIP = azureCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case ROLE_CP:
		publicIP = azureCloudState.InfoControlPlanes.PublicIPNames[index]
	case ROLE_LB:
		publicIP = azureCloudState.InfoLoadBalancer.PublicIPName
	case ROLE_DS:
		publicIP = azureCloudState.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) == 0 {
		storage.Logger().Success("[skip] pub ip already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeletePubIP(publicIP, nil)
	if err != nil {
		return err
	}
	storage.Logger().Print("[azure] Deleting the pubip..", publicIP)

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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.PublicIPNames[index] = ""
			azureCloudState.InfoWorkerPlanes.PublicIPIDs[index] = ""
			azureCloudState.InfoWorkerPlanes.PublicIPs[index] = ""
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.PublicIPNames[index] = ""
			azureCloudState.InfoControlPlanes.PublicIPIDs[index] = ""
			azureCloudState.InfoControlPlanes.PublicIPs[index] = ""
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.PublicIPID = ""
			azureCloudState.InfoLoadBalancer.PublicIPName = ""
			azureCloudState.InfoLoadBalancer.PublicIP = ""
		case ROLE_DS:
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
	storage.Logger().Success("[azure] Deleted the pub IP", publicIP)
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory,
	nicName string, subnetID string, publicIPID string, networkSecurityGroupID string, index int, role KsctlRole) error {

	interfaceName := ""
	switch role {
	case ROLE_WP:
		interfaceName = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case ROLE_CP:
		interfaceName = azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index]
	case ROLE_LB:
		interfaceName = azureCloudState.InfoLoadBalancer.NetworkInterfaceName
	case ROLE_DS:
		interfaceName = azureCloudState.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) != 0 {
		storage.Logger().Success("[skip] network interface already created", interfaceName)
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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = nicName
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = nicName

		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.NetworkInterfaceName = nicName
		case ROLE_DS:
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
	storage.Logger().Print("[azure] Creating the network interface...", nicName)

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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *resp.ID
			azureCloudState.InfoWorkerPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = *resp.ID
			azureCloudState.InfoControlPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress

		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.NetworkInterfaceID = *resp.ID
			azureCloudState.InfoLoadBalancer.PrivateIP = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
		case ROLE_DS:
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
	storage.Logger().Success("[azure] Created network interface", nicName)
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role KsctlRole) error {
	interfaceName := ""
	switch role {
	case ROLE_WP:
		interfaceName = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case ROLE_CP:
		interfaceName = azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index]
	case ROLE_LB:
		interfaceName = azureCloudState.InfoLoadBalancer.NetworkInterfaceName
	case ROLE_DS:
		interfaceName = azureCloudState.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) == 0 {
		storage.Logger().Success("[skip] network interface already deleted")
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteNIC(interfaceName, nil)
	if err != nil {
		return err
	}
	storage.Logger().Print("[azure] Deleting the network interface...", interfaceName)

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
		case ROLE_WP:
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = ""
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
			azureCloudState.InfoWorkerPlanes.PrivateIPs[index] = ""
		case ROLE_CP:
			azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = ""
			azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
			azureCloudState.InfoControlPlanes.PrivateIPs[index] = ""
		case ROLE_LB:
			azureCloudState.InfoLoadBalancer.NetworkInterfaceName = ""
			azureCloudState.InfoLoadBalancer.NetworkInterfaceID = ""
			azureCloudState.InfoLoadBalancer.PrivateIP = ""
		case ROLE_DS:
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
	storage.Logger().Success("[azure] Deleted the network interface", interfaceName)

	return nil
}
