package azure

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// Sequence
// creation
// 1. PublicIP
// 2. Network Interface
// 3. Disk
// 4. VM
//
// for deletion it reverse

// DelVM implements resources.CloudFactory.
func (obj *AzureProvider) DelVM(storage resources.StorageFactory, indexNo int) error {

	vmName := ""
	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		vmName = azureCloudState.InfoControlPlanes.Names[indexNo]
	case utils.ROLE_DS:
		vmName = azureCloudState.InfoDatabase.Names[indexNo]
	case utils.ROLE_LB:
		vmName = azureCloudState.InfoLoadBalancer.Name
	case utils.ROLE_WP:
		vmName = azureCloudState.InfoWorkerPlanes.Names[indexNo]
	}

	if len(vmName) == 0 {
		storage.Logger().Success("[skip] vm already deleted")
	} else {
		vmClient, err := obj.Client.VirtualMachineClient()
		if err != nil {
			return err
		}

		pollerResponse, err := obj.Client.BeginDeleteVM(ctx, vmClient, azureCloudState.ResourceGroupName, vmName, nil)
		if err != nil {
			return err
		}
		storage.Logger().Print("[azure] deleting vm...", vmName)

		_, err = obj.Client.PollUntilDoneDelVM(ctx, pollerResponse, nil)
		if err != nil {
			return err
		}

		switch obj.Metadata.Role {
		case utils.ROLE_WP:
			azureCloudState.InfoWorkerPlanes.Names[indexNo] = ""
			azureCloudState.InfoWorkerPlanes.Hostnames[indexNo] = ""
		case utils.ROLE_CP:
			azureCloudState.InfoControlPlanes.Names[indexNo] = ""
			azureCloudState.InfoControlPlanes.Hostnames[indexNo] = ""
		case utils.ROLE_LB:
			azureCloudState.InfoLoadBalancer.Name = ""
			azureCloudState.InfoLoadBalancer.HostName = ""
		case utils.ROLE_DS:
			azureCloudState.InfoDatabase.Names[indexNo] = ""
			azureCloudState.InfoDatabase.Hostnames[indexNo] = ""
		}

		if err := saveStateHelper(storage); err != nil {
			return err
		}

		storage.Logger().Success("[azure] Deleted the vm", vmName)
	}

	if err := obj.DeleteDisk(ctx, storage, indexNo); err != nil {
		return err
	}

	if err := obj.DeleteNetworkInterface(ctx, storage, indexNo); err != nil {
		return err
	}

	if err := obj.DeletePublicIP(ctx, storage, indexNo); err != nil {
		return err
	}

	return nil
}

