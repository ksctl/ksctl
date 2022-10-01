/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteClusterCmd represents the deleteCluster command
var deleteClusterCmd = &cobra.Command{
	Use:   "delete-cluster",
	Short: "Use to delete a cluster",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl delete-cluster <name-cluster> `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubesimpctl delete-cluster [CALLED]")
	},
}

func init() {
	rootCmd.AddCommand(deleteClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
