package cmd

import "github.com/spf13/cobra"

var switchCluster = &cobra.Command{
	Use:   "switch-context",
	Short: "Use to switch between clusters",
	Long: `It is used to create cluster with the given name from user. For example:

kubesimpctl switch-context -p <civo,local>  -c <clustername> -r <region> <arguments to civo cloud provider>
`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(switchCluster)
}
