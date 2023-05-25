package cmd

import (
	"github.com/kubesimplify/ksctl/api/azure"
	"github.com/kubesimplify/ksctl/api/civo"
	"github.com/kubesimplify/ksctl/api/local"
	"github.com/kubesimplify/ksctl/api/logger"
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
		switch sprovider {
		case "local":
			err := local.SwitchContext(log, sclusterName)
			if err != nil {
				log.Err(err.Error())
			}
		case "civo", "ha-civo":
			if len(sregion) == 0 {
				log.Err("Region is Required")
			}
			payload := civo.CivoProvider{
				ClusterName: sclusterName,
				Region:      sregion,
			}
			if "civo" == sprovider {
				payload.HACluster = false
			} else {
				payload.HACluster = true
			}

			err := payload.SwitchContext(log)
			if err != nil {
				log.Err(err.Error())
			}

		case "azure", "ha-azure":
			if len(sregion) == 0 {
				log.Err("Region is Required")
			}
			payload := azure.AzureProvider{
				ClusterName: sclusterName,
				Region:      sregion,
			}
			if "azure" == sprovider {
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

var (
	sclusterName string
	sregion      string
	sprovider    string
)

func init() {
	rootCmd.AddCommand(switchCluster)
	switchCluster.Flags().StringVarP(&sclusterName, "name", "n", "", "Cluster name")
	switchCluster.Flags().StringVarP(&sregion, "region", "r", "", "Region")
	switchCluster.Flags().StringVarP(&sprovider, "provider", "p", "", "Provider")
	switchCluster.MarkFlagRequired("name")
	switchCluster.MarkFlagRequired("provider")
}
