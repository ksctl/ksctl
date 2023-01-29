/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
Anurag Kumar <contact.anurag7@gmail.com>
Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package cmd

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/aks"
	"github.com/spf13/cobra"
)

var createClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a AKS cluster in Azure",
	Long: `It is used to create cluster with the given name from user. For example:

	ksctl create-cluster azure <arguments to civo cloud provider>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// NOTE: Every time we use this command a new azure instance in created which is not correct
		// we need to move this along with other providers to top level instead of it being on command level
		azure := aks.AzureProvider{}
		
		cluster := &aks.Cluster{
			ClusterName: clusterName,
			NodeCount: nodeCount,
		}
		
		// if param is not useful later on we can remove this and add the key value to the azure provider
		params := map[string]string{
			"resourceGroupName": resourceGroupName,
			"region": region,
		}
		// change it based on requirement
		// below fields are manually filled
		fmt.Println(cluster) 
		if err := azure.Create(cluster, nil, params); err !=nil {
			fmt.Println("failed to initalize the azure struct:", err)		
		}
	},
}

var (
	clusterName string
	resourceGroupName string
	nodeCount int32
	region string
)

func init() {
	createClusterCmd.AddCommand(createClusterAzure)
	createClusterAzure.Flags().StringVarP(&clusterName, "name", "n", "", "cluster name") // what if the cluster name is same to a previously created cluster
	createClusterAzure.Flags().Int32VarP(&nodeCount, "nodes", "N",2, "Number of nodes")
	createClusterAzure.Flags().StringVarP(&resourceGroupName, "resources", "r", "", "resource group name")
	createClusterAzure.Flags().StringVarP(&region, "region", "l", "", "region")	
}
