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

var addMoreWorkerNodesHACivo = &cobra.Command{
	Use:   "add-nodes",
	Short: "Use to add more worker nodes in HA CIVO k3s cluster",
	Long: `It is used to add nodes to worker nodes in cluster with the given name from user. For example:

ksctl create-cluster ha-civo add-nodes <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}
		payload := civo.CivoProvider{
			ClusterName: awhcclustername,
			Region:      awhcregion,
			HACluster:   true,
			Spec: utils.Machine{
				Disk:          awhcnodesize,
				HAWorkerNodes: awhcnowp,
			},
		}
		err := payload.AddMoreWorkerNodes(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("ADDED WORKKER NODE(s)")
	},
}

var (
	// aw hc -> add workernodes to ha-civo
	awhcregion      string
	awhcclustername string
	awhcnodesize    string
	awhcnowp        int
)

func init() {
	createClusterHACivo.AddCommand(addMoreWorkerNodesHACivo)
	addMoreWorkerNodesHACivo.Flags().StringVarP(&awhcclustername, "name", "n", "", "Cluster name")
	addMoreWorkerNodesHACivo.Flags().StringVarP(&awhcnodesize, "nodeSize", "s", "g3.small", "Node size")
	addMoreWorkerNodesHACivo.Flags().StringVarP(&awhcregion, "region", "r", "", "Region")
	addMoreWorkerNodesHACivo.Flags().IntVarP(&awhcnowp, "worker-nodes", "w", 1, "no of worker nodes to be added")
	addMoreWorkerNodesHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	addMoreWorkerNodesHACivo.MarkFlagRequired("name")
	addMoreWorkerNodesHACivo.MarkFlagRequired("region")
}
