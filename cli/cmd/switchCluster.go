package cmd

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/civo"
	"github.com/kubesimplify/ksctl/api/local"
	"github.com/spf13/cobra"
)

var switchCluster = &cobra.Command{
	Use:     "switch-cluster",
	Aliases: []string{"switch"},
	Short:   "Use to switch between clusters",
	Long: `It is used to switch cluster with the given clusterName from user. For example:

ksctl switch-context -p <civo,local,ha-civo>  -c <clustername> -r <region> <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		switch sprovider {
		case "local":
			err := local.SwitchContext(sclusterName)
			if err != nil {
				fmt.Printf("\033[31;40m%v\033[0m\n", err)
			}
		case "civo", "ha-civo":
			if len(sregion) == 0 {
				fmt.Println(fmt.Errorf("\033[31;40mRegion is Required\033[0m\n"))
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

			err := payload.SwitchContext()
			// err := civo.SwitchContext(sclusterName, sregion)
			if err != nil {
				fmt.Printf("\033[31;40m%v\033[0m\n", err)
			}

		case "azure":
			fmt.Println("UNDER DEVELOPMENT!")
		case "aws":
			fmt.Println("UNDER DEVELOPMENT!")
		default:
			fmt.Println(fmt.Errorf("\033[31;40mINVALID provider (given)\033[0m\n"))
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
	switchCluster.Flags().StringVarP(&sclusterName, "name", "c", "", "Cluster name")
	switchCluster.Flags().StringVarP(&sregion, "region", "r", "", "Region")
	switchCluster.Flags().StringVarP(&sprovider, "provider", "p", "", "Provider")
	switchCluster.MarkFlagRequired("name")
	switchCluster.MarkFlagRequired("provider")
}
