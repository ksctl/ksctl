/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
Anurag Kumar <contact.anurag7@gmail.com>
Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package cmd

import (
	log "github.com/kubesimplify/ksctl/api/logger"

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
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := &azure.AzureProvider{
			ClusterName: azhdclusterName,
			HACluster:   true,
			Region:      azhdregion,
		}
		err := payload.DeleteCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER", "")
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
	deleteClusterHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterHAAzure.MarkFlagRequired("name")
	deleteClusterHAAzure.MarkFlagRequired("region")
}
