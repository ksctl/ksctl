package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type AzureGo interface {
	NSGClient() (*armnetwork.SecurityGroupsClient, error)
	ManagedClusterClient() (*armcontainerservice.ManagedClustersClient, error)
	ResourceGroupsClient() (*armresources.ResourceGroupsClient, error)
	VirtualNetworkClient() (*armnetwork.VirtualNetworksClient, error)
	SubnetClient() (*armnetwork.SubnetsClient, error)
	SSHKeyClient() (*armcompute.SSHPublicKeysClient, error)
	VirtualMachineClient() (*armcompute.VirtualMachinesClient, error)
	DiskClient() (*armcompute.DisksClient, error)
	NetInterfaceClient() (*armnetwork.InterfacesClient, error)
	PublicIPClient() (*armnetwork.PublicIPAddressesClient, error)
}

type AzureGoClient struct {
}
