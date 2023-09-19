package cmd

// maintainer: 	Dipankar Das <dipankardas0115@gmail.com>

import (
	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
	"github.com/spf13/cobra"
)

// deleteClusterCmd represents the deleteCluster command
var deleteClusterCmd = &cobra.Command{
	Use:     "delete-cluster",
	Short:   "Use to delete a cluster",
	Aliases: []string{"delete"},
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster ["azure", "ha-<provider>", "civo", "local"]
`,
}

var deleteClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a azure managed cluster",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster azure <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		cli.Client.Metadata.Provider = CLOUD_AZURE
		SetDefaults(CLOUD_AZURE, CLUSTER_TYPE_MANG)

		deleteManaged(cmd.Flags().Lookup("approve").Changed)
	},
}

var deleteClusterCivo = &cobra.Command{
	Use:   "civo",
	Short: "Use to delete a CIVO cluster",
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster civo
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		cli.Client.Metadata.Provider = CLOUD_CIVO
		SetDefaults(CLOUD_CIVO, CLUSTER_TYPE_MANG)

		deleteManaged(cmd.Flags().Lookup("approve").Changed)

	},
}

var deleteClusterHAAzure = &cobra.Command{
	Use:   "ha-azure",
	Short: "Use to delete a HA k3s cluster in Azure",
	Long: `It is used to delete cluster with the given name from user. For example:

	ksctl delete-cluster ha-azure <arguments to civo cloud provider>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		cli.Client.Metadata.Provider = CLOUD_AZURE
		SetDefaults(CLOUD_AZURE, CLUSTER_TYPE_HA)

		deleteHA(cmd.Flags().Lookup("approve").Changed)
	},
}

var deleteClusterHACivo = &cobra.Command{
	Use:   "ha-civo",
	Short: "Use to delete a HA CIVO k3s cluster",
	Long: `It is used to delete cluster with the given name from user. For example:

ksctl delete-cluster ha-civo <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		cli.Client.Metadata.Provider = CLOUD_CIVO
		SetDefaults(CLOUD_CIVO, CLUSTER_TYPE_HA)

		deleteHA(cmd.Flags().Lookup("approve").Changed)
	},
}

var deleteClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to delete a LOCAL cluster",
	Long: `It is used to delete cluster of given provider. For example:

ksctl delete-cluster local <arguments to local/Docker provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		cli.Client.Metadata.Provider = CLOUD_LOCAL
		SetDefaults(CLOUD_LOCAL, CLUSTER_TYPE_MANG)

		deleteManaged(cmd.Flags().Lookup("approve").Changed)
	},
}

func init() {
	rootCmd.AddCommand(deleteClusterCmd)

	deleteClusterCmd.AddCommand(deleteClusterHACivo)
	deleteClusterCmd.AddCommand(deleteClusterCivo)
	deleteClusterCmd.AddCommand(deleteClusterHAAzure)
	deleteClusterCmd.AddCommand(deleteClusterAzure)
	deleteClusterCmd.AddCommand(deleteClusterLocal)

	deleteClusterAzure.MarkFlagRequired("name")
	deleteClusterAzure.MarkFlagRequired("region")
	deleteClusterCivo.MarkFlagRequired("name")
	deleteClusterCivo.MarkFlagRequired("region")
	deleteClusterHAAzure.MarkFlagRequired("name")
	deleteClusterHAAzure.MarkFlagRequired("region")
	deleteClusterHACivo.MarkFlagRequired("name")
	deleteClusterLocal.MarkFlagRequired("name")
}
