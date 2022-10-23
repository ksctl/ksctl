package cmd

import (
	"fmt"
	"github.com/kubesimplify/Kubesimpctl/src/api/local"
	"github.com/kubesimplify/Kubesimpctl/src/api/payload"
	"github.com/spf13/cobra"
)

var createClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to create a LOCAL cluster in Docker",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster local <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		cargo := local.ClusterInfoInjecter(clocalclusterName, clocalspec.Nodes)
		fmt.Println("Building...")
		if err := local.CreateCluster(cargo); err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\nDeleting configs...\n", err)
			err := local.DeleteCluster(clocalclusterName)
			if err != nil {
				return
			}
			return
		}
		fmt.Printf("\033[32;40mCREATED!\033[0m\n")
	},
}

var (
	clocalclusterName string
	clocalspec        payload.Machine
)

func init() {
	createClusterCmd.AddCommand(createClusterLocal)
	createClusterLocal.Flags().StringVarP(&clocalclusterName, "name", "c", "demo", "Cluster name")
	createClusterLocal.Flags().IntVarP(&clocalspec.Nodes, "nodes", "n", 1, "Number of Nodes")
	createClusterLocal.MarkFlagRequired("name")
	//createClusterLocal.MarkFlagRequired("nodes")
}
