package cmd

import "github.com/spf13/cobra"

var createClusterAws = &cobra.Command{
	Use:   "aws",
	Short: "Use to create a EKS cluster in AWS",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster aws <arguments to civo cloud provider>
CONSTRAINS: only single provider can be used at a time.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	createClusterCmd.AddCommand(createClusterAws)
}
