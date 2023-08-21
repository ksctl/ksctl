package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type AzureGo interface {
	NSGClient() (*armnetwork.SecurityGroupsClient, error)

	AKSClient() (*armcontainerservice.ManagedClustersClient, error)

	AKSClientFactory(subscriptionID string, credential azcore.TokenCredential,
		options *arm.ClientOptions) (*armcontainerservice.ClientFactory, error)

	NewClientFactory(credential azcore.TokenCredential,
		options *arm.ClientOptions) (*armsubscriptions.ClientFactory, error)

	// TODO: make these 2 have do all the things and if fails just return error

	ListLocations(subscriptionID string, client *armsubscriptions.Client,
		options *armsubscriptions.ClientListLocationsOptions) *runtime.Pager[armsubscriptions.ClientListLocationsResponse]

	ListKubernetesVersions(ctx context.Context, client *armcontainerservice.ManagedClustersClient,
		location string, options *armcontainerservice.ManagedClustersClientListKubernetesVersionsOptions) (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error)

	ListVMTypes(location string, client *armcompute.VirtualMachineSizesClient,
		options *armcompute.VirtualMachineSizesClientListOptions) *runtime.Pager[armcompute.VirtualMachineSizesClientListResponse]

	ResourceGroupsClient() (*armresources.ResourceGroupsClient, error)
	VirtualNetworkClient() (*armnetwork.VirtualNetworksClient, error)
	SubnetClient() (*armnetwork.SubnetsClient, error)
	SSHKeyClient() (*armcompute.SSHPublicKeysClient, error)
	VirtualMachineClient() (*armcompute.VirtualMachinesClient, error)
	DiskClient() (*armcompute.DisksClient, error)
	NetInterfaceClient() (*armnetwork.InterfacesClient, error)
	PublicIPClient() (*armnetwork.PublicIPAddressesClient, error)

	// Resource group

	CreateResourceGrp(ctx context.Context, client *armresources.ResourceGroupsClient,
		resourceGroupName string, parameters armresources.ResourceGroup,
		options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error)

	BeginDeleteResourceGrp(ctx context.Context, client *armresources.ResourceGroupsClient,
		resourceGroupName string,
		options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error)

	// VirtualNet

	BeginCreateVirtNet(ctx context.Context, client *armnetwork.VirtualNetworksClient,
		resourceGroupName string, virtualNetworkName string,
		parameters armnetwork.VirtualNetwork,
		options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error)

	BeginDeleteVirtNet(ctx context.Context, client *armnetwork.VirtualNetworksClient,
		resourceGroupName string, virtualNetworkName string,
		options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error)

	// Subnet

	BeginCreateSubNet(ctx context.Context, client *armnetwork.SubnetsClient,
		resourceGroupName string, virtualNetworkName string, subnetName string,
		subnetParameters armnetwork.Subnet,
		options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error)

	BeginDeleteSubNet(ctx context.Context, client *armnetwork.SubnetsClient,
		resourceGroupName string, virtualNetworkName string,
		subnetName string,
		options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error)

	// Firewall

	BeginDeleteSecurityGrp(ctx context.Context, client *armnetwork.SecurityGroupsClient,
		resourceGroupName string, networkSecurityGroupName string,
		options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error)

	BeginCreateSecurityGrp(ctx context.Context, client *armnetwork.SecurityGroupsClient,
		resourceGroupName string, networkSecurityGroupName string,
		parameters armnetwork.SecurityGroup,
		options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error)

	// SSH Key

	CreateSSHKey(ctx context.Context, client *armcompute.SSHPublicKeysClient,
		resourceGroupName string, sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource,
		options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error)

	DeleteSSHKey(ctx context.Context, client *armcompute.SSHPublicKeysClient,
		resourceGroupName string, sshPublicKeyName string,
		options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error)

	// Virtual Machine

	BeginCreateVM(ctx context.Context, client *armcompute.VirtualMachinesClient, resourceGroupName string,
		vmName string, parameters armcompute.VirtualMachine,
		options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error)

	BeginDeleteVM(ctx context.Context, client *armcompute.VirtualMachinesClient, resourceGroupName string, vmName string,
		options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error)

	// Virtual Disks

	BeginDelete(ctx context.Context, client *armcompute.DisksClient, resourceGroupName string,
		diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error)

	// PublicIP

	BeginCreatePubIP(ctx context.Context, client *armnetwork.PublicIPAddressesClient, resourceGroupName string,
		publicIPAddressName string, parameters armnetwork.PublicIPAddress,
		options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error)

	BeginDeletePubIP(ctx context.Context, client *armnetwork.PublicIPAddressesClient, resourceGroupName string,
		publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error)

	// Network interface card

	BeginCreateNIC(ctx context.Context, client *armnetwork.InterfacesClient,
		resourceGroupName string, networkInterfaceName string, parameters armnetwork.Interface,
		options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error)

	BeginDeleteNIC(ctx context.Context, client *armnetwork.InterfacesClient, resourceGroupName string,
		networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error)

	// AKS

	BeginDeleteAKS(ctx context.Context, client *armcontainerservice.ManagedClustersClient, resourceGroupName string,
		resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error)

	BeginCreateOrUpdate(ctx context.Context, client *armcontainerservice.ManagedClustersClient,
		resourceGroupName string, resourceName string, parameters armcontainerservice.ManagedCluster,
		options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error)

	ListClusterAdminCredentials(ctx context.Context, client *armcontainerservice.ManagedClustersClient, resourceGroupName string,
		resourceName string,
		options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error)

	// Polling

	Poller
}

type Poller interface {
	PollUntilDone(ctx context.Context, p *runtime.Poller[any], options *runtime.PollUntilDoneOptions) (any, error)
}

//type AzureGoClient struct{}
//
//type AzureGoMockClient struct{}
//
//func ProvideMockAzureClient() AzureGo {
//	return &AzureGoMockClient{}
//}
//
//func ProvideClient() AzureGo {
//	return &AzureGoClient{}
//}
