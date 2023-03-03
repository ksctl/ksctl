package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/azure"
	"github.com/kubesimplify/ksctl/api/utils"
	"github.com/spf13/cobra"
)

var deleteNodesHAAzure = &cobra.Command{
	Use:   "delete-nodes",
	Short: "Use to delete a HA CIVO k3s cluster",
	Long: `It is used to delete cluster with the given name from user. For example:

ksctl delete-cluster ha-azure delete-nodes <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		payload := azure.AzureProvider{
			ClusterName: azdhdclustername,
			Region:      azdhdregion,
			HACluster:   true,
			Spec: utils.Machine{
				HAWorkerNodes: azdhdwp,
			},
		}
		err := payload.DeleteSomeWorkerNodes()
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
		}
	},
}

var (
	// dw hc -> delete worker-nodes to ha-civo
	azdhdregion      string
	azdhdclustername string
	azdhdwp          int
)

func init() {
	deleteClusterHAAzure.AddCommand(deleteNodesHAAzure)
	deleteNodesHAAzure.Flags().StringVarP(&azdhdclustername, "name", "n", "", "Cluster name")
	deleteNodesHAAzure.Flags().StringVarP(&azdhdregion, "region", "r", "eastus", "Region")
	deleteNodesHAAzure.Flags().IntVarP(&azdhdwp, "worker-nodes", "w", 1, "no of worker nodes to delete")
	deleteNodesHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteNodesHAAzure.MarkFlagRequired("name")
	deleteNodesHAAzure.MarkFlagRequired("region")
}
