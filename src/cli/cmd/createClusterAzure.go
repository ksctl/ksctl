package cmd

import "github.com/spf13/cobra"

var createClusterAzure = &cobra.Command{
	Use:   "azure",
	Short: "Use to create a AKS cluster in Azure",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster azure <arguments to civo cloud provider>
CONSTRAINS: only single provider can be used at a time.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	createClusterCmd.AddCommand(createClusterAzure)
}
