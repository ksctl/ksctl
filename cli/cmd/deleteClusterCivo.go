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
			ClusterName: dclusterName,
			Region:      dregion,
			HACluster:   false,
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
	dclusterName string
	dregion      string
)

func init() {
	deleteClusterCmd.AddCommand(deleteClusterCivo)
	deleteClusterCivo.Flags().StringVarP(&dclusterName, "name", "n", "demo", "Cluster name")
	deleteClusterCivo.Flags().StringVarP(&dregion, "region", "r", "", "Region based on different cloud providers")
	deleteClusterCivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteClusterCivo.MarkFlagRequired("name")
	deleteClusterCivo.MarkFlagRequired("region")
}
