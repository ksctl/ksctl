package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// DelManagedCluster implements resources.CloudFactory.
func (obj *AzureProvider) DelManagedCluster(storage resources.StorageFactory) error {
	if len(azureCloudState.ManagedClusterName) == 0 {
		storage.Logger().Success("[skip] already deleted AKS cluster")
		return nil
	}
	managedClustersClient, err := obj.managedClusterClient()
	if err != nil {
		return err
	}
	storage.Logger().Print("[azure] Deleting AKS cluster...")

	pollerResp, err := managedClustersClient.BeginDelete(ctx, azureCloudState.ResourceGroupName, azureCloudState.ManagedClusterName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("[azure] Deleted the AKS cluster", azureCloudState.ManagedClusterName)
	azureCloudState.ManagedClusterName = ""
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	printKubeconfig(storage, utils.OPERATION_STATE_DELETE)

	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (obj *AzureProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	if len(azureCloudState.ManagedClusterName) != 0 {
		storage.Logger().Success("[skip] already created AKS cluster %s", azureCloudState.ManagedClusterName)
		return nil
	}

	managedClustersClient, err := obj.managedClusterClient()
	if err != nil {
		return err
	}

	azureCloudState.NoManagedNodes = noOfNodes
	azureCloudState.KubernetesVer = obj.Metadata.K8sVersion

	parameter := armcontainerservice.ManagedCluster{
		Location: to.Ptr(azureCloudState.Region),
		Properties: &armcontainerservice.ManagedClusterProperties{
			DNSPrefix:         to.Ptr("aksgosdk"),
			KubernetesVersion: to.Ptr(azureCloudState.KubernetesVer),
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					Name:              to.Ptr("askagent"),
					Count:             to.Ptr[int32](int32(noOfNodes)),
					VMSize:            to.Ptr(obj.Metadata.VmType),
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
	}
	pollerResp, err := managedClustersClient.BeginCreateOrUpdate(
		ctx,
		azureCloudState.ResourceGroupName,
		obj.Metadata.ResName,
		parameter,
		nil,
	)
	if err != nil {
		return err
	}
	azureCloudState.ManagedClusterName = obj.Metadata.ResName

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Print("[azure] Creating AKS cluster...")

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	azureCloudState.IsCompleted = true
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	kubeconfig, err := managedClustersClient.ListClusterAdminCredentials(ctx, azureCloudState.ResourceGroupName, obj.Metadata.ResName, nil)
	if err != nil {
		return err
	}
	kubeconfigStr := string(kubeconfig.Kubeconfigs[0].Value)

	if err := saveKubeconfigHelper(storage, kubeconfigStr); err != nil {
		return err
	}

	printKubeconfig(storage, utils.OPERATION_STATE_CREATE)

	storage.Logger().Success("[azure] created AKS", *resp.Name)
	return nil
}

func (obj *AzureProvider) managedClusterClient() (*armcontainerservice.ManagedClustersClient, error) {
	managedClustersClient, err := armcontainerservice.NewManagedClustersClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return nil, err
	}
	return managedClustersClient, nil
}
