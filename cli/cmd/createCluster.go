package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/provider/local"
	log "github.com/kubesimplify/ksctl/api/provider/logger"
	"github.com/kubesimplify/ksctl/api/provider/utils"
	util "github.com/kubesimplify/ksctl/api/provider/utils"
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
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := &azure.AzureProvider{
			ClusterName: clusterName,
			HACluster:   false,
			Region:      region,
			Spec: util.Machine{
				ManagedNodes: noManagedNodes,
				Disk:         nodeSize,
			},
		}
		err := payload.CreateCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER")
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
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := civo.CivoProvider{
			ClusterName: clusterName,
			Region:      region,
			Application: apps,
			CNIPlugin:   cni,
			HACluster:   false,
			Spec: utils.Machine{
				Disk:         nodeSize,
				ManagedNodes: noManagedNodes,
			},
		}
		err := payload.CreateCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER")
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
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		cargo := local.ClusterInfoInjecter(clusterName, noManagedNodes)
		logger.Info("Building cluster", "")
		if err := local.CreateCluster(logger, cargo); err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER")
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
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}
		payload := civo.CivoProvider{
			ClusterName: clusterName,
			Region:      region,
			HACluster:   true,
			Spec: utils.Machine{
				Disk:                nodeSize,
				HAControlPlaneNodes: noControlPlaneNodes,
				HAWorkerNodes:       noWorkerNodes,
			},
		}
		err := payload.CreateCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER")
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
		logger := log.Logger{Verbose: true}
		if !isSet {
			logger.Verbose = false
		}

		payload := &azure.AzureProvider{
			ClusterName: clusterName,
			HACluster:   true,
			Region:      region,
			Spec: util.Machine{
				Disk:                nodeSize,
				HAControlPlaneNodes: noControlPlaneNodes,
				HAWorkerNodes:       noWorkerNodes,
			},
		}
		err := payload.CreateCluster(logger)
		if err != nil {
			logger.Err(err.Error())
			return
		}
		logger.Info("CREATED CLUSTER")
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

	createClusterAws.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterAzure.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterCivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterLocal.Flags().BoolP("verbose", "v", true, "Verbose output")
	createClusterHACivo.Flags().BoolP("verbose", "v", true, "for verbose output")
	createClusterHAAzure.Flags().BoolP("verbose", "v", true, "for verbose output")

	createClusterAzure.MarkFlagRequired("name")
	createClusterCivo.MarkFlagRequired("name")
	createClusterCivo.MarkFlagRequired("region")
	createClusterLocal.MarkFlagRequired("name")
	createClusterHAAzure.MarkFlagRequired("name")
	createClusterHACivo.MarkFlagRequired("name")

	createClusterAzure.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	createClusterAzure.Flags().StringVarP(&nodeSize, "node-size", "s", "Standard_DS2_v2", "Node size")
	createClusterAzure.Flags().StringVarP(&region, "region", "r", "eastus", "Region")
	createClusterAzure.Flags().IntVarP(&noManagedNodes, "nodes", "N", 1, "Number of Nodes")

	createClusterCivo.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	createClusterCivo.Flags().StringVarP(&nodeSize, "nodeSize", "s", "g4s.kube.xsmall", "Node size")
	createClusterCivo.Flags().StringVarP(&region, "region", "r", "", "Region")
	createClusterCivo.Flags().StringVarP(&apps, "apps", "a", "", "PreInstalled Apps with comma seperated string")
	createClusterCivo.Flags().StringVarP(&cni, "cni", "c", "", "CNI Plugin to be installed")
	createClusterCivo.Flags().IntVarP(&noManagedNodes, "nodes", "N", 1, "Number of Nodes")

	createClusterLocal.Flags().StringVarP(&clusterName, "name", "n", "demo", "Cluster name")
	createClusterLocal.Flags().IntVarP(&noManagedNodes, "nodes", "N", 1, "Number of Nodes")

	createClusterHACivo.Flags().StringVarP(&nodeSize, "nodeSize", "s", "g3.small", "Node size")
	createClusterHACivo.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	createClusterHACivo.Flags().StringVarP(&region, "region", "r", "LON1", "Region")
	createClusterHACivo.Flags().IntVarP(&noControlPlaneNodes, "control-nodes", "c", 3, "no of control plane nodes")
	createClusterHACivo.Flags().IntVarP(&noWorkerNodes, "worker-nodes", "w", 1, "no of worker nodes")

	createClusterHAAzure.Flags().StringVarP(&clusterName, "name", "n", "", "Cluster name")
	createClusterHAAzure.Flags().StringVarP(&nodeSize, "node-size", "s", "Standard_F2s", "Node size")
	createClusterHAAzure.Flags().StringVarP(&region, "region", "r", "eastus", "Region")
	createClusterHAAzure.Flags().IntVarP(&noWorkerNodes, "worker-nodes", "w", 1, "Number of worker Nodes")
	createClusterHAAzure.Flags().IntVarP(&noControlPlaneNodes, "control-nodes", "c", 3, "Number of control Nodes")
}
