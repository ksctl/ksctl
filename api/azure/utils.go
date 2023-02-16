package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	util "github.com/kubesimplify/ksctl/api/utils"
	"golang.org/x/net/context"
)

type AzureOperations interface {
	CreateCluster()
	DeleteCluster()
}

type AzureInfra interface {
	CreateResourceGroup(context.Context) error
	DeleteResourceGroup(context.Context) error
	CreateVM() error
	DeleteVM() error
}

func getAzureManagedClusterClient(cred *AzureProvider) (*armcontainerservice.ManagedClustersClient, error) {

	if len(os.Getenv("AZURE_TENANT_ID")) == 0 {
		tokens, err := util.GetCred("azure")
		if err != nil {
			return nil, err
		}
		cred.SubscriptionID = tokens["subscription_id"]
		err = os.Setenv("AZURE_TENANT_ID", tokens["tenant_id"])
		if err != nil {
			return nil, err
		}
		err = os.Setenv("AZURE_CLIENT_ID", tokens["client_id"])
		if err != nil {
			return nil, err
		}
		err = os.Setenv("AZURE_CLIENT_SECRET", tokens["client_secret"])
		if err != nil {
			return nil, err
		}
	}

	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(cred.SubscriptionID, cred.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return managedClustersClient, nil
}

func getAzureResourceGroupsClient(cred *AzureProvider) (*armresources.ResourceGroupsClient, error) {
	if len(os.Getenv("AZURE_TENANT_ID")) == 0 {
		// not set
		tokens, err := util.GetCred("azure")
		if err != nil {
			return nil, err
		}
		cred.SubscriptionID = tokens["subscription_id"]
		err = os.Setenv("AZURE_TENANT_ID", tokens["tenant_id"])
		if err != nil {
			return nil, err
		}
		err = os.Setenv("AZURE_CLIENT_ID", tokens["client_id"])
		if err != nil {
			return nil, err
		}
		err = os.Setenv("AZURE_CLIENT_SECRET", tokens["client_secret"])
		if err != nil {
			return nil, err
		}
	}
	resourceGroupClient, err := armresources.NewResourceGroupsClient(cred.SubscriptionID, cred.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return resourceGroupClient, nil
}

func (obj *AzureProvider) CreateResourceGroup(ctx context.Context) error {
	resourceGroupClient, err := getAzureResourceGroupsClient(obj)
	if err != nil {
		return err
	}
	_, err = resourceGroupClient.CreateOrUpdate(
		ctx,
		obj.ResourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(obj.Region),
		},
		nil)
	if err != nil {
		return err
	}
	return nil
}
