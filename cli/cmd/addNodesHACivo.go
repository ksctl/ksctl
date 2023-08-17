package cmd

// maintainer: 	Dipankar Das <dipankardas0115@gmail.com>

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
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
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		cli.Client.Metadata.Provider = utils.CLOUD_CIVO
		SetDefaults(utils.CLOUD_CIVO, utils.CLUSTER_TYPE_HA)
		cli.Client.Metadata.NoWP = noWP
		cli.Client.Metadata.WorkerPlaneNodeType = nodeSizeWP
		cli.Client.Metadata.ClusterName = clusterName
		cli.Client.Metadata.Region = region
		cli.Client.Metadata.K8sDistro = distro
		cli.Client.Metadata.K8sVersion = k8sVer

		cli.Client.Metadata.IsHA = true

		if err := createApproval(cmd.Flags().Lookup("approve").Changed); err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			os.Exit(1)
		}

		stat, err := controller.AddWorkerPlaneNode(&cli.Client)
		if err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			os.Exit(1)
		}
		cli.Client.Storage.Logger().Success(stat)
	},
}

func init() {
	createClusterHACivo.AddCommand(addMoreWorkerNodesHACivo)
	clusterNameFlag(addMoreWorkerNodesHACivo)
	noOfWPFlag(addMoreWorkerNodesHACivo)
	nodeSizeWPFlag(addMoreWorkerNodesHACivo)
	regionFlag(addMoreWorkerNodesHACivo)
	k8sVerFlag(addMoreWorkerNodesHACivo)
	distroFlag(addMoreWorkerNodesHACivo)

	addMoreWorkerNodesHACivo.MarkFlagRequired("name")
	addMoreWorkerNodesHACivo.MarkFlagRequired("region")
}
