package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/kubesimplify/ksctl/pkg/resources"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// DelManagedCluster implements resources.CloudFactory.
func (obj *AzureProvider) DelManagedCluster(storage resources.StorageFactory) error {
	if len(azureCloudState.ManagedClusterName) == 0 {
		log.Print("skipped already deleted AKS cluster")
		return nil
	}

	pollerResp, err := obj.client.BeginDeleteAKS(azureCloudState.ManagedClusterName, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Print("Deleting AKS cluster...", "name", azureCloudState.ManagedClusterName)

	_, err = obj.client.PollUntilDoneDelAKS(ctx, pollerResp, nil)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Deleted the AKS cluster", "name", azureCloudState.ManagedClusterName)

	azureCloudState.ManagedClusterName = ""
	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}
	printKubeconfig(storage, OperationStateDelete)

	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (obj *AzureProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {
	name := obj.metadata.resName
	vmtype := obj.metadata.vmType
	obj.mxName.Unlock()
	obj.mxVMType.Unlock()

	log.Debug("Printing", "name", name, "vmtype", vmtype)

	if len(azureCloudState.ManagedClusterName) != 0 {
		log.Print("skipped already created AKS cluster", "name", azureCloudState.ManagedClusterName)
		return nil
	}

	azureCloudState.NoManagedNodes = noOfNodes
	azureCloudState.KubernetesVer = obj.metadata.k8sVersion

	parameter := armcontainerservice.ManagedCluster{
		Location: to.Ptr(azureCloudState.Region),
		Properties: &armcontainerservice.ManagedClusterProperties{
			DNSPrefix:         to.Ptr("aksgosdk"),
			KubernetesVersion: to.Ptr(azureCloudState.KubernetesVer),
			NetworkProfile: &armcontainerservice.NetworkProfile{
				NetworkPlugin: to.Ptr[armcontainerservice.NetworkPlugin](armcontainerservice.NetworkPlugin(obj.metadata.cni)),
			},
			AutoUpgradeProfile: &armcontainerservice.ManagedClusterAutoUpgradeProfile{
				NodeOSUpgradeChannel: to.Ptr[armcontainerservice.NodeOSUpgradeChannel](armcontainerservice.NodeOSUpgradeChannelNodeImage),
				UpgradeChannel:       to.Ptr[armcontainerservice.UpgradeChannel](armcontainerservice.UpgradeChannelPatch),
			},
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

	log.Debug("Printing", "AKSConfig", parameter)

	pollerResp, err := obj.client.BeginCreateAKS(name, parameter, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	azureCloudState.ManagedClusterName = name

	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	log.Print("Creating AKS cluster...")

	resp, err := obj.client.PollUntilDoneCreateAKS(ctx, pollerResp, nil)
	if err != nil {
		return log.NewError(err.Error())
	}

	azureCloudState.IsCompleted = true
	if err := saveStateHelper(storage); err != nil {
		return log.NewError(err.Error())
	}

	kubeconfig, err := obj.client.ListClusterAdminCredentials(name, nil)
	if err != nil {
		return log.NewError(err.Error())
	}
	kubeconfigStr := string(kubeconfig.Kubeconfigs[0].Value)

	if err := saveKubeconfigHelper(storage, kubeconfigStr); err != nil {
		return log.NewError(err.Error())
	}

	printKubeconfig(storage, OperationStateCreate)

	log.Success("created AKS", "name", *resp.Name)
	return nil
}

func (obj *AzureProvider) GetKubeconfigPath() string {
	return generatePath(UtilClusterPath, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
}