// NewVM implements resources.CloudFactory.
func (obj *AzureProvider) NewVM(storage resources.StorageFactory, indexNo int) error {
	if obj.Metadata.Role == utils.ROLE_DS && indexNo > 0 {
		storage.Logger().Note("[skip] currently multiple datastore not supported")
		return nil
	}
	pubIPName := obj.Metadata.ResName + "-pub"
	nicName := obj.Metadata.ResName + "-nic"
	diskName := obj.Metadata.ResName + "-disk"
	if err := obj.CreatePublicIP(ctx, storage, pubIPName, indexNo); err != nil {
		return err
	}

	pubIPID := ""
	nsgID := ""

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		pubIPID = azureCloudState.InfoWorkerPlanes.PublicIPIDs[indexNo]
		nsgID = azureCloudState.InfoWorkerPlanes.NetworkSecurityGroupID
	case utils.ROLE_CP:
		nsgID = azureCloudState.InfoControlPlanes.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoControlPlanes.PublicIPIDs[indexNo]
	case utils.ROLE_LB:
		nsgID = azureCloudState.InfoLoadBalancer.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoLoadBalancer.PublicIPID
	case utils.ROLE_DS:
		nsgID = azureCloudState.InfoDatabase.NetworkSecurityGroupID
		pubIPID = azureCloudState.InfoDatabase.PublicIPIDs[indexNo]
	}

	if err := obj.CreateNetworkInterface(ctx, storage, nicName, azureCloudState.SubnetID, pubIPID, nsgID, indexNo); err != nil {
		return err
	}

	// NOTE: check if the VM is already created
	vmName := ""
	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		vmName = azureCloudState.InfoControlPlanes.Names[indexNo]
	case utils.ROLE_DS:
		vmName = azureCloudState.InfoDatabase.Names[indexNo]
	case utils.ROLE_LB:
		vmName = azureCloudState.InfoLoadBalancer.Name
	case utils.ROLE_WP:
		vmName = azureCloudState.InfoWorkerPlanes.Names[indexNo]
	}
	if len(vmName) != 0 {
		storage.Logger().Success("[skip] vm already created", vmName)
		return nil
	}

	vmClient, err := obj.Client.VirtualMachineClient()
	if err != nil {
		return err
	}
	sshPublicKeyPath := utils.GetPath(utils.OTHER_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName, "keypair.pub")
	var sshBytes []byte
	_, err = os.Stat(sshPublicKeyPath)
	if err != nil {
		return err
	}
	sshBytes, err = storage.Path(sshPublicKeyPath).Load()
	if err != nil {
		return err
	}

	netInterfaceID := ""
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		netInterfaceID = azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[indexNo]
	case utils.ROLE_CP:
		netInterfaceID = azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[indexNo]
	case utils.ROLE_LB:
		netInterfaceID = azureCloudState.InfoLoadBalancer.NetworkInterfaceID
	case utils.ROLE_DS:
		netInterfaceID = azureCloudState.InfoDatabase.NetworkInterfaceIDs[indexNo]
	}

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(obj.Region),
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
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(obj.Metadata.VmType)), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr(obj.Metadata.ResName),
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
	pollerResponse, err := obj.Client.BeginCreateVM(ctx, vmClient, obj.ResourceGroup, obj.Metadata.ResName, parameters, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.Names[indexNo] = obj.Metadata.ResName
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.Names[indexNo] = obj.Metadata.ResName
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.Name = obj.Metadata.ResName
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.Names[indexNo] = obj.Metadata.ResName
	}
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Print("[azure] creating vm...", obj.Metadata.ResName)
	resp, err := obj.Client.PollUntilDoneCreateVM(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.DiskNames[indexNo] = diskName
		azureCloudState.InfoWorkerPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.DiskNames[indexNo] = diskName
		azureCloudState.InfoControlPlanes.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.DiskName = diskName
		azureCloudState.InfoLoadBalancer.HostName = *resp.Properties.OSProfile.ComputerName
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.DiskNames[indexNo] = diskName
		azureCloudState.InfoDatabase.Hostnames[indexNo] = *resp.Properties.OSProfile.ComputerName
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[azure] Created virtual machine", *resp.Name)
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int) error {
	diskName := ""
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		diskName = azureCloudState.InfoWorkerPlanes.DiskNames[index]
	case utils.ROLE_CP:
		diskName = azureCloudState.InfoControlPlanes.DiskNames[index]
	case utils.ROLE_LB:
		diskName = azureCloudState.InfoLoadBalancer.DiskName
	case utils.ROLE_DS:
		diskName = azureCloudState.InfoDatabase.DiskNames[index]
	}
	if len(diskName) == 0 {
		storage.Logger().Success("[skip] disk already deleted")
		return nil
	}
	diskClient, err := obj.Client.DiskClient()
	if err != nil {
		return err
	}

	pollerResponse, err := obj.Client.BeginDeleteDisk(ctx, diskClient, obj.ResourceGroup, diskName, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present

	_, err = obj.Client.PollUntilDoneDelDisk(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.DiskNames[index] = ""
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.DiskNames[index] = ""
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.DiskName = ""
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.DiskNames[index] = ""
	}
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Deleted disk", diskName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int) error {

	publicIP := ""
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		publicIP = azureCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case utils.ROLE_CP:
		publicIP = azureCloudState.InfoControlPlanes.PublicIPNames[index]
	case utils.ROLE_LB:
		publicIP = azureCloudState.InfoLoadBalancer.PublicIPName
	case utils.ROLE_DS:
		publicIP = azureCloudState.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) != 0 {
		storage.Logger().Success("[skip] pub ip already created", publicIP)
		return nil
	}
	publicIPAddressClient, err := obj.Client.PublicIPClient()
	if err != nil {
		return err
	}

	parameters := armnetwork.PublicIPAddress{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

	pollerResponse, err := obj.Client.BeginCreatePubIP(ctx, publicIPAddressClient, azureCloudState.ResourceGroupName, publicIPName, parameters, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.PublicIPNames[index] = publicIPName
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.PublicIPNames[index] = publicIPName
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.PublicIPName = publicIPName
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.PublicIPNames[index] = publicIPName
	}
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	resp, err := obj.Client.PollUntilDoneCreatePubIP(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.PublicIPIDs[index] = *resp.ID
		azureCloudState.InfoWorkerPlanes.PublicIPs[index] = *resp.Properties.IPAddress
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.PublicIPIDs[index] = *resp.ID
		azureCloudState.InfoControlPlanes.PublicIPs[index] = *resp.Properties.IPAddress
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.PublicIPID = *resp.ID
		azureCloudState.InfoLoadBalancer.PublicIP = *resp.Properties.IPAddress
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.PublicIPIDs[index] = *resp.ID
		azureCloudState.InfoDatabase.PublicIPs[index] = *resp.Properties.IPAddress
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[azure] Created public IP address", *resp.Name)
	return nil
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int) error {

	publicIP := ""
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		publicIP = azureCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case utils.ROLE_CP:
		publicIP = azureCloudState.InfoControlPlanes.PublicIPNames[index]
	case utils.ROLE_LB:
		publicIP = azureCloudState.InfoLoadBalancer.PublicIPName
	case utils.ROLE_DS:
		publicIP = azureCloudState.InfoDatabase.PublicIPNames[index]
	}

	if len(publicIP) == 0 {
		storage.Logger().Success("[skip] pub ip already deleted")
		return nil
	}

	publicIPAddressClient, err := obj.Client.PublicIPClient()
	if err != nil {
		return err
	}

	pollerResponse, err := obj.Client.BeginDeletePubIP(ctx, publicIPAddressClient, azureCloudState.ResourceGroupName, publicIP, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present

	_, err = obj.Client.PollUntilDoneDelPubIP(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.PublicIPNames[index] = ""
		azureCloudState.InfoWorkerPlanes.PublicIPIDs[index] = ""
		azureCloudState.InfoWorkerPlanes.PublicIPs[index] = ""
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.PublicIPNames[index] = ""
		azureCloudState.InfoControlPlanes.PublicIPIDs[index] = ""
		azureCloudState.InfoControlPlanes.PublicIPs[index] = ""
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.PublicIPID = ""
		azureCloudState.InfoLoadBalancer.PublicIPName = ""
		azureCloudState.InfoLoadBalancer.PublicIP = ""
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.PublicIPNames[index] = ""
		azureCloudState.InfoDatabase.PublicIPIDs[index] = ""
		azureCloudState.InfoDatabase.PublicIPs[index] = ""
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Deleted the pub IP", publicIP)
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory,
	nicName string, subnetID string, publicIPID string, networkSecurityGroupID string, index int) error {

	interfaceName := ""
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		interfaceName = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case utils.ROLE_CP:
		interfaceName = azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index]
	case utils.ROLE_LB:
		interfaceName = azureCloudState.InfoLoadBalancer.NetworkInterfaceName
	case utils.ROLE_DS:
		interfaceName = azureCloudState.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) != 0 {
		storage.Logger().Success("[skip] network interface already created", interfaceName)
		return nil
	}
	nicClient, err := obj.Client.NetInterfaceClient()
	if err != nil {
		return err
	}
	parameters := armnetwork.Interface{
		Location: to.Ptr(obj.Region),
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

	pollerResponse, err := obj.Client.BeginCreateNIC(ctx, nicClient, obj.ResourceGroup, nicName, parameters, nil)
	if err != nil {
		return err
	}
	// NOTE: Add the entry for name before polling starts so that state is present
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = nicName
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = nicName

	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.NetworkInterfaceName = nicName
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.NetworkInterfaceNames[index] = nicName
	}
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	resp, err := obj.Client.PollUntilDoneCreateNetInterface(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *resp.ID
		azureCloudState.InfoWorkerPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = *resp.ID
		azureCloudState.InfoControlPlanes.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress

	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.NetworkInterfaceID = *resp.ID
		azureCloudState.InfoLoadBalancer.PrivateIP = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.NetworkInterfaceIDs[index] = *resp.ID
		azureCloudState.InfoDatabase.PrivateIPs[index] = *resp.Properties.IPConfigurations[0].Properties.PrivateIPAddress
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Created network interface", *resp.Name)
	return nil
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int) error {
	interfaceName := ""
	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		interfaceName = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index]
	case utils.ROLE_CP:
		interfaceName = azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index]
	case utils.ROLE_LB:
		interfaceName = azureCloudState.InfoLoadBalancer.NetworkInterfaceName
	case utils.ROLE_DS:
		interfaceName = azureCloudState.InfoDatabase.NetworkInterfaceNames[index]
	}
	if len(interfaceName) == 0 {
		storage.Logger().Success("[skip] network interface already deleted")
		return nil
	}

	nicClient, err := obj.Client.NetInterfaceClient()
	if err != nil {
		return err
	}

	pollerResponse, err := obj.Client.BeginDeleteNIC(ctx, nicClient, azureCloudState.ResourceGroupName, interfaceName, nil)
	if err != nil {
		return err
	}

	// NOTE: Add the entry for name before polling starts so that state is present

	_, err = obj.Client.PollUntilDoneDelNetInterface(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	switch obj.Metadata.Role {
	case utils.ROLE_WP:
		azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = ""
		azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
		azureCloudState.InfoWorkerPlanes.PrivateIPs[index] = ""
	case utils.ROLE_CP:
		azureCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = ""
		azureCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
		azureCloudState.InfoControlPlanes.PrivateIPs[index] = ""
	case utils.ROLE_LB:
		azureCloudState.InfoLoadBalancer.NetworkInterfaceName = ""
		azureCloudState.InfoLoadBalancer.NetworkInterfaceID = ""
		azureCloudState.InfoLoadBalancer.PrivateIP = ""
	case utils.ROLE_DS:
		azureCloudState.InfoDatabase.NetworkInterfaceNames[index] = ""
		azureCloudState.InfoDatabase.NetworkInterfaceIDs[index] = ""
		azureCloudState.InfoDatabase.PrivateIPs[index] = ""
	}
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[azure] Deleted the network interface", interfaceName)

	return nil
}
