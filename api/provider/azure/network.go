package azure

import (
	"context"
	"fmt"
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

// DelNetwork implements resources.CloudFactory.
func (obj *AzureProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(azureCloudState.ResourceGroupName) == 0 {
		storage.Logger().Success("[skip] already deleted the resource group")
		return nil
	} else {
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
	if obj.HACluster {
		// create Virtual network, subnet, nsg, ....
		return fmt.Errorf("[azure] ha is not added")
	}
	storage.Logger().Success("[azure] created the resource group", *resourceGroup.Name)

	return nil
}

func (obj *AzureProvider) CreateResourceGroup(ctx context.Context, storage resources.StorageFactory) (*armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	resourceGroupClient, err := obj.resourceGroupsClient()
	if err != nil {
		return nil, err
	}
	resourceGroup, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		azCloudState.ResourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(obj.Region),
		},
		nil)
	if err != nil {
		return nil, err
	}
	storage.Logger().Success("Created resource group", *resourceGroup.Name)
	return &resourceGroup, nil
}

func (obj *AzureProvider) DeleteResourceGroup(ctx context.Context, storage resources.StorageFactory) error {
	resourceGroupClient, err := obj.resourceGroupsClient()
	if err != nil {
		return err
	}
	pollerResp, err := resourceGroupClient.BeginDelete(ctx, azureCloudState.ResourceGroupName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("Deleted resource group", azureCloudState.ResourceGroupName)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) (*armnetwork.Subnet, error) {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx, azureCloudState.ResourceGroupName, azureCloudState.VirtualNetworkName, subnetName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	azureCloudState.SubnetName = subnetName
	azureCloudState.SubnetID = *resp.ID

	storage.Logger().Success("Created subnet", *resp.Name)
	return &resp.Subnet, nil
}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {
	subnetClient, err := armnetwork.NewSubnetsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := subnetClient.BeginDelete(ctx, azureCloudState.ResourceGroupName, azureCloudState.VirtualNetworkName, subnetName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("Deleted subnet", subnetName)
	return nil
}

func (obj *AzureProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string) (*armnetwork.PublicIPAddress, error) {
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.PublicIPAddress{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic), // Static or Dynamic
		},
	}

	pollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(ctx, azureCloudState.ResourceGroupName, publicIPName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	storage.Logger().Success("Created public IP address", *resp.Name)
	return &resp.PublicIPAddress, err
}

func (obj *AzureProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string) error {
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := publicIPAddressClient.BeginDelete(ctx, azureCloudState.ResourceGroupName, publicIPName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("Deleted the pubIP", publicIPName)
	return nil
}

func (obj *AzureProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory, resourceName, nicName string, subnetID string, publicIPID string, networkSecurityGroupID string) (*armnetwork.Interface, error) {
	nicClient, err := armnetwork.NewInterfacesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	parameters := armnetwork.Interface{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr(resourceName),
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

	pollerResponse, err := nicClient.BeginCreateOrUpdate(ctx, obj.ResourceGroup, nicName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	storage.Logger().Success("Created network interface", *resp.Name)
	return &resp.Interface, err
}

func (obj *AzureProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, nicName string) error {
	nicClient, err := armnetwork.NewInterfacesClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := nicClient.BeginDelete(ctx, obj.ResourceGroup, nicName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	storage.Logger().Success("Deleted the nic", nicName)

	return nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context, storage resources.StorageFactory) error {
	vnetClient, err := armnetwork.NewVirtualNetworksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}

	pollerResponse, err := vnetClient.BeginDelete(ctx, obj.ResourceGroup, azureCloudState.VirtualNetworkName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	storage.Logger().Success("Deleted virtual network", azureCloudState.VirtualNetworkName)
	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	vnetClient, err := armnetwork.NewVirtualNetworksClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
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

	pollerResponse, err := vnetClient.BeginCreateOrUpdate(ctx, azureCloudState.ResourceGroupName, virtualNetworkName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	azureCloudState.VirtualNetworkName = *resp.Name
	azureCloudState.VirtualNetworkID = *resp.ID
	storage.Logger().Success("Created virtual network", *resp.Name)
	return &resp.VirtualNetwork, nil
}
