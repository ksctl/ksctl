package cmd

import (
	"fmt"
	civoHandler "github.com/kubesimplify/Kubesimpctl/src/api/civo"
	"github.com/spf13/cobra"
)

var deleteClusterCivo = &cobra.Command{
	Use:   "civo",
	Short: "Use to delete a CIVO cluster",
	Long: `It is used to delete cluster of given provider. For example:

kubesimpctl delete-cluster civo 
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := civoHandler.DeleteCluster(dregion, dclusterName)
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
	deleteClusterCivo.Flags().StringVarP(&dclusterName, "name", "C", "demo", "Cluster name")
	deleteClusterCivo.Flags().StringVarP(&dregion, "region", "r", "", "Region based on different cloud providers")
	deleteClusterCivo.MarkFlagRequired("name")
	deleteClusterCivo.MarkFlagRequired("region")
}
