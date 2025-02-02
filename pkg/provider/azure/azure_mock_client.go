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

//go:build testing_azure

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func ProvideClient() CloudSDK {
	return &AzureClient{}
}

type AzureClient struct {
	SubscriptionID string
	AzureTokenCred azcore.TokenCredential
	b              *Provider
}

func (mock *AzureClient) InitClient(b *Provider) error {
	mock.SubscriptionID = "XUZE"
	mock.AzureTokenCred = nil
	mock.b = b
	return nil
}

func (mock *AzureClient) ListLocations() ([]string, error) {
	return []string{"eastus2", "centralindia", "fake"}, nil
}

func (mock *AzureClient) ListKubernetesVersions() (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error) {
	return armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse{
		KubernetesVersionListResult: armcontainerservice.KubernetesVersionListResult{
			Values: []*armcontainerservice.KubernetesVersion{
				{
					Version: utilities.Ptr("1.27.1"),
				},
				{
					Version: utilities.Ptr("1.26"),
				},
				{
					Version: utilities.Ptr("1.27"),
				},
			},
		},
	}, nil
}

func (mock *AzureClient) ListVMTypes() ([]string, error) {
	return []string{"Standard_DS2_v2", "fake"}, nil
}

func (mock *AzureClient) CreateResourceGrp(parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	return armresources.ResourceGroupsClientCreateOrUpdateResponse{
		ResourceGroup: armresources.ResourceGroup{
			Name: utilities.Ptr(mock.b.resourceGroup),
		},
	}, nil
}

func (mock *AzureClient) BeginDeleteResourceGrp(options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	return &runtime.Poller[armresources.ResourceGroupsClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginCreateVirtNet(virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteVirtNet(virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginCreateSubNet(virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteSubNet(virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.SubnetsClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteSecurityGrp(networkSecurityGroupName string, options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginCreateSecurityGrp(networkSecurityGroupName string, parameters armnetwork.SecurityGroup, options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) CreateSSHKey(sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource, options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error) {
	return armcompute.SSHPublicKeysClientCreateResponse{}, nil
}

func (mock *AzureClient) DeleteSSHKey(sshPublicKeyName string, options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error) {
	return armcompute.SSHPublicKeysClientDeleteResponse{}, nil
}

func (mock *AzureClient) BeginCreateVM(vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteVM(vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error) {
	return &runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteDisk(diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error) {
	return &runtime.Poller[armcompute.DisksClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginCreatePubIP(publicIPAddressName string, parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) BeginDeletePubIP(publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginCreateNIC(networkInterfaceName string, parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteNIC(networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.InterfacesClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginDeleteAKS(resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error) {
	return &runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse]{}, nil
}

func (mock *AzureClient) BeginCreateAKS(resourceName string, parameters armcontainerservice.ManagedCluster, options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureClient) ListClusterAdminCredentials(resourceName string, options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error) {
	return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{
		CredentialResults: armcontainerservice.CredentialResults{
			Kubeconfigs: []*armcontainerservice.CredentialResult{
				{
					Name:  utilities.Ptr("fake-kubeconfig"),
					Value: []byte("fake kubeconfig"),
				},
			},
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientDeleteResponse, error) {
	// as the result is not used
	return armnetwork.SecurityGroupsClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreateNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error) {
	return armnetwork.SecurityGroupsClientCreateOrUpdateResponse{
		SecurityGroup: armnetwork.SecurityGroup{
			ID:   utilities.Ptr("XXYY"),
			Name: utilities.Ptr("fake-firewall-123"),
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelResourceGrp(ctx context.Context, poll *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armresources.ResourceGroupsClientDeleteResponse, error) {
	// as the result is not used
	return armresources.ResourceGroupsClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreateSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientCreateOrUpdateResponse, error) {
	return armnetwork.SubnetsClientCreateOrUpdateResponse{
		Subnet: armnetwork.Subnet{
			ID:   utilities.Ptr("XXYY"),
			Name: utilities.Ptr("fake-subnet-123"),
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientDeleteResponse, error) {
	// as the result is not used
	return armnetwork.SubnetsClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreateVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error) {

	return armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
		VirtualNetwork: armnetwork.VirtualNetwork{
			ID:   utilities.Ptr("XXYY"),
			Name: utilities.Ptr("fake-virt-net-123"),
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientDeleteResponse, error) {
	// as the result is not used

	return armnetwork.VirtualNetworksClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreateAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientCreateOrUpdateResponse, error) {

	return armcontainerservice.ManagedClustersClientCreateOrUpdateResponse{
		ManagedCluster: armcontainerservice.ManagedCluster{
			Name: utilities.Ptr("fake-ksctl-managed-resgrp"),
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientDeleteResponse, error) {

	return armcontainerservice.ManagedClustersClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneDelVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientDeleteResponse, error) {

	return armcompute.VirtualMachinesClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreateVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error) {

	return armcompute.VirtualMachinesClientCreateOrUpdateResponse{
		VirtualMachine: armcompute.VirtualMachine{
			Properties: &armcompute.VirtualMachineProperties{
				OSProfile: &armcompute.OSProfile{
					ComputerName: utilities.Ptr("fake-hostname"),
				},
			},
			Name: utilities.Ptr("fake-vm-123"),
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelDisk(ctx context.Context, poll *runtime.Poller[armcompute.DisksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.DisksClientDeleteResponse, error) {
	return armcompute.DisksClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreatePubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error) {

	return armnetwork.PublicIPAddressesClientCreateOrUpdateResponse{
		PublicIPAddress: armnetwork.PublicIPAddress{
			ID:   utilities.Ptr("fake-XXYYY"),
			Name: utilities.Ptr("fake-pubip"),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				IPAddress: utilities.Ptr("A.B.C.D"),
			},
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelPubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientDeleteResponse, error) {

	return armnetwork.PublicIPAddressesClientDeleteResponse{}, nil
}

func (mock *AzureClient) PollUntilDoneCreateNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientCreateOrUpdateResponse, error) {

	return armnetwork.InterfacesClientCreateOrUpdateResponse{
		Interface: armnetwork.Interface{
			ID:   utilities.Ptr("XYYY"),
			Name: utilities.Ptr("fake-nic-123"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: utilities.Ptr("192.168.1.2"),
						},
					},
				},
			},
		},
	}, nil
}

func (mock *AzureClient) PollUntilDoneDelNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientDeleteResponse, error) {

	return armnetwork.InterfacesClientDeleteResponse{}, nil
}
