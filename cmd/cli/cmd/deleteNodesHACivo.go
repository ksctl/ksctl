package cmd

// maintainer: 	Dipankar Das <dipankardas0115@gmail.com>

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
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
		SetRequiredFeatureFlags(cmd)
		cli.Client.Metadata.Provider = CLOUD_CIVO
		cli.Client.Metadata.IsHA = true

		SetDefaults(CLOUD_CIVO, CLUSTER_TYPE_HA)
		cli.Client.Metadata.NoWP = noWP
		cli.Client.Metadata.ClusterName = clusterName
		cli.Client.Metadata.Region = region
		cli.Client.Metadata.K8sDistro = KsctlKubernetes(distro)

		if err := deleteApproval(cmd.Flags().Lookup("approve").Changed); err != nil {
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
