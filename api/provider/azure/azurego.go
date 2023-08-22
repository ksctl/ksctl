package azure

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	armcomputev5 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	armcontainerservicev4 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

type AzureGo interface {
	InitClient(storage resources.StorageFactory) error

	SetRegion(string)

	SetResourceGrp(string)

	ListLocations() (*runtime.Pager[armsubscriptions.ClientListLocationsResponse], error)

	ListKubernetesVersions() (armcontainerservicev4.ManagedClustersClientListKubernetesVersionsResponse, error)

	ListVMTypes() (*runtime.Pager[armcomputev5.VirtualMachineSizesClientListResponse], error)

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

type AzureGoClient struct {
	SubscriptionID string
	AzureTokenCred azcore.TokenCredential
	Region         string
	ResourceGrp    string
}

type AzureGoMockClient struct{}

func ProvideClient() AzureGo {
	return &AzureGoClient{}
}

// PollUntilDoneCreateNetInterface implements AzureGo.
func (*AzureGoClient) PollUntilDoneCreateNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

// PollUntilDoneCreatePubIP implements AzureGo.
func (*AzureGoClient) PollUntilDoneCreatePubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

// PollUntilDoneCreateVM implements AzureGo.
func (*AzureGoClient) PollUntilDoneCreateVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

// PollUntilDoneDelDisk implements AzureGo.
func (*AzureGoClient) PollUntilDoneDelDisk(ctx context.Context, poll *runtime.Poller[armcompute.DisksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.DisksClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

// PollUntilDoneDelNetInterface implements AzureGo.
func (*AzureGoClient) PollUntilDoneDelNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

// PollUntilDoneDelPubIP implements AzureGo.
func (*AzureGoClient) PollUntilDoneDelPubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

// PollUntilDoneDelVM implements AzureGo.
func (*AzureGoClient) PollUntilDoneDelVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneDelNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneCreateNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneDelResourceGrp(ctx context.Context, poll *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armresources.ResourceGroupsClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneCreateSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneDelSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneCreateVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneDelVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneCreateAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientCreateOrUpdateResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

func (obj *AzureGoClient) PollUntilDoneDelAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientDeleteResponse, error) {
	return poll.PollUntilDone(ctx, options)
}

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

func (azclient *AzureGoClient) SetResourceGrp(grp string) {
	azclient.ResourceGrp = grp
}

func (azclient *AzureGoClient) ListLocations() (*runtime.Pager[armsubscriptions.ClientListLocationsResponse], error) {
	clientFactory, err := armsubscriptions.NewClientFactory(azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return clientFactory.NewClient().NewListLocationsPager(azclient.SubscriptionID, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil}), nil
}

func (azclient *AzureGoClient) ListKubernetesVersions() (armcontainerservicev4.ManagedClustersClientListKubernetesVersionsResponse, error) {
	clientFactory, err := armcontainerservicev4.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcontainerservicev4.ManagedClustersClientListKubernetesVersionsResponse{}, fmt.Errorf("failed to create client: %v", err)
	}

	return clientFactory.NewManagedClustersClient().ListKubernetesVersions(ctx, azclient.Region, nil)
}

func (azclient *AzureGoClient) ListVMTypes() (*runtime.Pager[armcomputev5.VirtualMachineSizesClientListResponse], error) {
	clientFactory, err := armcomputev5.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return clientFactory.NewVirtualMachineSizesClient().NewListPager(azclient.Region, nil), nil
}

func (azclient *AzureGoClient) PublicIPClient() (*armnetwork.PublicIPAddressesClient, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (azclient *AzureGoClient) CreateResourceGrp(parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	client, err := armresources.NewResourceGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armresources.ResourceGroupsClientCreateOrUpdateResponse{}, err
	}
	return client.CreateOrUpdate(ctx, azclient.ResourceGrp, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteResourceGrp(options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	client, err := armresources.NewResourceGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, options)
}

func (azclient *AzureGoClient) BeginCreateVirtNet(virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewVirtualNetworksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, virtualNetworkName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteVirtNet(virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	client, err := armnetwork.NewVirtualNetworksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, virtualNetworkName, options)
}

func (azclient *AzureGoClient) BeginCreateSubNet(virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewSubnetsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, virtualNetworkName,
		subnetName, subnetParameters, options)
}

func (azclient *AzureGoClient) BeginDeleteSubNet(virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	client, err := armnetwork.NewSubnetsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, virtualNetworkName, subnetName, options)
}

func (azclient *AzureGoClient) BeginDeleteSecurityGrp(networkSecurityGroupName string, options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error) {
	client, err := armnetwork.NewSecurityGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, networkSecurityGroupName, options)
}

func (azclient *AzureGoClient) BeginCreateSecurityGrp(networkSecurityGroupName string, parameters armnetwork.SecurityGroup, options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewSecurityGroupsClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, networkSecurityGroupName, parameters, options)
}

func (azclient *AzureGoClient) CreateSSHKey(sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource, options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error) {
	client, err := armcompute.NewSSHPublicKeysClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcompute.SSHPublicKeysClientCreateResponse{}, err
	}
	return client.Create(ctx, azclient.ResourceGrp, sshPublicKeyName, parameters, options)
}

func (azclient *AzureGoClient) DeleteSSHKey(sshPublicKeyName string, options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error) {
	client, err := armcompute.NewSSHPublicKeysClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcompute.SSHPublicKeysClientDeleteResponse{}, err
	}
	return client.Delete(ctx, azclient.ResourceGrp, sshPublicKeyName, options)
}

func (azclient *AzureGoClient) BeginCreateVM(vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error) {
	client, err := armcompute.NewVirtualMachinesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, vmName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteVM(vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error) {
	client, err := armcompute.NewVirtualMachinesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, vmName, options)
}

func (azclient *AzureGoClient) BeginDeleteDisk(diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error) {
	client, err := armcompute.NewDisksClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, diskName, options)
}

func (azclient *AzureGoClient) BeginCreatePubIP(publicIPAddressName string, parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, publicIPAddressName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeletePubIP(publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error) {
	client, err := armnetwork.NewPublicIPAddressesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, publicIPAddressName, options)
}

func (azclient *AzureGoClient) BeginCreateNIC(networkInterfaceName string, parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error) {
	client, err := armnetwork.NewInterfacesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, networkInterfaceName, parameters, options)
}

func (azclient *AzureGoClient) BeginDeleteNIC(networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error) {
	client, err := armnetwork.NewInterfacesClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, networkInterfaceName, options)
}

func (azclient *AzureGoClient) BeginDeleteAKS(resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error) {
	client, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginDelete(ctx, azclient.ResourceGrp, resourceName, options)
}

func (azclient *AzureGoClient) BeginCreateAKS(resourceName string, parameters armcontainerservice.ManagedCluster, options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error) {
	client, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return client.BeginCreateOrUpdate(ctx, azclient.ResourceGrp, resourceName, parameters, options)
}

func (azclient *AzureGoClient) ListClusterAdminCredentials(resourceName string, options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error) {
	client, err := armcontainerservice.NewManagedClustersClient(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{}, err
	}
	return client.ListClusterAdminCredentials(ctx, azclient.ResourceGrp, resourceName, options)
}
