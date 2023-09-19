package cmd

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
	"github.com/spf13/cobra"
)

var switchCluster = &cobra.Command{
	Use:     "switch-cluster",
	Aliases: []string{"switch"},
	Short:   "Use to switch between clusters",
	Long: `It is used to switch cluster with the given ClusterName from user. For example:

ksctl switch-context -p <civo,local,civo-ha,azure-ha,azure>  -n <clustername> -r <region> <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		cli.Client.Metadata.ClusterName = clusterName
		cli.Client.Metadata.Region = region

		switch provider {
		case string(CLOUD_LOCAL):
			cli.Client.Metadata.Provider = CLOUD_LOCAL

		case string(CLUSTER_TYPE_HA) + "-" + string(CLOUD_CIVO):
			cli.Client.Metadata.Provider = CLOUD_CIVO
			cli.Client.Metadata.IsHA = true

		case string(CLOUD_CIVO):
			cli.Client.Metadata.Provider = CLOUD_CIVO

		case string(CLUSTER_TYPE_HA) + "-" + string(CLOUD_AZURE):
			cli.Client.Metadata.Provider = CLOUD_AZURE
			cli.Client.Metadata.IsHA = true

		case string(CLOUD_AZURE):
			cli.Client.Metadata.Provider = CLOUD_AZURE
		}

		stat, err := controller.SwitchCluster(&cli.Client)
		if err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			os.Exit(1)
		}
		cli.Client.Storage.Logger().Success(stat)
	},
}

func init() {
	rootCmd.AddCommand(switchCluster)
	clusterNameFlag(switchCluster)
	regionFlag(switchCluster)
	switchCluster.Flags().StringVarP(&provider, "provider", "p", "", "Provider")
	switchCluster.MarkFlagRequired("name")
	switchCluster.MarkFlagRequired("provider")
}
