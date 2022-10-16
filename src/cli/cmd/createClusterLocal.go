package cmd

import "github.com/spf13/cobra"

var createClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to create a LOCAL cluster in Docker",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl create-cluster local <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	createClusterCmd.AddCommand(createClusterLocal)
}
