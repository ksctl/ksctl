package cmd

// maintainer: 	Dipankar Das <dipankardas0115@gmail.com>

import (
	control_pkg "github.com/kubesimplify/ksctl/api/controllers"

	. "github.com/kubesimplify/ksctl/api/utils/consts"
	"github.com/spf13/cobra"
)

// createClusterCmd represents the createCluster command
var createClusterCmd = &cobra.Command{
	Use:     "create-cluster",
	Short:   "Use to create a cluster",
	Aliases: []string{"create"},
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster ["azure", "gcp", "aws", "local"]
`,
}

var createClusterAws = &cobra.Command{
	Use:   "aws",
	Short: "Use to create a EKS cluster in AWS",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster aws <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {},
}

var createClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a AKS cluster in Azure",
	Long: `It is used to create cluster with the given name from user. For example:

	ksctl create-cluster azure <arguments to civo cloud provider>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		SetRequiredFeatureFlags(cmd)
		cli.Client.Metadata.Provider = CLOUD_AZURE
		SetDefaults(CLOUD_AZURE, CLUSTER_TYPE_MANG)
		createManaged(cmd.Flags().Lookup("approve").Changed)
	},
}

var createClusterCivo = &cobra.Command{
	Use:   "civo",
	Short: "Use to create a CIVO k3s cluster",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster civo <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		SetRequiredFeatureFlags(cmd)
		cli.Client.Metadata.Provider = CLOUD_CIVO
		SetDefaults(CLOUD_CIVO, CLUSTER_TYPE_MANG)
		createManaged(cmd.Flags().Lookup("approve").Changed)
	},
}

var createClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to create a LOCAL cluster in Docker",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster local <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}

		SetRequiredFeatureFlags(cmd)
		cli.Client.Metadata.Provider = CLOUD_LOCAL
		SetDefaults(CLOUD_LOCAL, CLUSTER_TYPE_MANG)
		createManaged(cmd.Flags().Lookup("approve").Changed)
	},
}

var createClusterHACivo = &cobra.Command{
	Use:   "ha-civo",
	Short: "Use to create a HA CIVO k3s cluster",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster ha-civo <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		SetRequiredFeatureFlags(cmd)

		cli.Client.Metadata.Provider = CLOUD_CIVO
		SetDefaults(CLOUD_CIVO, CLUSTER_TYPE_HA)
		createHA(cmd.Flags().Lookup("approve").Changed)
	},
}

var createClusterHAAzure = &cobra.Command{
	Use:   "ha-azure",
	Short: "Use to create a HA k3s cluster in Azure",
	Long: `It is used to create cluster with the given name from user. For example:

	ksctl create-cluster ha-azure <arguments to civo cloud provider>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		SetRequiredFeatureFlags(cmd)
		cli.Client.Metadata.Provider = CLOUD_AZURE
		SetDefaults(CLOUD_AZURE, CLUSTER_TYPE_HA)
		createHA(cmd.Flags().Lookup("approve").Changed)
	},
}

func init() {
	rootCmd.AddCommand(createClusterCmd)

	createClusterCmd.AddCommand(createClusterAws)
	createClusterCmd.AddCommand(createClusterAzure)
	createClusterCmd.AddCommand(createClusterCivo)
	createClusterCmd.AddCommand(createClusterLocal)
	createClusterCmd.AddCommand(createClusterHACivo)
	createClusterCmd.AddCommand(createClusterHAAzure)

	createClusterAzure.MarkFlagRequired("name")
	createClusterCivo.MarkFlagRequired("name")
	createClusterCivo.MarkFlagRequired("region")
	createClusterLocal.MarkFlagRequired("name")
	createClusterHAAzure.MarkFlagRequired("name")
	createClusterHACivo.MarkFlagRequired("name")
}
