package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/ksctl/ksctl/pkg/resources"
)

// NewNetwork implements resources.CloudFactory
func (obj *AzureProvider) NewNetwork(storage resources.StorageFactory) error {
	<-obj.chResName

	if len(mainStateDocument.CloudInfra.Azure.ResourceGroupName) != 0 {
		log.Print("skipped already created the resource group", "name", mainStateDocument.CloudInfra.Azure.ResourceGroupName)
	} else {
		var err error
		var resourceGroup armresources.ResourceGroupsClientCreateOrUpdateResponse

		// NOTE: for the azure resource group we are not using the resName field
		parameter := armresources.ResourceGroup{
			Location: to.Ptr(obj.region),
		}

		log.Debug("Printing", "resourceGrpConfig", parameter)

		resourceGroup, err = obj.client.CreateResourceGrp(parameter, nil)
		if err != nil {
			return log.NewError(err.Error())
		}

		mainStateDocument.CloudInfra.Azure.ResourceGroupName = *resourceGroup.Name

		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}
		log.Success("created the resource group", "name", *resourceGroup.Name)
	}
	if obj.haCluster {
		virtNet := obj.clusterName + "-vnet"
		subNet := obj.clusterName + "-subnet"
		// virtual net
		if err := obj.CreateVirtualNetwork(ctx, storage, virtNet); err != nil {
			return log.NewError(err.Error())
		}

		// subnet
		if err := obj.CreateSubnet(ctx, storage, subNet); err != nil {
			return log.NewError(err.Error())
		}
	}

	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, storage resources.StorageFactory, resName string) error {

	if len(mainStateDocument.CloudInfra.Azure.VirtualNetworkName) != 0 {
		log.Print("skipped virtualNetwork already created", "name", mainStateDocument.CloudInfra.Azure.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.VirtualNetwork{
		Location: to.Ptr(obj.region),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.Ptr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}
	mainStateDocument.CloudInfra.Azure.NetCidr = "10.1.0.0/16"

	log.Debug("Printing", "virtualNetworkConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateVirtNet(resName, parameters, nil)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.VirtualNetworkName = resName

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Print("creating virtual network...", "name", resName)

	resp, err := obj.client.PollUntilDoneCreateVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.VirtualNetworkID = *resp.ID
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success("Created virtual network", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(mainStateDocument.CloudInfra.Azure.SubnetName) != 0 {
		log.Print("skipped subnet already created", "name", mainStateDocument.CloudInfra.Azure.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	log.Debug("Printing", "subnetConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateSubNet(mainStateDocument.CloudInfra.Azure.VirtualNetworkName, subnetName, parameters, nil)

	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.SubnetName = subnetName
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Print("creating subnet...", "name", subnetName)

	resp, err := obj.client.PollUntilDoneCreateSubNet(ctx, pollerResponse, nil)

	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.SubnetID = *resp.ID
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success("Created subnet", "name", *resp.Name)
	return nil
}

// DelNetwork implements resources.CloudFactory.
func (obj *AzureProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Azure.ResourceGroupName) == 0 {
		log.Print("skipped already deleted the resource group")
		return nil
	} else {
		if obj.haCluster {
			// delete subnet
			if err := obj.DeleteSubnet(ctx, storage); err != nil {
				return log.NewError(err.Error())
			}

			// delete vnet
			if err := obj.DeleteVirtualNetwork(ctx, storage); err != nil {
				return log.NewError(err.Error())
			}
		}

		pollerResp, err := obj.client.BeginDeleteResourceGrp(nil)
		if err != nil {
			return log.NewError(err.Error())
		}
		_, err = obj.client.PollUntilDoneDelResourceGrp(ctx, pollerResp, nil)
		if err != nil {
			return log.NewError(err.Error())
		}

		rgname := mainStateDocument.CloudInfra.Azure.ResourceGroupName

		log.Debug("Printing", "resourceGrpName", rgname)

		mainStateDocument.CloudInfra.Azure.ResourceGroupName = ""
		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}
		log.Success("deleted the resource group", "name", rgname)
	}

	if err := storage.DeleteCluster(); err != nil {
		return log.NewError(err.Error())
	}

	return nil

}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory) error {

	subnet := mainStateDocument.CloudInfra.Azure.SubnetName
	log.Debug("Printing", "subnetName", subnet)
	if len(subnet) == 0 {
		log.Print("skipped subnet already deleted", "name", subnet)
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

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success("Deleted subnet", "name", subnet)
	return nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context, storage resources.StorageFactory) error {

	vnet := mainStateDocument.CloudInfra.Azure.VirtualNetworkName
	log.Debug("Printing", "virtNetName", vnet)
	if len(vnet) == 0 {
		log.Print("virtual network already deleted", "name", vnet)
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
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success("Deleted virtual network", "name", vnet)
	return nil
}
