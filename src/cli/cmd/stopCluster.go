/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopClusterCmd represents the stopCluster command
var stopClusterCmd = &cobra.Command{
	Use:   "stop-cluster",
	Short: "Use to stop a cluster",
	Long: `It is used to stop cluster with the given name from user. For example:

kubesimpctl stop-cluster <name-cluster>
CONSTRAINS: one cluster are platform dependent some don't have start and stop features Ex. EKS`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubesimpctl stop-cluster [CALLED]")
	},
}

func init() {
	rootCmd.AddCommand(stopClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
