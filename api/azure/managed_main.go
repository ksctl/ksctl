package azure

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

func managedDeleteClusterHandler(ctx context.Context, azureConfig *AzureProvider) (*armcontainerservice.ManagedCluster, error) {
	// resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	// if err != nil {
	// 	return err
	// }

	// pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	// if err != nil {
	// 	return err
	// }

	// _, err = pollerResp.PollUntilDone(ctx, nil)
	// if err != nil {
	// 	return err
	// }
	// TODO: Add the delete Resource group
	// refer https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/d9f41170eaf6958209047f42c8ae4d0536577422/services/compute/container_cluster.go#L21
	// also this https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/main/sdk/resourcemanager/containerservice/managed_clusters/main.go
	return nil, nil
}

func managedCreateClusterHandler(ctx context.Context, azureConfig *AzureProvider) (*armcontainerservice.ManagedCluster, error) {

	err := azureConfig.CreateResourceGroup(ctx)
	if err != nil {
		return nil, err
	}

	managedClustersClient, err := getAzureManagedClusterClient(azureConfig)
	if err != nil {
		return nil, err
	}

	pollerResp, err := managedClustersClient.BeginCreateOrUpdate(
		ctx,
		azureConfig.ResourceGroupName,
		azureConfig.ClusterName,
		armcontainerservice.ManagedCluster{
			Location: to.Ptr(azureConfig.Region),
			Properties: &armcontainerservice.ManagedClusterProperties{
				DNSPrefix: to.Ptr("aksgosdk"),
				AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
					{
						Name:              to.Ptr("askagent"),
						Count:             to.Ptr[int32](int32(azureConfig.Spec.ManagedNodes)),
						VMSize:            to.Ptr("Standard_DS2_v2"),
						MaxPods:           to.Ptr[int32](110),
						MinCount:          to.Ptr[int32](1),
						MaxCount:          to.Ptr[int32](100),
						OSType:            to.Ptr(armcontainerservice.OSTypeLinux),
						Type:              to.Ptr(armcontainerservice.AgentPoolTypeVirtualMachineScaleSets),
						EnableAutoScaling: to.Ptr(true),
						Mode:              to.Ptr(armcontainerservice.AgentPoolModeSystem),
					},
				},
				ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
					ClientID: to.Ptr(os.Getenv("AZURE_CLIENT_ID")),
					Secret:   to.Ptr(os.Getenv("AZURE_CLIENT_SECRET")),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ManagedCluster, nil
}
