/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
Anurag Kumar <contact.anurag7@gmail.com>
Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package cmd

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/azure"
	util "github.com/kubesimplify/ksctl/api/utils"
	"github.com/spf13/cobra"
)

var createClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a AKS cluster in Azure",
	Long: `It is used to create cluster with the given name from user. For example:

	ksctl create-cluster azure <arguments to civo cloud provider>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		payload := &azure.AzureProvider{
			ClusterName: azmcclusterName,
			HACluster:   false,
			Region:      azmcregion,
			Spec: util.Machine{
				ManagedNodes: azmcnodeCount,
				Disk:         azmcsize,
			},
		}
		err := payload.CreateCluster()
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mCREATED!\033[0m\n")
	},
}

var (
	azmcclusterName string
	azmcnodeCount   int
	azmcsize        string
	azmcregion      string
)

func init() {
	createClusterCmd.AddCommand(createClusterAzure)
	createClusterAzure.Flags().StringVarP(&azmcclusterName, "name", "n", "", "Cluster name")
	createClusterAzure.Flags().StringVarP(&azmcsize, "node-size", "s", "Standard_DS2_v2", "Node size")
	createClusterAzure.Flags().StringVarP(&azmcregion, "region", "r", "eastus", "Region")
	createClusterAzure.Flags().IntVarP(&azmcnodeCount, "nodes", "N", 1, "Number of Nodes")
	createClusterAzure.MarkFlagRequired("name")
}
