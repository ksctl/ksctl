package cmd

import "github.com/spf13/cobra"

var deleteClusterLocal = &cobra.Command{
	Use:   "local",
	Short: "Use to delete a local cluster",
	Long: `It is used to delete cluster of given provider. For example:

kubesimpctl delete-cluster local <arguments to local/Docker provider>
`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	deleteClusterCmd.AddCommand(deleteClusterLocal)
}
