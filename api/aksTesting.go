package main

//"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

import (
	"fmt"

	az "github.com/kubesimplify/ksctl/api/azure"
)

func main() {

	fmt.Println(az.Credentials())
	// ctx := context.Background()
	// // export 3 variables to system
	// // AZURE_CLIENT_ID - after reg to your app in AD(Application id )
	// // AZURE_CLIENT_SECRET - create it
	// // AZURE_TENANT_ID
	// // export variable_name=<content>
	// // need to be authorized
	//
	// // checks for export varible and validates it
	// // if not then checks for azure cli
	// cred, err := azidentity.NewDefaultAzureCredential(nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//    subscriptionID := "your subscription id"

	// // create client resource group
	// // requires subsciption ID, cred from above, optional field
	// resourceClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // setting up parameters for our resource group
	// param := armresources.ResourceGroup{
	//        Location: to.Ptr("centralindia"),
	//    }

	//    resourceGroup, err := resourceClient.CreateOrUpdate(ctx, "abcd", param, nil)
	//    if err != nil {
	//        log.Fatalf("cannot create resource group: %+v", err)
	//    }
	// fmt.Println("Resource Group name: ", *resourceGroup.Name)

	// //creating managed cluster client
	// client, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // create parameters for the cluster
	// clusterParam := armcontainerservice.ManagedCluster{
	// 	Location: to.Ptr("centralindia"),
	// 	Properties: &armcontainerservice.ManagedClusterProperties{
	// 		DNSPrefix: to.Ptr("aksgosdk"),
	// 		AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
	// 			{
	// 				Name:              to.Ptr("askagent"),
	// 				Count:             to.Ptr[int32](1),
	// 				VMSize:            to.Ptr("Standard_DS2_v2"),
	// 				MaxPods:           to.Ptr[int32](110),
	// 				MinCount:          to.Ptr[int32](1),
	// 				MaxCount:          to.Ptr[int32](100),
	// 				OSType:            to.Ptr(armcontainerservice.OSTypeLinux),
	// 				Type:              to.Ptr(armcontainerservice.AgentPoolTypeVirtualMachineScaleSets),
	// 				EnableAutoScaling: to.Ptr(true),
	// 				Mode:              to.Ptr(armcontainerservice.AgentPoolModeSystem),
	// 			},
	// 		},
	// 		ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
	// 			ClientID: to.Ptr(os.Getenv("AZURE_CLIENT_ID")),
	// 			Secret: to.Ptr(os.Getenv("AZURE_CLIENT_SECRET")),
	// 		},
	// 	},
	// }

	// // create a managed cluster
	// cluster, err := client.BeginCreateOrUpdate(ctx, "abcd", "a", clusterParam, nil)
	// if err != nil {
	// 	fmt.Println("cluster creation err : ", err)
	// }

	// timeout , err := cluster.PollUntilDone(ctx, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// log.Println("Managed Cluster: ", timeout)
}
