package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/kubesimplify/ksctl/api/civo"
	"github.com/spf13/cobra"
)

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
			ClusterName: dhcclustername,
			Region:      dhcregion,
			HACluster:   true,
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
	// d hc -> delete to ha-civo
	dhcregion      string
	dhcclustername string
)

func init() {
	deleteClusterCmd.AddCommand(deleteClusterHACivo)
	deleteClusterHACivo.Flags().StringVarP(&dhcclustername, "name", "n", "", "Cluster name")
	deleteClusterHACivo.Flags().StringVarP(&dhcregion, "region", "r", "LON1", "Region")
	deleteClusterHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterHACivo.MarkFlagRequired("name")
}
