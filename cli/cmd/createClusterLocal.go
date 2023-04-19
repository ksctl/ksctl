package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	log "github.com/kubesimplify/ksctl/api/logger"

	"github.com/kubesimplify/ksctl/api/local"
	util "github.com/kubesimplify/ksctl/api/utils"
	"github.com/spf13/cobra"
)

var createClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to create a LOCAL cluster in Docker",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster local <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		cargo := local.ClusterInfoInjecter(clocalclusterName, clocalspec.ManagedNodes)

		logger.Info("Building cluster", "")
		if err := local.CreateCluster(logger, cargo); err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER", "")
	},
}

var (
	clocalclusterName string
	clocalspec        util.Machine
)

func init() {
	createClusterCmd.AddCommand(createClusterLocal)
	createClusterLocal.Flags().StringVarP(&clocalclusterName, "name", "n", "demo", "Cluster name")
	createClusterLocal.Flags().IntVarP(&clocalspec.ManagedNodes, "nodes", "N", 1, "Number of Nodes")
	createClusterLocal.Flags().BoolP("verbose", "v", true, "Verbose output")
	createClusterLocal.MarkFlagRequired("name")
}
