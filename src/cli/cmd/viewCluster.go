package cmd

/*
Kubesimplify (c)
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

import (
	"fmt"

	"github.com/spf13/cobra"
)

// viewClusterCmd represents the viewCluster command
var viewClusterCmd = &cobra.Command{
	Use:   "view-clusters",
	Short: "Use to view clusters",
	Long: `It is used to view clusters. For example:

kubesimpctl view-clusters `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubesimpctl view-clusters [CALLED]")
	},
}

func init() {
	rootCmd.AddCommand(viewClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// viewClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// viewClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
