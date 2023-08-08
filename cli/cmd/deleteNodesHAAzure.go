// package cmd
//
// /*
// Kubesimplify
// @maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
// 				Anurag Kumar <contact.anurag7@gmail.com>
// 				Avinesh Tripathi <avineshtripathi1@gmail.com>
// */
// import (
// 	log "github.com/kubesimplify/ksctl/api/provider/logger"
//
// 	"github.com/kubesimplify/ksctl/api/provider/azure"
// 	"github.com/kubesimplify/ksctl/api/provider/utils"
// 	"github.com/spf13/cobra"
// )
//
// var deleteNodesHAAzure = &cobra.Command{
// 	Use:   "delete-nodes",
// 	Short: "Use to delete a HA CIVO k3s cluster",
// 	Long: `It is used to delete cluster with the given name from user. For example:
//
// ksctl delete-cluster ha-azure delete-nodes <arguments to civo cloud provider>
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		isSet := cmd.Flags().Lookup("verbose").Changed
// 		logger := log.Logger{Verbose: true}
// 		if !isSet {
// 			logger.Verbose = false
// 		}
//
// 		payload := azure.AzureProvider{
// 			ClusterName: clusterName,
// 			Region:      region,
// 			HACluster:   true,
// 			Spec: utils.Machine{
// 				HAWorkerNodes: noWorkerNodes,
// 			},
// 		}
// 		err := payload.DeleteSomeWorkerNodes(logger)
// 		if err != nil {
// 			logger.Err(err.Error())
// 			return
// 		}
// 		logger.Info("DELETED WorkerNode(s)")
// 	},
// }
//
// func init() {
// 	deleteClusterHAAzure.AddCommand(deleteNodesHAAzure)
// 	deleteNodesHAAzure.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
// 	deleteNodesHAAzure.Flags().StringVarP(&region, "region", "r", "eastus", "Region")
// 	deleteNodesHAAzure.Flags().IntVarP(&noWorkerNodes, "worker-nodes", "w", 1, "no of worker nodes to delete")
// 	deleteNodesHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
// 	deleteNodesHAAzure.MarkFlagRequired("name")
// 	deleteNodesHAAzure.MarkFlagRequired("region")
// }
