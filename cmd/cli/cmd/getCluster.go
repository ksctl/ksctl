package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
*/

import (
	"os"

	"github.com/kubesimplify/ksctl/pkg/utils/consts"

	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/spf13/cobra"
)

type printer struct {
	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`
	Provider    string `json:"provider"`
}

// viewClusterCmd represents the viewCluster command
var getClusterCmd = &cobra.Command{
	Use:     "get-clusters",
	Aliases: []string{"get"},
	Short:   "Use to get clusters",
	Long: `It is used to view clusters. For example:

ksctl get-clusters `,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		if len(provider) == 0 {
			provider = "all"
		}
		SetRequiredFeatureFlags(cmd)
		cli.Client.Metadata.Provider = consts.KsctlCloud(provider)
		stat, err := controller.GetCluster(&cli.Client)
		if err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			os.Exit(1)
		}
		cli.Client.Storage.Logger().Success(stat)
	},
}

func init() {
	rootCmd.AddCommand(getClusterCmd)
	getClusterCmd.Flags().StringVarP(&provider, "provider", "p", "", "Provider")
}
