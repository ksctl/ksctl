/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startClusterCmd represents the startCluster command
var startClusterCmd = &cobra.Command{
	Use:   "start-cluster",
	Short: "Use to start a cluster",
	Long: `It is used to start cluster with the given name from user. For example:

kubesimpctl start-cluster <name-cluster>
CONSTRAINS: one cluster are platform dependent some don't have start and stop features Ex. EKS`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubesimpctl start-cluster [CALLED]")
	},
}

func init() {
	rootCmd.AddCommand(startClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
