package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
)

type AzureGo interface {
	InitClient(storage resources.StorageFactory) error

	SetRegion(string)

	NSGClient() (*armnetwork.SecurityGroupsClient, error)

	AKSClient() (*armcontainerservice.ManagedClustersClient, error)

	ListLocations() (*runtime.Pager[armsubscriptions.ClientListLocationsResponse], error)

	ListKubernetesVersions(ctx context.Context) (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error)

	ListVMTypes() (*runtime.Pager[armcompute.VirtualMachineSizesClientListResponse], error)

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

	BeginDeleteDisk(ctx context.Context, client *armcompute.DisksClient, resourceGroupName string,
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

	BeginCreateAKS(ctx context.Context, client *armcontainerservice.ManagedClustersClient,
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

// TODO: moved the data to here
type AzureGoClient struct {
	SubscriptionID string
	AzureTokenCred azcore.TokenCredential
	Region         string
}

type AzureGoMockClient struct{}

//	func ProvideMockAzureClient() AzureGo {
//		return &AzureGoMockClient{}
//	}

func ProvideClient() AzureGo {
	return &AzureGoClient{}
}

// TODO: remove the other variant
func (obj *AzureGoClient) setRequiredENV_VAR(storage resources.StorageFactory, ctx context.Context) error {

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

	storage.Logger().Warn(msg)

	tokens, err := utils.GetCred(storage, "azure")
	if err != nil {
		return err
	}

	obj.SubscriptionID = tokens["subscription_id"]

	err = os.Setenv("AZURE_SUBSCRIPTION_ID", tokens["subscription_id"])
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_TENANT_ID", tokens["tenant_id"])
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_CLIENT_ID", tokens["client_id"])
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_CLIENT_SECRET", tokens["client_secret"])
	if err != nil {
		return err
	}
	return nil
}

func (azclient *AzureGoClient) InitClient(storage resources.StorageFactory) error {
	err := azclient.setRequiredENV_VAR(storage, ctx)
	if err != nil {
		return err
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	azclient.AzureTokenCred = cred
	return nil
}

func (azclient *AzureGoClient) SetRegion(reg string) {
	azclient.Region = reg
}

func (azclient *AzureGoClient) NSGClient() (*armnetwork.SecurityGroupsClient, error) {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return nsgClient, nil
}

func (azclient *AzureGoClient) AKSClient() (*armcontainerservice.ManagedClustersClient, error) {
	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return managedClustersClient, nil
}

func (azclient *AzureGoClient) ListLocations() (*runtime.Pager[armsubscriptions.ClientListLocationsResponse], error) {
	clientFactory, err := armsubscriptions.NewClientFactory(azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return clientFactory.NewClient().NewListLocationsPager(azclient.SubscriptionID, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil}), nil
}

func (azclient *AzureGoClient) ListKubernetesVersions(ctx context.Context) (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error) {
	clientFactory, err := armcontainerservice.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse{}, fmt.Errorf("failed to create client: %v", err)
	}

	return clientFactory.NewManagedClustersClient().ListKubernetesVersions(ctx, azclient.Region, nil)
}

func (azclient *AzureGoClient) ListVMTypes() (*runtime.Pager[armcompute.VirtualMachineSizesClientListResponse], error) {
	clientFactory, err := armcompute.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return clientFactory.NewVirtualMachineSizesClient().NewListPager(azclient.Region, nil), nil
}

func (azclient *AzureGoClient) ResourceGroupsClient() (*armresources.ResourceGroupsClient, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}

	return resourceGroupClient, nil
}

func (azclient *AzureGoClient) VirtualNetworkClient() (*armnetwork.VirtualNetworksClient, error) {
	vnetClient, err := armnetwork.NewVirtualNetworksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return vnetClient, nil
}

func (azclient *AzureGoClient) SubnetClient() (*armnetwork.SubnetsClient, error) {
	subnetClient, err := armnetwork.NewSubnetsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return subnetClient, nil
}

func (azclient *AzureGoClient) SSHKeyClient() (*armcompute.SSHPublicKeysClient, error) {
	client, err := armcompute.NewSSHPublicKeysClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (azclient *AzureGoClient) VirtualMachineClient() (*armcompute.VirtualMachinesClient, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return vmClient, nil
}

func (azclient *AzureGoClient) DiskClient() (*armcompute.DisksClient, error) {
	diskClient, err := armcompute.NewDisksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return diskClient, nil
}

func (azclient *AzureGoClient) NetInterfaceClient() (*armnetwork.InterfacesClient, error) {
	client, err := armnetwork.NewInterfacesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (azclient *AzureGoClient) PublicIPClient() (*armnetwork.PublicIPAddressesClient, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (azclient *AzureGoClient) CreateResourceGrp(ctx context.Context, client *armresources.ResourceGroupsClient, resourceGroupName string, parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	return client.CreateOrUpdate(ctx, resourceGroupName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteResourceGrp(ctx context.Context, client *armresources.ResourceGroupsClient, resourceGroupName string, options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, options)
}

func (azclient *AzureGoClient) BeginCreateVirtNet(ctx context.Context, client *armnetwork.VirtualNetworksClient, resourceGroupName string, virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteVirtNet(ctx context.Context, client *armnetwork.VirtualNetworksClient, resourceGroupName string, virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, virtualNetworkName, options)
}

func (azclient *AzureGoClient) BeginCreateSubNet(ctx context.Context, client *armnetwork.SubnetsClient, resourceGroupName string, virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName,
		subnetName, subnetParameters, options)
}

func (azclient *AzureGoClient) BeginDeleteSubNet(ctx context.Context, client *armnetwork.SubnetsClient, resourceGroupName string, virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, virtualNetworkName, subnetName, options)
}

func (azclient *AzureGoClient) BeginDeleteSecurityGrp(ctx context.Context, client *armnetwork.SecurityGroupsClient, resourceGroupName string, networkSecurityGroupName string, options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, networkSecurityGroupName, options)
}

func (azclient *AzureGoClient) BeginCreateSecurityGrp(ctx context.Context, client *armnetwork.SecurityGroupsClient, resourceGroupName string, networkSecurityGroupName string, parameters armnetwork.SecurityGroup, options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, networkSecurityGroupName, parameters, options)
}

func (azclient *AzureGoClient) CreateSSHKey(ctx context.Context, client *armcompute.SSHPublicKeysClient, resourceGroupName string, sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource, options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error) {
	return client.Create(ctx, resourceGroupName, sshPublicKeyName, parameters, options)
}

func (azclient *AzureGoClient) DeleteSSHKey(ctx context.Context, client *armcompute.SSHPublicKeysClient, resourceGroupName string, sshPublicKeyName string, options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error) {
	return client.Delete(ctx, resourceGroupName, sshPublicKeyName, options)
}

func (azclient *AzureGoClient) BeginCreateVM(ctx context.Context, client *armcompute.VirtualMachinesClient, resourceGroupName string, vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, vmName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteVM(ctx context.Context, client *armcompute.VirtualMachinesClient, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, vmName, options)
}

func (azclient *AzureGoClient) BeginDeleteDisk(ctx context.Context, client *armcompute.DisksClient, resourceGroupName string, diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, diskName, options)
}

func (azclient *AzureGoClient) BeginCreatePubIP(ctx context.Context, client *armnetwork.PublicIPAddressesClient, resourceGroupName string, publicIPAddressName string, parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, publicIPAddressName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeletePubIP(ctx context.Context, client *armnetwork.PublicIPAddressesClient, resourceGroupName string, publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, publicIPAddressName, options)
}

func (azclient *AzureGoClient) BeginCreateNIC(ctx context.Context, client *armnetwork.InterfacesClient, resourceGroupName string, networkInterfaceName string, parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, networkInterfaceName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteNIC(ctx context.Context, client *armnetwork.InterfacesClient, resourceGroupName string, networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, networkInterfaceName, options)
}

func (azclient *AzureGoClient) BeginDeleteAKS(ctx context.Context, client *armcontainerservice.ManagedClustersClient, resourceGroupName string, resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error) {
	return client.BeginDelete(ctx, resourceGroupName, resourceName, options)
}

func (azclient *AzureGoClient) BeginCreateAKS(ctx context.Context, client *armcontainerservice.ManagedClustersClient, resourceGroupName string, resourceName string, parameters armcontainerservice.ManagedCluster, options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error) {
	return client.BeginCreateOrUpdate(ctx, resourceGroupName, resourceName, parameters, options)
}

func (azclient *AzureGoClient) ListClusterAdminCredentials(ctx context.Context, client *armcontainerservice.ManagedClustersClient, resourceGroupName string, resourceName string, options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error) {
	return client.ListClusterAdminCredentials(ctx, resourceGroupName, resourceName, options)
}

func (azclient *AzureGoClient) PollUntilDone(ctx context.Context, p *runtime.Poller[any], options *runtime.PollUntilDoneOptions) (any, error) {
	return p.PollUntilDone(ctx, options)
}
