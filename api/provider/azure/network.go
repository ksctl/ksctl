package azure

import (
	"context"
	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kubesimplify/ksctl/api/resources"
)

// NewNetwork implements resources.CloudFactory.
func (obj *AzureProvider) NewNetwork(storage resources.StorageFactory) error {

	if len(azureCloudState.ResourceGroupName) != 0 {
		storage.Logger().Success("[skip] already created the resource group", azureCloudState.ResourceGroupName)
		return nil
	}
	var err error
	var resourceGroup armresources.ResourceGroupsClientCreateOrUpdateResponse

	// NOTE: for the azure resource group we are not using the resName field
	parameter := armresources.ResourceGroup{
		Location: to.Ptr(obj.Region),
	}
	resourceGroup, err = obj.Client.CreateResourceGrp(parameter, nil)
	if err != nil {
		return err
	}

	azureCloudState.ResourceGroupName = *resourceGroup.Name

	if err := storage.Path(generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)).
		Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		return err
	}

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] created the resource group", *resourceGroup.Name)

	if obj.HACluster {
		virtNet := obj.ClusterName + "-vnet"
		subNet := obj.ClusterName + "-subnet"
		// virtual net
		if err := obj.CreateVirtualNetwork(ctx, storage, virtNet); err != nil {
			return err
		}

		// subnet
		if err := obj.CreateSubnet(ctx, storage, subNet); err != nil {
			return err
		}
	}

	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, resName string) error {

	if len(azureCloudState.VirtualNetworkName) != 0 {
		storage.Logger().Success("[skip] virtualNetwork already created", azureCloudState.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.VirtualNetwork{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.Ptr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}

	pollerResponse, err := obj.Client.BeginCreateVirtNet(resName, parameters, nil)
	if err != nil {
		return err
	}
	azureCloudState.VirtualNetworkName = resName

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Print("[azure] creating virtual network...", resName)

	resp, err := obj.Client.PollUntilDoneCreateVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	azureCloudState.VirtualNetworkID = *resp.ID
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Created virtual network", *resp.Name)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(azureCloudState.SubnetName) != 0 {
		storage.Logger().Success("[skip] subnet already created", azureCloudState.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	pollerResponse, err := obj.Client.BeginCreateSubNet(azureCloudState.VirtualNetworkName, subnetName, parameters, nil)

	if err != nil {
		return err
	}
	azureCloudState.SubnetName = subnetName
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Print("[azure] creating subnet...", subnetName)

	resp, err := obj.Client.PollUntilDoneCreateSubNet(ctx, pollerResponse, nil)

	if err != nil {
		return err
	}
	azureCloudState.SubnetID = *resp.ID
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[azure] Created subnet", *resp.Name)
	return nil
}

// DelNetwork implements resources.CloudFactory.
func (obj *AzureProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(azureCloudState.ResourceGroupName) == 0 {
		storage.Logger().Success("[skip] already deleted the resource group")
		return nil
	} else {
		if obj.HACluster {
			// delete subnet
			if err := obj.DeleteSubnet(ctx, storage); err != nil {
				return err
			}

			// delete vnet
			if err := obj.DeleteVirtualNetwork(ctx, storage); err != nil {
				return err
			}
		}

		pollerResp, err := obj.Client.BeginDeleteResourceGrp(nil)
		if err != nil {
			return err
		}
		_, err = obj.Client.PollUntilDoneDelResourceGrp(ctx, pollerResp, nil)
		if err != nil {
			return err
		}

		rgname := azureCloudState.ResourceGroupName

		azureCloudState.ResourceGroupName = ""
		if err := saveStateHelper(storage); err != nil {
			return err
		}
		storage.Logger().Success("[azure] deleted the resource group", rgname)

	}

	if err := storage.Path(generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)).
		DeleteDir(); err != nil {
		return err
	}

	return nil

}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory) error {

	subnet := azureCloudState.SubnetName
	if len(subnet) == 0 {
		storage.Logger().Success("[skip] subnet already deleted", subnet)
		return nil
	}

	pollerResponse, err := obj.Client.BeginDeleteSubNet(azureCloudState.VirtualNetworkName, subnet, nil)
	if err != nil {
		return err
	}

	_, err = obj.Client.PollUntilDoneDelSubNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	azureCloudState.SubnetName = ""
	azureCloudState.SubnetID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[azure] Deleted subnet", subnet)
	return nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context, storage resources.StorageFactory) error {

	vnet := azureCloudState.VirtualNetworkName
	if len(vnet) == 0 {
		storage.Logger().Success("[skip] subnet already deleted", vnet)
		return nil
	}

	pollerResponse, err := obj.Client.BeginDeleteVirtNet(vnet, nil)
	if err != nil {
		return err
	}

	_, err = obj.Client.PollUntilDoneDelVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	azureCloudState.VirtualNetworkID = ""
	azureCloudState.VirtualNetworkName = ""
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Success("[azure] Deleted virtual network", vnet)
	return nil
}
