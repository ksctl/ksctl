package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/ha_civo"
	"github.com/spf13/cobra"
)

var deleteClusterHACivo = &cobra.Command{
	Use:   "ha-civo",
	Short: "Use to delete a HA CIVO k3s cluster",
	Long: `It is used to delete cluster with the given name from user. For example:

ksctl delete-cluster ha-civo <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ha_civo.DeleteCluster(dhcclustername, dhcregion, true)
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
		}
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
