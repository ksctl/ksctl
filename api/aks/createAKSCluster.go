package aks

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

func (cluster *Cluster) CreateCluster(ctx context.Context, cred *azidentity.DefaultAzureCredential, subscriptionID string, params map[string]string) error {

	//creating the client
	client, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	// create parameters for the cluster
	clusterParam := armcontainerservice.ManagedCluster{
		Location: to.Ptr(params["region"]),
		Properties: &armcontainerservice.ManagedClusterProperties{
			DNSPrefix: to.Ptr("aksgosdk"),
			// some of the parts are manual for now and needs to be fixed/make dynamic
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					Name:              to.Ptr("askagent"),
					Count:             to.Ptr(cluster.NodeCount),
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
			// Can we do anything about this?
			ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
				ClientID: to.Ptr(os.Getenv("AZURE_CLIENT_ID")),
				Secret:   to.Ptr(os.Getenv("AZURE_CLIENT_SECRET")),
			},
		},
	}

	managedCluster, err := client.BeginCreateOrUpdate(ctx, params["resourceGroupName"], cluster.ClusterName, clusterParam, nil)
	if err != nil {
		return err
	}

	timeout, err := managedCluster.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Println("Created:", timeout)
	return nil
}
