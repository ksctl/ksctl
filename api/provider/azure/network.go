package azure

import (
	"context"
	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kubesimplify/ksctl/api/resources"
)

func (obj *AzureProvider) resourceGroupsClient() (*armresources.ResourceGroupsClient, error) {

	resourceGroupClient, err := armresources.NewResourceGroupsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	return resourceGroupClient, nil
}

func (obj *AzureProvider) virtualNetworkClient() (*armnetwork.VirtualNetworksClient, error) {
	vnetClient, err := armnetwork.NewVirtualNetworksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return vnetClient, nil
}

func (obj *AzureProvider) subnetClient() (*armnetwork.SubnetsClient, error) {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return subnetClient, nil
}

// NewNetwork implements resources.CloudFactory.
func (obj *AzureProvider) NewNetwork(storage resources.StorageFactory) error {

	if len(azureCloudState.ResourceGroupName) != 0 {
		storage.Logger().Success("[skip] already created the resource group", azureCloudState.ResourceGroupName)
		return nil
	}
	var err error
	var resourceGroup armresources.ResourceGroupsClientCreateOrUpdateResponse

	rgclient, err := obj.resourceGroupsClient()
	if err != nil {
		return err
	}

	// NOTE: for the azure resource group we are not using the resName field
	resourceGroup, err = rgclient.CreateOrUpdate(
		ctx,
		obj.ResourceGroup,
		armresources.ResourceGroup{
			Location: to.Ptr(obj.Region),
		},
		nil)
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

	// TODO: create subnet and virtual network
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

	vnetClient, err := obj.virtualNetworkClient()
	if err != nil {
		return err
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

	pollerResponse, err := vnetClient.BeginCreateOrUpdate(ctx, azureCloudState.ResourceGroupName,
		resName, parameters, nil)
	if err != nil {
		return err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	azureCloudState.VirtualNetworkName = *resp.Name
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

	subnetClient, err := obj.subnetClient()
	if err != nil {
		return err
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx,
		azureCloudState.ResourceGroupName, azureCloudState.VirtualNetworkName,
		subnetName, parameters, nil)
	if err != nil {
		return err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	azureCloudState.SubnetName = subnetName
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
		rgclient, err := obj.resourceGroupsClient()
		if err != nil {
			return err
		}
		pollerResp, err := rgclient.BeginDelete(ctx, azureCloudState.ResourceGroupName, nil)
		if err != nil {
			return err
		}
		_, err = pollerResp.PollUntilDone(ctx, nil)
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

	subnetClient, err := obj.subnetClient()
	if err != nil {
		return err
	}

	pollerResponse, err := subnetClient.BeginDelete(ctx, azureCloudState.ResourceGroupName, azureCloudState.VirtualNetworkName, subnet, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
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

	vnetClient, err := obj.virtualNetworkClient()
	if err != nil {
		return err
	}

	pollerResponse, err := vnetClient.BeginDelete(ctx, obj.ResourceGroup, vnet, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
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
