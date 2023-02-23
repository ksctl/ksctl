package azure

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	util "github.com/kubesimplify/ksctl/api/utils"
)

func managedDeleteClusterHandler(ctx context.Context, azureConfig *AzureProvider, showMsg bool) error {
	if showMsg {
		log.Printf(`NOTE 🚨
	THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'
	`, azureConfig.ClusterName+" "+azureConfig.Config.ResourceGroupName+" "+azureConfig.Region)
		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return nil
		}
	}
	managedClustersClient, err := getAzureManagedClusterClient(azureConfig)
	if err != nil {
		return err
	}

	if err := azureConfig.ConfigReader("managed"); err != nil {
		return err
	}

	log.Println("Deleting AKS cluster...")

	pollerResp, err := managedClustersClient.BeginDelete(ctx, azureConfig.Config.ResourceGroupName, azureConfig.ClusterName, nil)
	if err != nil {
		return err
	}
	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("Deleted the AKS cluster " + azureConfig.ClusterName)
	err = azureConfig.DeleteResourceGroup(ctx)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "azure", "managed", azureConfig.ClusterName+" "+azureConfig.Config.ResourceGroupName+" "+azureConfig.Region)); err != nil {
		return err
	}
	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: azureConfig.ClusterName, Region: azureConfig.Region, ResourceName: azureConfig.Config.ResourceGroupName}
	printKubeconfig.Printer(false, 1)
	return nil
}

type printer struct {
	ClusterName  string
	Region       string
	ResourceName string
}

func managedCreateClusterHandler(ctx context.Context, azureConfig *AzureProvider) (*armcontainerservice.ManagedCluster, error) {
	defer azureConfig.ConfigWriter("managed")

	_, err := azureConfig.CreateResourceGroup(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Created the Resource Group " + azureConfig.Config.ResourceGroupName)

	managedClustersClient, err := getAzureManagedClusterClient(azureConfig)
	if err != nil {
		return nil, err
	}

	// INFO: do check the CreatorUpdate function used https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/d9f41170eaf6958209047f42c8ae4d0536577422/services/compute/container_cluster.go
	pollerResp, err := managedClustersClient.BeginCreateOrUpdate(
		ctx,
		azureConfig.Config.ResourceGroupName,
		azureConfig.ClusterName,
		armcontainerservice.ManagedCluster{
			Location: to.Ptr(azureConfig.Region),
			Properties: &armcontainerservice.ManagedClusterProperties{
				DNSPrefix: to.Ptr("aksgosdk"),
				AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
					{
						Name:              to.Ptr("askagent"),
						Count:             to.Ptr[int32](int32(azureConfig.Spec.ManagedNodes)),
						VMSize:            to.Ptr(azureConfig.Spec.Disk),
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

	kubeconfig, err := managedClustersClient.ListClusterAdminCredentials(ctx, azureConfig.Config.ResourceGroupName, azureConfig.ClusterName, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(kubeconfig.Kubeconfigs[0].Name)
	KUBECONFIG := string(kubeconfig.Kubeconfigs[0].Value)

	log.Println("NOTE: the kubeconfig to be saved has admin credentials")

	if err := azureConfig.kubeconfigWriter(KUBECONFIG); err != nil {
		return nil, err
	}

	// TODO: Try to make KubeconfigPrinter as global utility function across all Providers
	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: azureConfig.ClusterName, Region: azureConfig.Region, ResourceName: azureConfig.Config.ResourceGroupName}
	printKubeconfig.Printer(false, 0)
	return &resp.ManagedCluster, nil
}
