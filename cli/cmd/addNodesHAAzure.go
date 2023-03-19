package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/kubesimplify/ksctl/api/azure"
	"github.com/kubesimplify/ksctl/api/utils"
	"github.com/spf13/cobra"
)

var addMoreWorkerNodesHAAzure = &cobra.Command{
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

		payload := azure.AzureProvider{
			ClusterName: azhncclustername,
			Region:      azhncregion,
			HACluster:   true,
			Spec: utils.Machine{
				Disk:          azhncnodesize,
				HAWorkerNodes: azhncwp,
			},
		}
		err := payload.AddMoreWorkerNodes(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("ADDED WORKKER NODE(s)", "")
	},
}

var (
	// aw hc -> add workernodes to ha-civo
	azhncregion      string
	azhncclustername string
	azhncnodesize    string
	azhncwp          int
)

func init() {
	createClusterHAAzure.AddCommand(addMoreWorkerNodesHAAzure)
	addMoreWorkerNodesHAAzure.Flags().StringVarP(&azhncclustername, "name", "n", "", "Cluster name")
	addMoreWorkerNodesHAAzure.Flags().StringVarP(&azhncnodesize, "node-size", "s", "Standard_F2s", "Node size")
	addMoreWorkerNodesHAAzure.Flags().StringVarP(&azhncregion, "region", "r", "", "Region")
	addMoreWorkerNodesHAAzure.Flags().IntVarP(&azhncwp, "worker-nodes", "w", 1, "no of worker nodes to be added")
	addMoreWorkerNodesHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	addMoreWorkerNodesHAAzure.MarkFlagRequired("name")
	addMoreWorkerNodesHAAzure.MarkFlagRequired("region")
}
