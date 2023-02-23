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
	"github.com/spf13/cobra"
)

var deleteClusterHAAzure = &cobra.Command{
	Use:   "ha-azure",
	Short: "Use to delete a HA k3s cluster in Azure",
	Long: `It is used to delete cluster with the given name from user. For example:

	ksctl delete-cluster ha-azure <arguments to civo cloud provider>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		payload := &azure.AzureProvider{
			ClusterName: azhdclusterName,
			HACluster:   true,
			Region:      azhdregion,
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
	azhdclusterName string
	azhdregion      string
)

func init() {
	deleteClusterCmd.AddCommand(deleteClusterHAAzure)
	deleteClusterHAAzure.Flags().StringVarP(&azhdclusterName, "name", "n", "", "Cluster name")
	deleteClusterHAAzure.Flags().StringVarP(&azhdregion, "region", "r", "eastus", "Region")
	deleteClusterHAAzure.MarkFlagRequired("name")
	deleteClusterHAAzure.MarkFlagRequired("region")
}
