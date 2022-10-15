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
	"github.com/spf13/cobra"
	"strings"
)

// deleteClusterCmd represents the deleteCluster command
var deleteClusterCmd = &cobra.Command{
	Use:   "delete-cluster",
	Short: "Use to delete a cluster",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl delete-cluster <name-cluster> `,
	Run: func(cmd *cobra.Command, args []string) {
		if len(dregion) == 0 && strings.Compare(dprovider, "local") != 0 {
			panic(fmt.Errorf("region needs to be specifyed when using cloud providers"))
		}
		switch dprovider {
		case "civo":
			err := civoHandler.DeleteCluster(dregion, dclusterName)
			if err != nil {
				fmt.Printf("\033[31;40m%v\033[0m\n", err)
				return
			}
			fmt.Printf("\033[32;40mDELETED!\033[0m\n")

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
	dclusterName string
	dprovider    string
	dregion      string
)

func init() {
	rootCmd.AddCommand(deleteClusterCmd)
	deleteClusterCmd.Flags().StringVarP(&dclusterName, "name", "C", "demo", "Cluster name")
	deleteClusterCmd.Flags().StringVarP(&dregion, "region", "r", "", "Region based on different cloud providers")
	deleteClusterCmd.Flags().StringVarP(&dprovider, "provider", "p", "local", "Provider Name [aws, azure, civo, local]")
	deleteClusterCmd.MarkFlagRequired("name")
	deleteClusterCmd.MarkFlagRequired("provider")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
