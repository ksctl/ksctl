package cmd

import (
	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/provider/logger"
	"github.com/spf13/cobra"
)

var switchCluster = &cobra.Command{
	Use:     "switch-cluster",
	Aliases: []string{"switch"},
	Short:   "Use to switch between clusters",
	Long: `It is used to switch cluster with the given clusterName from user. For example:

ksctl switch-context -p <civo,local,ha-civo,ha-azure,azure>  -n <clustername> -r <region> <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {

		log := logger.Logger{}
		switch provider {
		case "local":
			err := local.SwitchContext(log, clusterName)
			if err != nil {
				log.Err(err.Error())
			}
		case "civo", "ha-civo":
			if len(region) == 0 {
				log.Err("Region is Required")
			}
			payload := civo.CivoProvider{
				ClusterName: clusterName,
				Region:      region,
			}
			if "civo" == provider {
				payload.HACluster = false
			} else {
				payload.HACluster = true
			}

			err := payload.SwitchContext(log)
			if err != nil {
				log.Err(err.Error())
			}

		case "azure", "ha-azure":
			if len(region) == 0 {
				log.Err("Region is Required")
			}
			payload := azure.AzureProvider{
				ClusterName: clusterName,
				Region:      region,
			}
			if "azure" == provider {
				payload.HACluster = false
			} else {
				payload.HACluster = true
			}
			err := payload.SwitchContext(log)
			if err != nil {
				log.Err(err.Error())
			}
		case "aws":
			log.Warn("UNDER DEVELOPMENT")
		default:
			log.Err("invalid provider!")
		}
	},
}

func init() {
	rootCmd.AddCommand(switchCluster)
	switchCluster.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	switchCluster.Flags().StringVarP(&region, "region", "r", "", "Region")
	switchCluster.Flags().StringVarP(&provider, "provider", "p", "", "Provider")
	switchCluster.MarkFlagRequired("name")
	switchCluster.MarkFlagRequired("provider")
}
