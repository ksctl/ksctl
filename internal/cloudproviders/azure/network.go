package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func (obj *AzureProvider) NewNetwork() error {
	<-obj.chResName

	if len(mainStateDocument.CloudInfra.Azure.ResourceGroupName) != 0 {
		log.Print(azureCtx, "skipped already created the resource group", "name", mainStateDocument.CloudInfra.Azure.ResourceGroupName)
	} else {
		var err error
		var resourceGroup armresources.ResourceGroupsClientCreateOrUpdateResponse

		// NOTE: for the azure resource group we are not using the resName field
		parameter := armresources.ResourceGroup{
			Location: utilities.Ptr(obj.region),
		}

		log.Debug(azureCtx, "Printing", "resourceGrpConfig", parameter)

		resourceGroup, err = obj.client.CreateResourceGrp(parameter, nil)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Azure.ResourceGroupName = *resourceGroup.Name

		if err := obj.storage.Write(mainStateDocument); err != nil {
			return err
		}
		log.Success(azureCtx, "created the resource group", "name", *resourceGroup.Name)
	}
	if obj.haCluster {
		virtNet := obj.clusterName + "-vnet"
		subNet := obj.clusterName + "-subnet"

		if err := obj.CreateVirtualNetwork(azureCtx, virtNet); err != nil {
			return err
		}

		if err := obj.CreateSubnet(azureCtx, subNet); err != nil {
			return err
		}
	}

	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, resName string) error {

	if len(mainStateDocument.CloudInfra.Azure.VirtualNetworkName) != 0 {
		log.Print(azureCtx, "skipped virtualNetwork already created", "name", mainStateDocument.CloudInfra.Azure.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.VirtualNetwork{
		Location: utilities.Ptr(obj.region),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					utilities.Ptr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}
	mainStateDocument.CloudInfra.Azure.NetCidr = "10.1.0.0/16"

	log.Debug(azureCtx, "Printing", "virtualNetworkConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateVirtNet(resName, parameters, nil)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.VirtualNetworkName = resName

	if err := obj.storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Print(azureCtx, "creating virtual network...", "name", resName)

	resp, err := obj.client.PollUntilDoneCreateVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.VirtualNetworkID = *resp.ID
	if err := obj.storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success(azureCtx, "Created virtual network", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, subnetName string) error {

	if len(mainStateDocument.CloudInfra.Azure.SubnetName) != 0 {
		log.Print(azureCtx, "skipped subnet already created", "name", mainStateDocument.CloudInfra.Azure.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: utilities.Ptr("10.1.0.0/16"),
		},
	}

	log.Debug(azureCtx, "Printing", "subnetConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateSubNet(mainStateDocument.CloudInfra.Azure.VirtualNetworkName, subnetName, parameters, nil)

	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.SubnetName = subnetName
	if err := obj.storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Print(azureCtx, "creating subnet...", "name", subnetName)

	resp, err := obj.client.PollUntilDoneCreateSubNet(ctx, pollerResponse, nil)

	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.SubnetID = *resp.ID
	if err := obj.storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success(azureCtx, "Created subnet", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) DelNetwork() error {

	if len(mainStateDocument.CloudInfra.Azure.ResourceGroupName) == 0 {
		log.Print(azureCtx, "skipped already deleted the resource group")
		return nil
	} else {
		if obj.haCluster {
			if err := obj.DeleteSubnet(azureCtx); err != nil {
				return err
			}

			if err := obj.DeleteVirtualNetwork(azureCtx); err != nil {
				return err
			}
		}

		pollerResp, err := obj.client.BeginDeleteResourceGrp(nil)
		if err != nil {
			return err
		}
		_, err = obj.client.PollUntilDoneDelResourceGrp(azureCtx, pollerResp, nil)
		if err != nil {
			return err
		}

		rgname := mainStateDocument.CloudInfra.Azure.ResourceGroupName

		log.Debug(azureCtx, "Printing", "resourceGrpName", rgname)

		mainStateDocument.CloudInfra.Azure.ResourceGroupName = ""
		if err := obj.storage.Write(mainStateDocument); err != nil {
			return err
		}
		log.Success(azureCtx, "deleted the resource group", "name", rgname)
	}

	return obj.storage.DeleteCluster()
}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context) error {

	subnet := mainStateDocument.CloudInfra.Azure.SubnetName
	log.Debug(azureCtx, "Printing", "subnetName", subnet)
	if len(subnet) == 0 {
		log.Print(azureCtx, "skipped subnet already deleted", "name", subnet)
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteSubNet(mainStateDocument.CloudInfra.Azure.VirtualNetworkName, subnet, nil)
	if err != nil {
		return err
	}

	_, err = obj.client.PollUntilDoneDelSubNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Azure.SubnetName = ""
	mainStateDocument.CloudInfra.Azure.SubnetID = ""

	if err := obj.storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "Deleted subnet", "name", subnet)
	return nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context) error {

	vnet := mainStateDocument.CloudInfra.Azure.VirtualNetworkName
	log.Debug(azureCtx, "Printing", "virtNetName", vnet)
	if len(vnet) == 0 {
		log.Print(azureCtx, "virtual network already deleted", "name", vnet)
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteVirtNet(vnet, nil)
	if err != nil {
		return err
	}

	_, err = obj.client.PollUntilDoneDelVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Azure.VirtualNetworkID = ""
	mainStateDocument.CloudInfra.Azure.VirtualNetworkName = ""
	if err := obj.storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "Deleted virtual network", "name", vnet)
	return nil
}
