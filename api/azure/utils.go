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

type AzureManagedState struct {
	ClusterName       string `json:"cluster_name"`
	ResourceGroupName string `json:"resource_group_name"`
}

type AzureInfra interface {
	CreateResourceGroup(context.Context) error
	DeleteResourceGroup(context.Context) error
	CreateVM() error
	DeleteVM() error
	ConfigWriterManagedClusteName() error
	ConfigWriterManagedResourceName() error
	ConfigReaderManaged() error
}

func (config *AzureProvider) ConfigWriterManagedClusteName() error {
	config.Config.ClusterName = config.ClusterName
	return util.SaveState(config.Config, "azure", config.ClusterName+" "+config.ResourceGroupName+" "+config.Region)
}

func (config *AzureProvider) ConfigWriterManagedResourceName() error {
	config.Config.ResourceGroupName = config.ResourceGroupName
	return util.SaveState(config.Config, "azure", config.ClusterName+" "+config.ResourceGroupName+" "+config.Region)
}

func (config *AzureProvider) ConfigReaderManaged() error {
	data, err := util.GetState("azure", config.ClusterName+" "+config.ResourceGroupName+" "+config.Region)
	if err != nil {
		return err
	}
	// populating the state data
	config.Config.ClusterName = data["cluster_name"]
	config.Config.ResourceGroupName = data["resource_group_name"]
	config.ClusterName = config.Config.ClusterName
	config.ResourceGroupName = config.Config.ResourceGroupName
	return nil
}

func setRequiredENV_VAR(ctx context.Context, cred *AzureProvider) error {
	tokens, err := util.GetCred("azure")
	if err != nil {
		return err
	}
	cred.SubscriptionID = tokens["subscription_id"]
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

func getAzureManagedClusterClient(cred *AzureProvider) (*armcontainerservice.ManagedClustersClient, error) {

	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(cred.SubscriptionID, cred.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return managedClustersClient, nil
}

func getAzureResourceGroupsClient(cred *AzureProvider) (*armresources.ResourceGroupsClient, error) {

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

func (obj *AzureProvider) DeleteResourceGroup(ctx context.Context) error {
	resourceGroupClient, err := getAzureResourceGroupsClient(obj)
	if err != nil {
		return err
	}
	pollerResp, err := resourceGroupClient.BeginDelete(ctx, obj.ResourceGroupName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
