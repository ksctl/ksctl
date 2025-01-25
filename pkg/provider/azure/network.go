// Copyright 2024 Ksctl Authors
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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func (p *Provider) NewNetwork() error {
	<-p.chResName

	if len(p.state.CloudInfra.Azure.ResourceGroupName) != 0 {
		p.l.Print(p.ctx, "skipped already created the resource group", "name", p.state.CloudInfra.Azure.ResourceGroupName)
	} else {
		var err error
		var resourceGroup armresources.ResourceGroupsClientCreateOrUpdateResponse

		// NOTE: for the azure resource group we are not using the resName field
		parameter := armresources.ResourceGroup{
			Location: utilities.Ptr(p.Region),
		}

		p.l.Debug(p.ctx, "Printing", "resourceGrpConfig", parameter)

		resourceGroup, err = p.client.CreateResourceGrp(parameter, nil)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Azure.ResourceGroupName = *resourceGroup.Name

		if err := p.store.Write(p.state); err != nil {
			return err
		}
		p.l.Success(p.ctx, "created the resource group", "name", *resourceGroup.Name)
	}
	if p.SelfManaged {
		virtNet := p.ClusterName + "-vnet"
		subNet := p.ClusterName + "-subnet"

		if err := p.CreateVirtualNetwork(virtNet); err != nil {
			return err
		}

		if err := p.CreateSubnet(subNet); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) CreateVirtualNetwork(resName string) error {

	if len(p.state.CloudInfra.Azure.VirtualNetworkName) != 0 {
		p.l.Print(p.ctx, "skipped virtualNetwork already created", "name", p.state.CloudInfra.Azure.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.VirtualNetwork{
		Location: utilities.Ptr(p.Region),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					utilities.Ptr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}
	p.state.CloudInfra.Azure.NetCidr = "10.1.0.0/16"

	p.l.Debug(p.ctx, "Printing", "virtualNetworkConfig", parameters)

	pollerResponse, err := p.client.BeginCreateVirtNet(resName, parameters, nil)
	if err != nil {
		return err
	}
	p.state.CloudInfra.Azure.VirtualNetworkName = resName

	if err := p.store.Write(p.state); err != nil {
		return err
	}
	p.l.Print(p.ctx, "creating virtual network...", "name", resName)

	resp, err := p.client.PollUntilDoneCreateVirtNet(p.ctx, pollerResponse, nil)
	if err != nil {
		return err
	}
	p.state.CloudInfra.Azure.VirtualNetworkID = *resp.ID
	if err := p.store.Write(p.state); err != nil {
		return err
	}
	p.l.Success(p.ctx, "Created virtual network", "name", *resp.Name)
	return nil
}

func (p *Provider) CreateSubnet(subnetName string) error {

	if len(p.state.CloudInfra.Azure.SubnetName) != 0 {
		p.l.Print(p.ctx, "skipped subnet already created", "name", p.state.CloudInfra.Azure.VirtualNetworkName)
		return nil
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: utilities.Ptr("10.1.0.0/16"),
		},
	}

	p.l.Debug(p.ctx, "Printing", "subnetConfig", parameters)

	pollerResponse, err := p.client.BeginCreateSubNet(p.state.CloudInfra.Azure.VirtualNetworkName, subnetName, parameters, nil)

	if err != nil {
		return err
	}
	p.state.CloudInfra.Azure.SubnetName = subnetName
	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Print(p.ctx, "creating subnet...", "name", subnetName)

	resp, err := p.client.PollUntilDoneCreateSubNet(p.ctx, pollerResponse, nil)

	if err != nil {
		return err
	}
	p.state.CloudInfra.Azure.SubnetID = *resp.ID
	if err := p.store.Write(p.state); err != nil {
		return err
	}
	p.l.Success(p.ctx, "Created subnet", "name", *resp.Name)
	return nil
}

func (p *Provider) DelNetwork() error {

	if len(p.state.CloudInfra.Azure.ResourceGroupName) == 0 {
		p.l.Print(p.ctx, "skipped already deleted the resource group")
		return nil
	} else {
		if p.SelfManaged {
			if err := p.DeleteSubnet(); err != nil {
				return err
			}

			if err := p.DeleteVirtualNetwork(); err != nil {
				return err
			}
		}

		pollerResp, err := p.client.BeginDeleteResourceGrp(nil)
		if err != nil {
			return err
		}
		_, err = p.client.PollUntilDoneDelResourceGrp(p.ctx, pollerResp, nil)
		if err != nil {
			return err
		}

		rgname := p.state.CloudInfra.Azure.ResourceGroupName

		p.l.Debug(p.ctx, "Printing", "resourceGrpName", rgname)

		p.state.CloudInfra.Azure.ResourceGroupName = ""
		if err := p.store.Write(p.state); err != nil {
			return err
		}
		p.l.Success(p.ctx, "deleted the resource group", "name", rgname)
	}

	return p.store.DeleteCluster()
}

func (p *Provider) DeleteSubnet() error {

	subnet := p.state.CloudInfra.Azure.SubnetName
	p.l.Debug(p.ctx, "Printing", "subnetName", subnet)
	if len(subnet) == 0 {
		p.l.Print(p.ctx, "skipped subnet already deleted", "name", subnet)
		return nil
	}

	pollerResponse, err := p.client.BeginDeleteSubNet(p.state.CloudInfra.Azure.VirtualNetworkName, subnet, nil)
	if err != nil {
		return err
	}

	_, err = p.client.PollUntilDoneDelSubNet(p.ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	p.state.CloudInfra.Azure.SubnetName = ""
	p.state.CloudInfra.Azure.SubnetID = ""

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "Deleted subnet", "name", subnet)
	return nil
}

func (p *Provider) DeleteVirtualNetwork() error {

	vnet := p.state.CloudInfra.Azure.VirtualNetworkName
	p.l.Debug(p.ctx, "Printing", "virtNetName", vnet)
	if len(vnet) == 0 {
		p.l.Print(p.ctx, "virtual network already deleted", "name", vnet)
		return nil
	}

	pollerResponse, err := p.client.BeginDeleteVirtNet(vnet, nil)
	if err != nil {
		return err
	}

	_, err = p.client.PollUntilDoneDelVirtNet(p.ctx, pollerResponse, nil)
	if err != nil {
		return err
	}

	p.state.CloudInfra.Azure.VirtualNetworkID = ""
	p.state.CloudInfra.Azure.VirtualNetworkName = ""
	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "Deleted virtual network", "name", vnet)
	return nil
}
