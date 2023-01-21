package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/local"
	"github.com/spf13/cobra"
)

var deleteClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to delete a LOCAL cluster",
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster local <arguments to local/Docker provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deleting...")
		if err := local.DeleteCluster(dlocalclusterName); err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
		}
		fmt.Printf("\033[32;40mDELETED!\033[0m\n")
	},
}

var (
	dlocalclusterName string
)

func init() {
	deleteClusterCmd.AddCommand(deleteClusterLocal)
	deleteClusterLocal.Flags().StringVarP(&dlocalclusterName, "name", "n", "demo", "Cluster name")
	deleteClusterLocal.MarkFlagRequired("name")
}
