package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/kubesimplify/ksctl/pkg/resources"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// DelManagedCluster implements resources.CloudFactory.
func (obj *AzureProvider) DelManagedCluster(storage resources.StorageFactory) error {
	if len(azureCloudState.ManagedClusterName) == 0 {
		storage.Logger().Success("[skip] already deleted AKS cluster")
		return nil
	}

	pollerResp, err := obj.client.BeginDeleteAKS(azureCloudState.ManagedClusterName, nil)
	if err != nil {
		return err
	}
	storage.Logger().Print("[azure] Deleting AKS cluster...")

	_, err = obj.client.PollUntilDoneDelAKS(ctx, pollerResp, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("[azure] Deleted the AKS cluster", azureCloudState.ManagedClusterName)
	azureCloudState.ManagedClusterName = ""
	if err := saveStateHelper(storage); err != nil {
		return err
	}
	printKubeconfig(storage, OPERATION_STATE_DELETE)

	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (obj *AzureProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {
	name := obj.metadata.resName
	vmtype := obj.metadata.vmType
	obj.mxName.Unlock()
	obj.mxVMType.Unlock()

	if len(azureCloudState.ManagedClusterName) != 0 {
		storage.Logger().Success("[skip] already created AKS cluster %s", azureCloudState.ManagedClusterName)
		return nil
	}

	azureCloudState.NoManagedNodes = noOfNodes
	azureCloudState.KubernetesVer = obj.metadata.k8sVersion

	parameter := armcontainerservice.ManagedCluster{
		Location: to.Ptr(azureCloudState.Region),
		Properties: &armcontainerservice.ManagedClusterProperties{
			DNSPrefix:         to.Ptr("aksgosdk"),
			KubernetesVersion: to.Ptr(azureCloudState.KubernetesVer),
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					Name:              to.Ptr("askagent"),
					Count:             to.Ptr[int32](int32(noOfNodes)),
					VMSize:            to.Ptr(vmtype),
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
	pollerResp, err := obj.client.BeginCreateAKS(name, parameter, nil)
	if err != nil {
		return err
	}
	azureCloudState.ManagedClusterName = name

	if err := saveStateHelper(storage); err != nil {
		return err
	}

	storage.Logger().Print("[azure] Creating AKS cluster...")

	resp, err := obj.client.PollUntilDoneCreateAKS(ctx, pollerResp, nil)
	if err != nil {
		return err
	}

	azureCloudState.IsCompleted = true
	if err := saveStateHelper(storage); err != nil {
		return err
	}

	kubeconfig, err := obj.client.ListClusterAdminCredentials(name, nil)
	if err != nil {
		return err
	}
	kubeconfigStr := string(kubeconfig.Kubeconfigs[0].Value)

	if err := saveKubeconfigHelper(storage, kubeconfigStr); err != nil {
		return err
	}

	printKubeconfig(storage, OPERATION_STATE_CREATE)

	storage.Logger().Success("[azure] created AKS", *resp.Name)
	return nil
}
