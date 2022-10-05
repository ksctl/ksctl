/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	awsHandler "github.com/dipankardas011/Kubesimpctl/src/api/aks"
	"github.com/spf13/cobra"
)

//go get github.com/dipankardas011/Kubesimpctl/src/api@HEAD
// createClusterCmd represents the createCluster command
var createClusterCmd = &cobra.Command{
	Use:   "create-cluster",
	Short: "Use to create a cluster",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster <name-cluster> --provider or -p ["azure", "gcp", "aws", "local"]
CONSTRAINS: only single provider can be used at a time.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubesimpctl create-cluster [CALLED]")
		fmt.Println(awsHandler.AKSHandler())
	},
}

func init() {
	rootCmd.AddCommand(createClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
