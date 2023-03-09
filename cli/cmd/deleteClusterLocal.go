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
	"github.com/spf13/cobra"
)

var deleteClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to delete a LOCAL cluster",
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster local <arguments to local/Docker provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.Logger{Verbose: true}

		if err := local.DeleteCluster(dlocalclusterName); err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("DELETED!", "")
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
