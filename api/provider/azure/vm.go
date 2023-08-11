package azure

import (
	"context"
	"encoding/base64"
	"errors"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// DelVM implements resources.CloudFactory.
func (*AzureProvider) DelVM(state resources.StorageFactory, indexNo int) error {
	panic("unimplemented")
}

// NewVM implements resources.CloudFactory.
func (*AzureProvider) NewVM(state resources.StorageFactory, indexNo int) error {
	return errors.New("unimplemented")
}

func (obj *AzureProvider) CreateVM(ctx context.Context, storage resources.StorageFactory, vmName, networkInterfaceID, diskName, script string) (*armcompute.VirtualMachine, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	//require ssh key for authentication on linux
	// TODO: check which functionality can help here
	sshPublicKeyPath := utils.GetPath(utils.OTHER_PATH, "azure", "ha", obj.ClusterName+" "+obj.ResourceGroup+" "+obj.Region, "keypair.pub")
	var sshBytes []byte
	_, err = os.Stat(sshPublicKeyPath)
	if err != nil {
		return nil, err
	}

	sshBytes, err = os.ReadFile(sshPublicKeyPath)
	if err != nil {
		return nil, err
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
				ComputerName:  to.Ptr(vmName),
				AdminUsername: to.Ptr("azureuser"),
				CustomData:    to.Ptr(base64.StdEncoding.EncodeToString([]byte(script))),
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
						ID: to.Ptr(networkInterfaceID),
					},
				},
			},
		},
	}

	pollerResponse, err := vmClient.BeginCreateOrUpdate(ctx, obj.ResourceGroup, vmName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	storage.Logger().Success("Created network virtual machine", *resp.Name)
	return &resp.VirtualMachine, nil
}

func (obj *AzureProvider) DeleteVM(ctx context.Context, storage resources.StorageFactory, vmName string) error {
	vmClient, err := armcompute.NewVirtualMachinesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := vmClient.BeginDelete(ctx, obj.ResourceGroup, vmName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("Deleted the vm", vmName)
	return nil
}

func (obj *AzureProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, diskName string) error {
	diskClient, err := armcompute.NewDisksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := diskClient.BeginDelete(ctx, obj.ResourceGroup, diskName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	storage.Logger().Success("Deleted disk", diskName)
	return nil
}
