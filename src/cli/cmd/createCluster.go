package cmd

/*
Kubesimplify (c)
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

import (
	"fmt"
	azHandler "github.com/kubesimplify/Kubesimpctl/src/api/aks"
	civoHandler "github.com/kubesimplify/Kubesimpctl/src/api/civo"
	awsHandler "github.com/kubesimplify/Kubesimpctl/src/api/eks"
	localHandler "github.com/kubesimplify/Kubesimpctl/src/api/local"
	payload "github.com/kubesimplify/Kubesimpctl/src/api/payload"
	"github.com/spf13/cobra"
)

// createClusterCmd represents the createCluster command
var createClusterCmd = &cobra.Command{
	Use:   "create-cluster",
	Short: "Use to create a cluster",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster <name-cluster> --provider or -p ["azure", "gcp", "aws", "local"]
CONSTRAINS: only single provider can be used at a time.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch provider {
		case "civo":
			fmt.Println(civoHandler.K3sHandler())
		case "azure":
			fmt.Println(azHandler.AKSHandler())
		case "aws":
			fmt.Println(awsHandler.EKSHandler())
		case "local":
			fmt.Println(localHandler.DockerHandler())
		}
		fmt.Printf(`
Name: %s
Provider: %s
  cpu: %s
  mem: %s
  disk: %s
  nodes: %v
`, clusterName, provider, spec.Cpu, spec.Mem, spec.Disk, spec.Nodes)
	},
}

var (
	clusterName string
	provider    string
	spec        payload.Machine
)

func init() {
	rootCmd.AddCommand(createClusterCmd)
	createClusterCmd.Flags().StringVarP(&clusterName, "name", "C", "demo", "Cluster name")
	createClusterCmd.Flags().StringVarP(&spec.Cpu, "cpus", "c", "2", "CPU size")
	createClusterCmd.Flags().StringVarP(&spec.Mem, "memory", "m", "4Gi", "Memory size")
	createClusterCmd.Flags().StringVarP(&spec.Disk, "disks", "d", "500M", "Disk Size")
	createClusterCmd.Flags().Uint8VarP(&spec.Nodes, "nodes", "n", 1, "Number of Nodes")
	createClusterCmd.Flags().StringVarP(&provider, "provider", "p", "local", "Provider Name [aws, azure, civo, local]")
	createClusterCmd.MarkFlagRequired("name")
	createClusterCmd.MarkFlagRequired("nodes")
	createClusterCmd.MarkFlagRequired("provider")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
