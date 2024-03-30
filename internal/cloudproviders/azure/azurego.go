package azure

import (
	"context"
	"fmt"
	"os"

	"github.com/ksctl/ksctl/pkg/helpers/consts"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	//"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	//"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/ksctl/ksctl/pkg/resources"
)

func ProvideClient() AzureGo {
	return &AzureGoClient{}
}

func ProvideMockClient() AzureGo {
	return &AzureGoMockClient{}
}

type AzureGo interface {
	InitClient(storage resources.StorageFactory) error

	SetRegion(string)

	SetResourceGrp(string)

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

type AzureGoClient struct {
	SubscriptionID string
	AzureTokenCred azcore.TokenCredential
	Region         string
	ResourceGrp    string
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

	log.Warn(msg)

	credentials, err := storage.ReadCredentials(consts.CloudAzure)
	if err != nil {
		return err
	}

	obj.SubscriptionID = credentials.Azure.SubscriptionID

	err = os.Setenv("AZURE_SUBSCRIPTION_ID", credentials.Azure.SubscriptionID)
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_TENANT_ID", credentials.Azure.TenantID)
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_CLIENT_ID", credentials.Azure.ClientID)
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_CLIENT_SECRET", credentials.Azure.ClientSecret)
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

func (azclient *AzureGoClient) ListLocations() ([]string, error) {
	clientFactory, err := armsubscriptions.NewClientFactory(azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	pager := clientFactory.NewClient().NewListLocationsPager(azclient.SubscriptionID, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil})

	var validReg []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			validReg = append(validReg, *v.Name)
		}
	}
	return validReg, nil
}

func (azclient *AzureGoClient) ListKubernetesVersions() (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error) {
	clientFactory, err := armcontainerservice.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse{}, fmt.Errorf("failed to create client: %v", err)
	}

	return clientFactory.NewManagedClustersClient().ListKubernetesVersions(ctx, azclient.Region, nil)
}

