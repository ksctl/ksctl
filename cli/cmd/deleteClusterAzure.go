package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/kubesimplify/ksctl/api/azure"
	"github.com/spf13/cobra"
)

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
			ClusterName: azdcclusterName,
			HACluster:   false,
			Region:      azdcregion,
		}
		err := payload.DeleteCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED CLUSTER", "")
	},
}

var (
	azdcclusterName string
	azdcregion      string
)

func init() {
	deleteClusterCmd.AddCommand(deleteClusterAzure)
	deleteClusterAzure.Flags().StringVarP(&azdcclusterName, "name", "n", "", "Cluster name")
	deleteClusterAzure.Flags().StringVarP(&azdcregion, "region", "r", "eastus", "Region")
	deleteClusterAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterAzure.MarkFlagRequired("name")
	deleteClusterAzure.MarkFlagRequired("region")
}
