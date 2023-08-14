package cmd

// maintainer: 	Dipankar Das <dipankardas0115@gmail.com>

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
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
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		cli.Client.Metadata.Provider = utils.CLOUD_CIVO
		cli.Client.Metadata.IsHA = true

		SetDefaults(utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA)
		cli.Client.Metadata.NoWP = noWP
		cli.Client.Metadata.ClusterName = clusterName
		cli.Client.Metadata.Region = region
		cli.Client.Metadata.K8sDistro = distro

		if err := deleteApproval(); err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			os.Exit(1)
		}

		stat, err := controller.DelWorkerPlaneNode(&cli.Client)
		if err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			os.Exit(1)
		}
		cli.Client.Storage.Logger().Success(stat)
	},
}

func init() {
	deleteClusterHACivo.AddCommand(deleteNodesHACivo)

	clusterNameFlag(deleteNodesHACivo)
	noOfWPFlag(deleteNodesHACivo)
	regionFlag(deleteNodesHACivo)
	//k8sVerFlag(deleteNodesHACivo)
	distroFlag(deleteNodesHACivo)

	deleteNodesHACivo.MarkFlagRequired("name")
	deleteNodesHACivo.MarkFlagRequired("region")
}
