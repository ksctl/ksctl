package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

import (
	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/provider/local"
	log "github.com/kubesimplify/ksctl/api/provider/logger"
	"github.com/spf13/cobra"
)

// deleteClusterCmd represents the deleteCluster command
var deleteClusterCmd = &cobra.Command{
	Use:     "delete-cluster",
	Short:   "Use to delete a cluster",
	Aliases: []string{"delete"},
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster ["azure", "ha-<provider>", "civo", "local"]
`,
}

var deleteClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a azure managed cluster",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster azure <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := &azure.AzureProvider{
			ClusterName: clusterName,
			HACluster:   false,
			Region:      region,
		}
		err := payload.DeleteCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED CLUSTER")
	},
}

var deleteClusterCivo = &cobra.Command{
	Use:   "civo",
	Short: "Use to delete a CIVO cluster",
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster civo
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := civo.CivoProvider{
			ClusterName: clusterName,
			Region:      region,
			HACluster:   false,
		}
		err := payload.DeleteCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED CLUSTER")
	},
}

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
			ClusterName: clusterName,
			HACluster:   true,
			Region:      region,
		}
		err := payload.DeleteCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED CLUSTER")
	},
}

var deleteClusterHACivo = &cobra.Command{
	Use:   "ha-civo",
	Short: "Use to delete a HA CIVO k3s cluster",
	Long: `It is used to delete cluster with the given name from user. For example:

ksctl delete-cluster ha-civo <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}
		payload := civo.CivoProvider{
			ClusterName: clusterName,
			Region:      region,
			HACluster:   true,
		}
		err := payload.DeleteCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED CLUSTER")
	},
}

var deleteClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to delete a LOCAL cluster",
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster local <arguments to local/Docker provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		if err := local.DeleteCluster(logger, clusterName); err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED!")
	},
}

func init() {
	rootCmd.AddCommand(deleteClusterCmd)

	deleteClusterCmd.AddCommand(deleteClusterAzure)
	deleteClusterAzure.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	deleteClusterAzure.Flags().StringVarP(&region, "region", "r", "eastus", "Region")
	deleteClusterAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterAzure.MarkFlagRequired("name")
	deleteClusterAzure.MarkFlagRequired("region")

	deleteClusterCmd.AddCommand(deleteClusterCivo)
	deleteClusterCivo.Flags().StringVarP(&clusterName, "name", "n", "demo", "Cluster name")
	deleteClusterCivo.Flags().StringVarP(&region, "region", "r", "", "Region based on different cloud providers")
	deleteClusterCivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterCivo.MarkFlagRequired("name")
	deleteClusterCivo.MarkFlagRequired("region")

	deleteClusterCmd.AddCommand(deleteClusterHAAzure)
	deleteClusterHAAzure.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	deleteClusterHAAzure.Flags().StringVarP(&region, "region", "r", "eastus", "Region")
	deleteClusterHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterHAAzure.MarkFlagRequired("name")
	deleteClusterHAAzure.MarkFlagRequired("region")

	deleteClusterCmd.AddCommand(deleteClusterHACivo)
	deleteClusterHACivo.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	deleteClusterHACivo.Flags().StringVarP(&region, "region", "r", "LON1", "Region")
	deleteClusterHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterHACivo.MarkFlagRequired("name")

	deleteClusterCmd.AddCommand(deleteClusterLocal)
	deleteClusterLocal.Flags().StringVarP(&clusterName, "name", "n", "demo", "Cluster name")
	deleteClusterLocal.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterLocal.MarkFlagRequired("name")
}
