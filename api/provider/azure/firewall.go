package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/kubesimplify/ksctl/api/resources"
)

func (obj *AzureProvider) NSGClient() (*armnetwork.SecurityGroupsClient, error) {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return nsgClient, nil
}

// DelFirewall implements resources.CloudFactory.
func (*AzureProvider) DelFirewall(state resources.StorageFactory) error {
	panic("unimplemented")
}

// NewFirewall implements resources.CloudFactory.
func (*AzureProvider) NewFirewall(state resources.StorageFactory) error {
	panic("unimplemented")
}

func (obj *AzureProvider) DeleteNSG(ctx context.Context, storage resources.StorageFactory, nsgName string) error {
	nsgClient, err := obj.NSGClient()
	if err != nil {
		return err
	}

	pollerResponse, err := nsgClient.BeginDelete(ctx, obj.ResourceGroup, nsgName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	storage.Logger().Success("Deleted the nsg", nsgName)
	return nil
}

func (obj *AzureProvider) CreateNSG(ctx context.Context, storage resources.StorageFactory, nsgName string, securityRules []*armnetwork.SecurityRule) (*armnetwork.SecurityGroup, error) {
	nsgClient, err := obj.NSGClient()
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.SecurityGroup{
		Location: to.Ptr(obj.Region),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: securityRules,
		},
	}

	pollerResponse, err := nsgClient.BeginCreateOrUpdate(ctx, obj.ResourceGroup, nsgName, parameters, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	storage.Logger().Success("Created network security group", *resp.Name)
	return &resp.SecurityGroup, nil
}
