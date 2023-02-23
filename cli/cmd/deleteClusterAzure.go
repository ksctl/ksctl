package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/
import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/azure"
	"github.com/spf13/cobra"
)

var deleteClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a azure managed cluster",
	Long: `It is used to create cluster with the given name from user. For example:

ksctl create-cluster azure <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {

		payload := &azure.AzureProvider{
			ClusterName: azdcclusterName,
			HACluster:   false,
			Region:      azdcregion,
		}
		err := payload.DeleteCluster()
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mCREATED!\033[0m\n")
	},
}

var (
	azdcclusterName string
	azdcregion      string
)

func init() {
	deleteClusterCmd.AddCommand(deleteClusterAzure)
	deleteClusterAzure.Flags().StringVarP(&azdcclusterName, "name", "n", "", "Cluster name")
	deleteClusterAzure.Flags().StringVarP(&azdcregion, "region", "r", "eastus", "Region")
	deleteClusterAzure.MarkFlagRequired("name")
	deleteClusterAzure.MarkFlagRequired("region")
}
