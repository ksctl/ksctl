package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kubesimplify/ksctl/pkg/resources"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// NewNetwork implements resources.CloudFactory.
func (obj *AzureProvider) NewNetwork(storage resources.StorageFactory) error {
	_ = obj.metadata.resName
	obj.mxName.Unlock()

	if len(azureCloudState.ResourceGroupName) != 0 {
		log.Print("skipped already created the resource group", "name", azureCloudState.ResourceGroupName)
		return nil
	}
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

	azureCloudState.ResourceGroupName = *resourceGroup.Name

	if err := storage.Path(generatePath(UtilClusterPath, clusterType, clusterDirName)).
		Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		return log.NewError(err.Error())
	}

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("created the resource group", "name", *resourceGroup.Name)

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

	if len(azureCloudState.VirtualNetworkName) != 0 {
		log.Print("skipped virtualNetwork already created", "name", azureCloudState.VirtualNetworkName)
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

	log.Debug("Printing", "virtualNetworkConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateVirtNet(resName, parameters, nil)
	if err != nil {
		return err
	}
	azureCloudState.VirtualNetworkName = resName

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	log.Print("creating virtual network...", "name", resName)

	resp, err := obj.client.PollUntilDoneCreateVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	azureCloudState.VirtualNetworkID = *resp.ID
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	log.Success("Created virtual network", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, storage resources.StorageFactory, subnetName string) error {

	if len(azureCloudState.SubnetName) != 0 {
		log.Print("skipped subnet already created", "name", azureCloudState.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.0.0/16"),
		},
	}

	log.Debug("Printing", "subnetConfig", parameters)

	pollerResponse, err := obj.client.BeginCreateSubNet(azureCloudState.VirtualNetworkName, subnetName, parameters, nil)

	if err != nil {
		return err
	}
	azureCloudState.SubnetName = subnetName
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Print("creating subnet...", "name", subnetName)

	resp, err := obj.client.PollUntilDoneCreateSubNet(ctx, pollerResponse, nil)

	if err != nil {
		return err
	}
	azureCloudState.SubnetID = *resp.ID
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	log.Success("Created subnet", "name", *resp.Name)
	return nil
}

// DelNetwork implements resources.CloudFactory.
func (obj *AzureProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(azureCloudState.ResourceGroupName) == 0 {
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

		rgname := azureCloudState.ResourceGroupName

		log.Debug("Printing", "resourceGrpName", rgname)

		azureCloudState.ResourceGroupName = ""
		if err := saveStateHelper(storage); err != nil {
			return log.NewError(err.Error())
		}
		log.Success("deleted the resource group", "name", rgname)
	}

	if err := storage.Path(generatePath(UtilClusterPath, clusterType, clusterDirName)).
		DeleteDir(); err != nil {
		return log.NewError(err.Error())
	}

	return nil

}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, storage resources.StorageFactory) error {

	subnet := azureCloudState.SubnetName
	log.Debug("Printing", "subnetName", subnet)
	if len(subnet) == 0 {
		log.Print("skipped subnet already deleted", "name", subnet)
		return nil
	}

	pollerResponse, err := obj.client.BeginDeleteSubNet(azureCloudState.VirtualNetworkName, subnet, nil)
	if err != nil {
		return err
	}

	_, err = obj.client.PollUntilDoneDelSubNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	azureCloudState.SubnetName = ""
	azureCloudState.SubnetID = ""

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("Deleted subnet", "name", subnet)
	return nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context, storage resources.StorageFactory) error {

	vnet := azureCloudState.VirtualNetworkName
	log.Debug("Printing", "virtNetName", vnet)
	if len(vnet) == 0 {
		log.Print("subnet already deleted", "name", vnet)
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

	azureCloudState.VirtualNetworkID = ""
	azureCloudState.VirtualNetworkName = ""
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	log.Success("Deleted virtual network", "name", vnet)
	return nil
}
