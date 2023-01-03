package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

import (
	"fmt"

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
		payload := civo.CivoProvider{
			ClusterName: dclusterName,
			Region:      dregion,
			HACluster:   false,
		}
		err := payload.DeleteCluster()
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mDELETED!\033[0m\n")
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
	deleteClusterCivo.MarkFlagRequired("name")
	deleteClusterCivo.MarkFlagRequired("region")
}