func (azclient *AzureGoClient) ListVMTypes() ([]string, error) {
	clientFactory, err := armcompute.NewClientFactory(azclient.SubscriptionID, azclient.AzureTokenCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	pager := clientFactory.NewVirtualMachineSizesClient().NewListPager(azclient.Region, nil)

	var validSize []string
	for pager.More() {

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			validSize = append(validSize, *v.Name)
		}
	}
	return validSize, nil
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

type AzureGoMockClient struct {
	SubscriptionID string
	AzureTokenCred azcore.TokenCredential
	Region         string
	ResourceGrp    string
}

func (mock *AzureGoMockClient) InitClient(storage resources.StorageFactory) error {
	mock.SubscriptionID = "XUZE"
	mock.AzureTokenCred = nil
	return nil
}

func (mock *AzureGoMockClient) SetRegion(s string) {
	mock.Region = s
}

func (mock *AzureGoMockClient) SetResourceGrp(s string) {
	mock.ResourceGrp = s
}

func (mock *AzureGoMockClient) ListLocations() ([]string, error) {
	return []string{"eastus2", "centralindia", "fake"}, nil
}

func (mock *AzureGoMockClient) ListKubernetesVersions() (armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse, error) {
	return armcontainerservice.ManagedClustersClientListKubernetesVersionsResponse{
		KubernetesVersionListResult: armcontainerservice.KubernetesVersionListResult{
			Values: []*armcontainerservice.KubernetesVersion{
				&armcontainerservice.KubernetesVersion{
					Version: to.Ptr("1.27.1"),
				},
				&armcontainerservice.KubernetesVersion{
					Version: to.Ptr("1.26"),
				},
				&armcontainerservice.KubernetesVersion{
					Version: to.Ptr("1.27"),
				},
			},
		},
	}, nil
}

func (mock *AzureGoMockClient) ListVMTypes() ([]string, error) {
	return []string{"Standard_DS2_v2", "fake"}, nil
}

func (mock *AzureGoMockClient) CreateResourceGrp(parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	return armresources.ResourceGroupsClientCreateOrUpdateResponse{
		ResourceGroup: armresources.ResourceGroup{
			Name: to.Ptr(mock.ResourceGrp),
		},
	}, nil
}

func (mock *AzureGoMockClient) BeginDeleteResourceGrp(options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	return &runtime.Poller[armresources.ResourceGroupsClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginCreateVirtNet(virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteVirtNet(virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (*runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginCreateSubNet(virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteSubNet(virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.SubnetsClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteSecurityGrp(networkSecurityGroupName string, options *armnetwork.SecurityGroupsClientBeginDeleteOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginCreateSecurityGrp(networkSecurityGroupName string, parameters armnetwork.SecurityGroup, options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) CreateSSHKey(sshPublicKeyName string, parameters armcompute.SSHPublicKeyResource, options *armcompute.SSHPublicKeysClientCreateOptions) (armcompute.SSHPublicKeysClientCreateResponse, error) {
	return armcompute.SSHPublicKeysClientCreateResponse{}, nil
}

func (mock *AzureGoMockClient) DeleteSSHKey(sshPublicKeyName string, options *armcompute.SSHPublicKeysClientDeleteOptions) (armcompute.SSHPublicKeysClientDeleteResponse, error) {
	return armcompute.SSHPublicKeysClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) BeginCreateVM(vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteVM(vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error) {
	return &runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteDisk(diskName string, options *armcompute.DisksClientBeginDeleteOptions) (*runtime.Poller[armcompute.DisksClientDeleteResponse], error) {
	return &runtime.Poller[armcompute.DisksClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginCreatePubIP(publicIPAddressName string, parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeletePubIP(publicIPAddressName string, options *armnetwork.PublicIPAddressesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginCreateNIC(networkInterfaceName string, parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteNIC(networkInterfaceName string, options *armnetwork.InterfacesClientBeginDeleteOptions) (*runtime.Poller[armnetwork.InterfacesClientDeleteResponse], error) {
	return &runtime.Poller[armnetwork.InterfacesClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginDeleteAKS(resourceName string, options *armcontainerservice.ManagedClustersClientBeginDeleteOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], error) {
	return &runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse]{}, nil
}

func (mock *AzureGoMockClient) BeginCreateAKS(resourceName string, parameters armcontainerservice.ManagedCluster, options *armcontainerservice.ManagedClustersClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], error) {
	return &runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse]{}, nil
}

func (mock *AzureGoMockClient) ListClusterAdminCredentials(resourceName string, options *armcontainerservice.ManagedClustersClientListClusterAdminCredentialsOptions) (armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse, error) {
	return armcontainerservice.ManagedClustersClientListClusterAdminCredentialsResponse{
		CredentialResults: armcontainerservice.CredentialResults{
			Kubeconfigs: []*armcontainerservice.CredentialResult{
				&armcontainerservice.CredentialResult{
					Name:  to.Ptr("fake-kubeconfig"),
					Value: []byte("fake kubeconfig"),
				},
			},
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientDeleteResponse, error) {
	// as the result is not used
	return armnetwork.SecurityGroupsClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreateNSG(ctx context.Context, poll *runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error) {
	return armnetwork.SecurityGroupsClientCreateOrUpdateResponse{
		SecurityGroup: armnetwork.SecurityGroup{
			ID:   to.Ptr("XXYY"),
			Name: to.Ptr("fake-firewall-123"),
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelResourceGrp(ctx context.Context, poll *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armresources.ResourceGroupsClientDeleteResponse, error) {
	// as the result is not used
	return armresources.ResourceGroupsClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreateSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientCreateOrUpdateResponse, error) {
	return armnetwork.SubnetsClientCreateOrUpdateResponse{
		Subnet: armnetwork.Subnet{
			ID:   to.Ptr("XXYY"),
			Name: to.Ptr("fake-subnet-123"),
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelSubNet(ctx context.Context, poll *runtime.Poller[armnetwork.SubnetsClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.SubnetsClientDeleteResponse, error) {
	// as the result is not used
	return armnetwork.SubnetsClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreateVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error) {

	return armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
		VirtualNetwork: armnetwork.VirtualNetwork{
			ID:   to.Ptr("XXYY"),
			Name: to.Ptr("fake-virt-net-123"),
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelVirtNet(ctx context.Context, poll *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.VirtualNetworksClientDeleteResponse, error) {
	// as the result is not used

	return armnetwork.VirtualNetworksClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreateAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientCreateOrUpdateResponse, error) {

	return armcontainerservice.ManagedClustersClientCreateOrUpdateResponse{
		ManagedCluster: armcontainerservice.ManagedCluster{
			Name: to.Ptr("fake-ksctl-managed-resgrp"),
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelAKS(ctx context.Context, poll *runtime.Poller[armcontainerservice.ManagedClustersClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcontainerservice.ManagedClustersClientDeleteResponse, error) {

	return armcontainerservice.ManagedClustersClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientDeleteResponse, error) {

	return armcompute.VirtualMachinesClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreateVM(ctx context.Context, poll *runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error) {

	return armcompute.VirtualMachinesClientCreateOrUpdateResponse{
		VirtualMachine: armcompute.VirtualMachine{
			Properties: &armcompute.VirtualMachineProperties{
				OSProfile: &armcompute.OSProfile{
					ComputerName: to.Ptr("fake-hostname"),
				},
			},
			Name: to.Ptr("fake-vm-123"),
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelDisk(ctx context.Context, poll *runtime.Poller[armcompute.DisksClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armcompute.DisksClientDeleteResponse, error) {
	return armcompute.DisksClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreatePubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error) {

	return armnetwork.PublicIPAddressesClientCreateOrUpdateResponse{
		PublicIPAddress: armnetwork.PublicIPAddress{
			ID:   to.Ptr("fake-XXYYY"),
			Name: to.Ptr("fake-pubip"),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				IPAddress: to.Ptr("A.B.C.D"),
			},
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelPubIP(ctx context.Context, poll *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.PublicIPAddressesClientDeleteResponse, error) {

	return armnetwork.PublicIPAddressesClientDeleteResponse{}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneCreateNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientCreateOrUpdateResponse, error) {

	return armnetwork.InterfacesClientCreateOrUpdateResponse{
		Interface: armnetwork.Interface{
			ID:   to.Ptr("XYYY"),
			Name: to.Ptr("fake-nic-123"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					&armnetwork.InterfaceIPConfiguration{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: to.Ptr("192.168.1.2"),
						},
					},
				},
			},
		},
	}, nil
}

func (mock *AzureGoMockClient) PollUntilDoneDelNetInterface(ctx context.Context, poll *runtime.Poller[armnetwork.InterfacesClientDeleteResponse], options *runtime.PollUntilDoneOptions) (armnetwork.InterfacesClientDeleteResponse, error) {

	return armnetwork.InterfacesClientDeleteResponse{}, nil
}
