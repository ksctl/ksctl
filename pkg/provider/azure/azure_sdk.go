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
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type CloudSDK interface {
	InitClient(b *Provider) error

	ListLocations() ([]string, error)

	ListKubernetesVersions() (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error)

	ListVMTypes() ([]string, error)

	// Resource group

	CreateResourceGrp(parameters armresources.ResourceGroup,
		options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error)

	BeginDeleteResourceGrp(
		options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error)

	// VirtualNet

	BeginCreateVirtNet(virtualNetworkName string, parameters armnetwork.VirtualNetwork,
		options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error)

	BeginDeleteVirtNet(virtualNetworkName string,
		options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error)

	// Subnet

	BeginCreateSubNet(virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet,
		options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error)

	BeginDeleteSubNet(virtualNetworkName string, subnetName string,
		options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error)

	// Firewall

	BeginDeleteSecurityGrp(networkSecurityGroupName string,
		options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error)

	BeginCreateSecurityGrp(networkSecurityGroupName string, parameters armnetwork.SecurityGroup,
		options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error)

	// SSH Key

	CreateSSHKey(sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource,
		options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error)

	DeleteSSHKey(sshPublicKeyName string,
		options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error)

	// Virtual Machine

	BeginCreateVM(vmName string, parameters armcompute.VirtualMachine,
		options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error)

	BeginDeleteVM(vmName string,
		options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error)

	// Virtual Disks

	BeginDeleteDisk(diskName string,
		options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error)

	// PublicIP

	BeginCreatePubIP(publicIPAddressName string, parameters armnetwork.PublicIPAddress,
		options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error)

	BeginDeletePubIP(publicIPAddressName string,
		options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error)

	// Network interface card

	BeginCreateNIC(networkInterfaceName string, parameters armnetwork.Interface,
		options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error)

	BeginDeleteNIC(networkInterfaceName string,
		options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error)

	// AKS

	BeginDeleteAKS(resourceName string,
		options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error)

	BeginCreateAKS(resourceName string, parameters armcontainerservice.ManagedCluster,
		options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error)

	ListClusterAdminCredentials(resourceName string,
		options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error)

	//-------------------
	//|	 Pollers
	//-------------------

	// NSG

	PollUntilDoneDelNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientDeleteResponse, error)

	PollUntilDoneCreateNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error)

	// Resource grp

	PollUntilDoneDelResourceGrp(ctx context.Context, poll *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armresources.ResourceGroupsClientDeleteResponse, error)

	// Subnet

	PollUntilDoneCreateSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientCreateOrUpdateResponse, error)

	PollUntilDoneDelSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientDeleteResponse, error)

	// virtual net

	PollUntilDoneCreateVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error)

	PollUntilDoneDelVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientDeleteResponse, error)

	// AKS

	PollUntilDoneCreateAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientCreateOrUpdateResponse, error)

	PollUntilDoneDelAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientDeleteResponse, error)

	// VM

	PollUntilDoneDelVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientDeleteResponse, error)

	PollUntilDoneCreateVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error)

	// Disk

	PollUntilDoneDelDisk(ctx context.Context, poll *runtime.Poller[armcompute.DisksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.DisksClientDeleteResponse, error)

	// Pub IP

	PollUntilDoneCreatePubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error)

	PollUntilDoneDelPubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientDeleteResponse, error)

	// net interface

	PollUntilDoneCreateNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientCreateOrUpdateResponse, error)

	PollUntilDoneDelNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientDeleteResponse, error)
}
