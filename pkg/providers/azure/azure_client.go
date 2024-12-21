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
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/ksctl/ksctl/pkg/types"
)

func ProvideClient() AzureGo {
	return &AzureClient{}
}

type AzureClient struct {
	SubscriptionID string
	AzureTokenCred azcore.TokenCredential
	Region         string
	ResourceGrp    string
}

func (*AzureClient) PollUntilDoneCreateNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (*AzureClient) PollUntilDoneCreatePubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (*AzureClient) PollUntilDoneCreateVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (*AzureClient) PollUntilDoneDelDisk(ctx context.Context, poll *runtime.Poller[armcompute.DisksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.DisksClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (*AzureClient) PollUntilDoneDelNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (*AzureClient) PollUntilDoneDelPubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (*AzureClient) PollUntilDoneDelVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneDelNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneCreateNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneDelResourceGrp(ctx context.Context, poll *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armresources.ResourceGroupsClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneCreateSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneDelSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneCreateVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneDelVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneCreateAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientCreateOrUpdateResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) PollUntilDoneDelAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientDeleteResponse, error) {
	res, err := poll.PollUntilDone(ctx, options)
	if err != nil {
		return res, ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(azureCtx, "failed waiting", "Reason", err),
		)
	}
	return res, nil
}

func (obj *AzureClient) setRequiredENV_VAR(storage types.StorageFactory, ctx context.Context) error {

	envTenant := os.Getenv("AZURE_TENANT_ID")
	envSub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	envClientid := os.Getenv("AZURE_CLIENT_ID")
	envClientsec := os.Getenv("AZURE_CLIENT_SECRET")

	if len(envTenant) != 0 &&
		len(envSub) != 0 &&
		len(envClientid) != 0 &&
		len(envClientsec) != 0 {

		obj.SubscriptionID = envSub
		return nil
	}

	msg := "environment vars not set:"
	if len(envTenant) == 0 {
		msg = msg + " AZURE_TENANT_ID"
	}

	if len(envSub) == 0 {
		msg = msg + " AZURE_SUBSCRIPTION_ID"
	}

	if len(envClientid) == 0 {
		msg = msg + " AZURE_CLIENT_ID"
	}

	if len(envClientsec) == 0 {
		msg = msg + " AZURE_CLIENT_SECRET"
	}

	log.Debug(azureCtx, msg)

	credentials, err := storage.ReadCredentials(consts.CloudAzure)
	if err != nil {
		return err
	}
	if credentials.Azure == nil {
		return ksctlErrors.ErrNilCredentials.Wrap(
			log.NewError(azureCtx, "no credentials was found"),
		)
	}

	obj.SubscriptionID = credentials.Azure.SubscriptionID

	err = os.Setenv("AZURE_SUBSCRIPTION_ID", credentials.Azure.SubscriptionID)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(err)
	}

	err = os.Setenv("AZURE_TENANT_ID", credentials.Azure.TenantID)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(err)
	}

	err = os.Setenv("AZURE_CLIENT_ID", credentials.Azure.ClientID)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(err)
	}

	err = os.Setenv("AZURE_CLIENT_SECRET", credentials.Azure.ClientSecret)
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(err)
	}
	return nil
}

func (azclient *AzureClient) InitClient(storage types.StorageFactory) error {
	err := azclient.setRequiredENV_VAR(storage, azureCtx)
	if err != nil {
		return err
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(azureCtx, "defaultAzureCredential", "Reason", err),
		)
	}
	azclient.AzureTokenCred = cred
	return nil
}

func (azclient *AzureClient) SetRegion(reg string) {
	azclient.Region = reg
}

func (azclient *AzureClient) SetResourceGrp(grp string) {
	azclient.ResourceGrp = grp
}

