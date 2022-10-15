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
	"strings"
)

// createClusterCmd represents the createCluster command
var createClusterCmd = &cobra.Command{
	Use:   "create-cluster",
	Short: "Use to create a cluster",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster <name-cluster> --provider or -p ["azure", "gcp", "aws", "local"]
CONSTRAINS: only single provider can be used at a time.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(cregion) == 0 && strings.Compare(cprovider, "local") != 0 {
			panic(fmt.Errorf("region needs to be specifyed when using cloud providers"))
		}
		switch cprovider {
		case "civo":
			clusterConfig := civoHandler.ClusterInfoInjecter(
				cclusterName,
				cregion,
				cspec.Disk,
				cspec.Nodes,
				"Nginx",  // TODO: Add Application addition option in CLI
				"cilium") // TODO: Add CNI plugin addition option in CLI
			err := civoHandler.CreateCluster(clusterConfig)
			if err != nil {
				fmt.Printf("\033[31;40m%v\033[0m\n", err)
				return
			}
			fmt.Printf("\033[32;40mCREATED!\033[0m\n")

		case "azure":
			fmt.Println(azHandler.AKSHandler())
		case "aws":
			fmt.Println(awsHandler.EKSHandler())
		case "local":
			fmt.Println(localHandler.DockerHandler())
		}
	},
}

var (
	cclusterName string
	cprovider    string
	cspec        payload.Machine
	cregion      string
)

func init() {
	rootCmd.AddCommand(createClusterCmd)
	createClusterCmd.Flags().StringVarP(&cclusterName, "name", "C", "demo", "Cluster name")
	createClusterCmd.Flags().StringVarP(&cspec.Cpu, "cpus", "c", "2", "CPU size")
	createClusterCmd.Flags().StringVarP(&cspec.Mem, "memory", "m", "4Gi", "Memory size")
	createClusterCmd.Flags().StringVarP(&cspec.Disk, "disks", "d", "500M", "Disk Size")
	createClusterCmd.Flags().StringVarP(&cregion, "region", "r", "", "Region based on different cloud providers")
	createClusterCmd.Flags().IntVarP(&cspec.Nodes, "nodes", "n", 1, "Number of Nodes")
	createClusterCmd.Flags().StringVarP(&cprovider, "provider", "p", "local", "Provider Name [aws, azure, civo, local]")
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
