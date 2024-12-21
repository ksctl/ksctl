// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *AzureProvider) NewNetwork(storage types.StorageFactory) error {
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

		if err := storage.Write(mainStateDocument); err != nil {
			return err
		}
		log.Success(azureCtx, "created the resource group", "name", *resourceGroup.Name)
	}
	if obj.haCluster {
		virtNet := obj.clusterName + "-vnet"
		subNet := obj.clusterName + "-subnet"

		if err := obj.CreateVirtualNetwork(azureCtx, storage, virtNet); err != nil {
			return err
		}

		if err := obj.CreateSubnet(azureCtx, storage, subNet); err != nil {
			return err
		}
	}

	return nil
}

func (obj *AzureProvider) CreateVirtualNetwork(ctx context.Context, storage types.StorageFactory, resName string) error {

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

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Print(azureCtx, "creating virtual network...", "name", resName)

	resp, err := obj.client.PollUntilDoneCreateVirtNet(ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.VirtualNetworkID = *resp.ID
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success(azureCtx, "Created virtual network", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) CreateSubnet(ctx context.Context, storage types.StorageFactory, subnetName string) error {

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
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Print(azureCtx, "creating subnet...", "name", subnetName)

	resp, err := obj.client.PollUntilDoneCreateSubNet(ctx, pollerResponse, nil)

	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.SubnetID = *resp.ID
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	log.Success(azureCtx, "Created subnet", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) DelNetwork(storage types.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Azure.ResourceGroupName) == 0 {
		log.Print(azureCtx, "skipped already deleted the resource group")
		return nil
	} else {
		if obj.haCluster {
			if err := obj.DeleteSubnet(azureCtx, storage); err != nil {
				return err
			}

			if err := obj.DeleteVirtualNetwork(azureCtx, storage); err != nil {
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
		if err := storage.Write(mainStateDocument); err != nil {
			return err
		}
		log.Success(azureCtx, "deleted the resource group", "name", rgname)
	}

	return storage.DeleteCluster()
}

func (obj *AzureProvider) DeleteSubnet(ctx context.Context, storage types.StorageFactory) error {

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

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "Deleted subnet", "name", subnet)
	return nil
}

func (obj *AzureProvider) DeleteVirtualNetwork(ctx context.Context, storage types.StorageFactory) error {

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
	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "Deleted virtual network", "name", vnet)
	return nil
}