func (azclient *AzureClient) ListLocations() ([]string, error) {
	clientFactory, err := armsubscriptions.NewClientFactory(azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(azureCtx, "failed in azure client", "Reason", err),
		)
	}
	pager := clientFactory.NewClient().NewListLocationsPager(azclient.SubscriptionID, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil})

	var validReg []string
	for pager.More() {
		page, err := pager.NextPage(azureCtx)
		if err != nil {
			return nil, ksctlErrors.ErrInternal.Wrap(
				log.NewError(azureCtx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			validReg = append(validReg, *v.Name)
		}
	}
	return validReg, nil
}

func (azclient *AzureClient) ListKubernetesVersions() (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error) {
	clientFactory, err := armcontainerservice.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse{},
			ksctlErrors.ErrInternal.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := clientFactory.
		NewManagedClustersClient().
		ListKubernetesVersions(azureCtx, azclient.Region, nil); err != nil {
		return res, ksctlErrors.ErrInternal.Wrap(
			log.NewError(azureCtx, "failed to get managed kubernetes versions", "Reason", err),
		)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) ListVMTypes() ([]string, error) {
	clientFactory, err := armcompute.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrInternal.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}
	pager := clientFactory.NewVirtualMachineSizesClient().NewListPager(azclient.Region, nil)

	var validSize []string
	for pager.More() {

		page, err := pager.NextPage(azureCtx)
		if err != nil {
			return nil, ksctlErrors.ErrInternal.Wrap(
				log.NewError(azureCtx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			validSize = append(validSize, *v.Name)
		}
	}
	return validSize, nil
}

func (azclient *AzureClient) PublicIPClient() (*armnetwork.PublicIPAddressesClient, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}
	return client, nil
}

func (azclient *AzureClient) CreateResourceGrp(parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	client, err := armresources.NewResourceGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armresources.ResourceGroupsClientCreateOrUpdateResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}
	if res, err := client.CreateOrUpdate(azureCtx, azclient.ResourceGrp, parameters, options); err != nil {
		return armresources.ResourceGroupsClientCreateOrUpdateResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create resource group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteResourceGrp(options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	client, err := armresources.NewResourceGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}
	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete resource group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreateVirtNet(virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewVirtualNetworksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, virtualNetworkName, parameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create virtual network", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteVirtNet(virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	client, err := armnetwork.NewVirtualNetworksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, virtualNetworkName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete virtual network", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreateSubNet(virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewSubnetsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, virtualNetworkName,
		subnetName, subnetParameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create subnet", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteSubNet(virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	client, err := armnetwork.NewSubnetsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, virtualNetworkName, subnetName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete subnet", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteSecurityGrp(networkSecurityGroupName string, options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error) {
	client, err := armnetwork.NewSecurityGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, networkSecurityGroupName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete security group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreateSecurityGrp(networkSecurityGroupName string, parameters armnetwork.SecurityGroup, options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewSecurityGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, networkSecurityGroupName, parameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create security group", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) CreateSSHKey(sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource, options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error) {
	client, err := armcompute.NewSSHPublicKeysClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcompute.SSHPublicKeysClientCreateResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.Create(azureCtx, azclient.ResourceGrp, sshPublicKeyName, parameters, options); err != nil {
		return armcompute.SSHPublicKeysClientCreateResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create sshkey", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) DeleteSSHKey(sshPublicKeyName string, options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error) {
	client, err := armcompute.NewSSHPublicKeysClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcompute.SSHPublicKeysClientDeleteResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.Delete(azureCtx, azclient.ResourceGrp, sshPublicKeyName, options); err != nil {
		return armcompute.SSHPublicKeysClientDeleteResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete sshkey", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreateVM(vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error) {
	client, err := armcompute.NewVirtualMachinesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, vmName, parameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create virtual machine", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteVM(vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error) {
	client, err := armcompute.NewVirtualMachinesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, vmName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete virtual machine", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteDisk(diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error) {
	client, err := armcompute.NewDisksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, diskName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete virtual disk", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreatePubIP(publicIPAddressName string, parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, publicIPAddressName, parameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create public IP", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeletePubIP(publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, publicIPAddressName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete public IP", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreateNIC(networkInterfaceName string, parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewInterfacesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, networkInterfaceName, parameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create network interface", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteNIC(networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error) {
	client, err := armnetwork.NewInterfacesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, networkInterfaceName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete network interface", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginDeleteAKS(resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error) {
	client, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginDelete(azureCtx, azclient.ResourceGrp, resourceName, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to delete aks", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) BeginCreateAKS(resourceName string, parameters armcontainerservice.ManagedCluster, options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error) {
	client, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.BeginCreateOrUpdate(azureCtx, azclient.ResourceGrp, resourceName, parameters, options); err != nil {
		return nil,
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to create aks", "Reason", err),
			)
	} else {
		return res, nil
	}
}

func (azclient *AzureClient) ListClusterAdminCredentials(resourceName string, options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error) {
	client, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed in azure client", "Reason", err),
			)
	}

	if res, err := client.ListClusterAdminCredentials(azureCtx, azclient.ResourceGrp, resourceName, options); err != nil {
		return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{},
			ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(azureCtx, "failed to list aks credentials", "Reason", err),
			)
	} else {
		return res, nil
	}
}
