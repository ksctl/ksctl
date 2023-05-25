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
	"github.com/kubesimplify/ksctl/api/utils"
	"github.com/spf13/cobra"
)

var deleteNodesHACivo = &cobra.Command{
	Use:   "delete-nodes",
	Short: "Use to delete a HA CIVO k3s cluster",
	Long: `It is used to delete cluster with the given name from user. For example:

ksctl delete-cluster ha-civo delete-nodes <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := civo.CivoProvider{
			ClusterName: dwhcclustername,
			Region:      dwhcregion,
			HACluster:   true,
			Spec: utils.Machine{
				HAWorkerNodes: dwhcwp,
			},
		}
		err := payload.DeleteSomeWorkerNodes(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED WorkerNode(s)")
	},
}

var (
	// dw hc -> delete worker-nodes to ha-civo
	dwhcregion      string
	dwhcclustername string
	dwhcwp          int
)

func init() {
	deleteClusterHACivo.AddCommand(deleteNodesHACivo)
	deleteNodesHACivo.Flags().StringVarP(&dwhcclustername, "name", "n", "", "Cluster name")
	deleteNodesHACivo.Flags().StringVarP(&dwhcregion, "region", "r", "LON1", "Region")
	deleteNodesHACivo.Flags().IntVarP(&dwhcwp, "worker-nodes", "w", 1, "no of worker nodes to delete")
	deleteNodesHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	deleteNodesHACivo.MarkFlagRequired("name")
}
