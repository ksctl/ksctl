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

//go:build !testing_azure

package azure

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"

	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

func ProvideClient() CloudSDK {
	return &AzureClient{}
}

type AzureClient struct {
	azureTokenCred azcore.TokenCredential
	region         string
	resourceGrp    string
	b              *Provider
}

func (p *AzureClient) PollUntilDoneCreateNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneCreatePubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneCreateVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelDisk(ctx context.Context, poll *runtime.Poller[armcompute.DisksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.DisksClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelPubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneCreateNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelResourceGrp(ctx context.Context, poll *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armresources.ResourceGroupsClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneCreateSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneCreateVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneCreateAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) PollUntilDoneDelAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			p.b.l.NewError(p.b.ctx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (p *AzureClient) setRequiredENV_VAR() error {

	err := os.Setenv("AZURE_SUBSCRIPTION_ID", p.b.subscriptionID)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			err,
		)
	}

	err = os.Setenv("AZURE_TENANT_ID", p.b.tenantID)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			err,
		)
	}

	err = os.Setenv("AZURE_CLIENT_ID", p.b.clientID)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			err,
		)
	}

	err = os.Setenv("AZURE_CLIENT_SECRET", p.b.clientSecret)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			err,
		)
	}
	return nil
}

func (p *AzureClient) InitClient(b *Provider) error {
	p.b = b
	p.region = b.Region
	p.resourceGrp = b.resourceGroup

	err := p.setRequiredENV_VAR()
	if err != nil {
		return err
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			p.b.l.NewError(p.b.ctx, "defaultAzureCredential", "Reason", err),
		)
	}
	p.azureTokenCred = cred
	return nil
}

func (p *AzureClient) ListLocations() ([]string, error) {
	clientFactory, err := armsubscriptions.NewClientFactory(p.azureTokenCred, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
		)
	}
	pager := clientFactory.NewClient().NewListLocationsPager(p.b.subscriptionID, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil})

	var validReg []string
	for pager.More() {
		page, err := pager.NextPage(p.b.ctx)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.b.l.NewError(p.b.ctx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			validReg = append(validReg, *v.Name)
		}
	}
	return validReg, nil
}

func (p *AzureClient) ListKubernetesVersions() (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error) {
	clientFactory, err := armcontainerservice.NewClientFactory(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := clientFactory.
		NewManagedClustersClient().
		ListKubernetesVersions(p.b.ctx, p.region, nil); err != nil {
		return res, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			p.b.l.NewError(p.b.ctx, "failed to get managed kubernetes versions", "Reason", err),
		)
	} else {
		return res, nil
	}
}

func (p *AzureClient) ListVMTypes() ([]string, error) {
	clientFactory, err := armcompute.NewClientFactory(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}
	pager := clientFactory.NewVirtualMachineSizesClient().NewListPager(p.region, nil)

	var validSize []string
	for pager.More() {

		page, err := pager.NextPage(p.b.ctx)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.b.l.NewError(p.b.ctx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			validSize = append(validSize, *v.Name)
		}
	}
	return validSize, nil
}

func (p *AzureClient) PublicIPClient() (*armnetwork.PublicIPAddressesClient, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}
	return client, nil
}

func (p *AzureClient) CreateResourceGrp(parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	client, err := armresources.NewResourceGroupsClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return armresources.ResourceGroupsClientCreateOrUpdateResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}
	if res, err := client.CreateOrUpdate(p.b.ctx, p.resourceGrp, parameters, options); err != nil {
		return armresources.ResourceGroupsClientCreateOrUpdateResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create resource group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteResourceGrp(options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	client, err := armresources.NewResourceGroupsClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}
	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete resource group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreateVirtNet(virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewVirtualNetworksClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, virtualNetworkName, parameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create virtual network", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteVirtNet(virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	client, err := armnetwork.NewVirtualNetworksClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, virtualNetworkName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete virtual network", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreateSubNet(virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewSubnetsClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, virtualNetworkName,
		subnetName, subnetParameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create subnet", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteSubNet(virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	client, err := armnetwork.NewSubnetsClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, virtualNetworkName, subnetName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete subnet", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteSecurityGrp(networkSecurityGroupName string, options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error) {
	client, err := armnetwork.NewSecurityGroupsClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, networkSecurityGroupName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete security group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreateSecurityGrp(networkSecurityGroupName string, parameters armnetwork.SecurityGroup, options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewSecurityGroupsClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, networkSecurityGroupName, parameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create security group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) CreateSSHKey(sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource, options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error) {
	client, err := armcompute.NewSSHPublicKeysClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return armcompute.SSHPublicKeysClientCreateResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.Create(p.b.ctx, p.resourceGrp, sshPublicKeyName, parameters, options); err != nil {
		return armcompute.SSHPublicKeysClientCreateResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create sshkey", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) DeleteSSHKey(sshPublicKeyName string, options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error) {
	client, err := armcompute.NewSSHPublicKeysClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return armcompute.SSHPublicKeysClientDeleteResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.Delete(p.b.ctx, p.resourceGrp, sshPublicKeyName, options); err != nil {
		return armcompute.SSHPublicKeysClientDeleteResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete sshkey", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreateVM(vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error) {
	client, err := armcompute.NewVirtualMachinesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, vmName, parameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create virtual machine", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteVM(vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error) {
	client, err := armcompute.NewVirtualMachinesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, vmName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete virtual machine", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteDisk(diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error) {
	client, err := armcompute.NewDisksClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, diskName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete virtual disk", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreatePubIP(publicIPAddressName string, parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewPublicIPAddressesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, publicIPAddressName, parameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create public IP", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeletePubIP(publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error) {
	client, err := armnetwork.NewPublicIPAddressesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, publicIPAddressName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete public IP", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreateNIC(networkInterfaceName string, parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewInterfacesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, networkInterfaceName, parameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create network interface", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteNIC(networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error) {
	client, err := armnetwork.NewInterfacesClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, networkInterfaceName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete network interface", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginDeleteAKS(resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error) {
	client, err := armcontainerservice.NewManagedClustersClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(p.b.ctx, p.resourceGrp, resourceName, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to delete aks", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) BeginCreateAKS(resourceName string, parameters armcontainerservice.ManagedCluster, options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error) {
	client, err := armcontainerservice.NewManagedClustersClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(p.b.ctx, p.resourceGrp, resourceName, parameters, options); err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to create aks", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (p *AzureClient) ListClusterAdminCredentials(resourceName string, options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error) {
	client, err := armcontainerservice.NewManagedClustersClient(p.b.subscriptionID, p.azureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.ListClusterAdminCredentials(p.b.ctx, p.resourceGrp, resourceName, options); err != nil {
		return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{},
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlClusterOperation,
				p.b.l.NewError(p.b.ctx, "failed to list aks credentials", "Reason", err),
			)
	} else {
		return res, nil
	}
}
