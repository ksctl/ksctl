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

func managedDeleteClusterHandler(ctx context.Context, azureConfig *AzureProvider) error {

	managedClustersClient, err := getAzureManagedClusterClient(azureConfig)
	if err != nil {
		return err
	}
	azureConfig.ResourceGroupName = azureConfig.ClusterName + "-ksctl"

	if err := azureConfig.ConfigReaderManaged(); err != nil {
		return err
	}

	log.Println("Deleting AKS cluster...")

	pollerResp, err := managedClustersClient.BeginDelete(ctx, azureConfig.ResourceGroupName, azureConfig.ClusterName, nil)
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
	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: azureConfig.ClusterName, Region: azureConfig.Region, ResourceName: azureConfig.ResourceGroupName}
	printKubeconfig.Printer(false, 1)
	return nil
}

type printer struct {
	ClusterName  string
	Region       string
	ResourceName string
}

func generateResourceName(azureConfig *AzureProvider) {
	// random resourcename didnt went well as the resourcename entered by the user must be easier
	// i.e. the user has mentioned the clusterName and so resourcename will be clusterName + "-ksctl"

	// letter := "abcdefghijklmnopqrstuvwxyz0123456789"

	// noOfCharacters := 5
	var ret strings.Builder
	ret.WriteString(azureConfig.ClusterName + "-ksctl")
	// for noOfCharacters > 0 {
	// 	char := string(letter[rand.Intn(len(letter))])
	// 	ret.WriteString(char)
	// 	noOfCharacters--
	// }

	azureConfig.ResourceGroupName = ret.String()
}

func managedCreateClusterHandler(ctx context.Context, azureConfig *AzureProvider) (*armcontainerservice.ManagedCluster, error) {

	err := azureConfig.CreateResourceGroup(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Created the Resource Group " + azureConfig.ResourceGroupName)

	managedClustersClient, err := getAzureManagedClusterClient(azureConfig)
	if err != nil {
		return nil, err
	}

	// INFO: do check the CreatorUpdate function used https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/d9f41170eaf6958209047f42c8ae4d0536577422/services/compute/container_cluster.go
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

	azureConfig.ConfigWriterManagedClusteName()
	azureConfig.ConfigWriterManagedResourceName()

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	kubeconfig, err := managedClustersClient.ListClusterAdminCredentials(ctx, azureConfig.ResourceGroupName, azureConfig.ClusterName, nil)
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
	printKubeconfig = printer{ClusterName: azureConfig.ClusterName, Region: azureConfig.Region, ResourceName: azureConfig.ResourceGroupName}
	printKubeconfig.Printer(false, 0)
	return &resp.ManagedCluster, nil
}
