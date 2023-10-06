// maintainer: 	Dipankar Das <dipankardas0115@gmail.com>

package cmd

import (
	"fmt"

	control_pkg "github.com/kubesimplify/ksctl/pkg/controllers"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var credCmd = &cobra.Command{
	Use:   "cred",
	Short: "Login to your Cloud-provider Credentials",
	Long: `login to your cloud provider credentials
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSet := cmd.Flags().Lookup("verbose").Changed
		if _, err := control_pkg.InitializeStorageFactory(&cli.Client, isSet); err != nil {
			panic(err)
		}
		SetRequiredFeatureFlags(cmd)

		fmt.Println(`
1> AWS (EKS)
2> Azure (AKS)
3> Civo (K3s)
`)

		choice := 0

		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			panic(err.Error())
		}
		if provider, ok := cloud[choice]; ok {
			cli.Client.Metadata.Provider = KsctlCloud(provider)
		} else {
			cli.Client.Storage.Logger().Err("invalid provider")
		}

		stat, err := controller.Credentials(&cli.Client)
		if err != nil {
			cli.Client.Storage.Logger().Err(err.Error())
			return
		}
		cli.Client.Storage.Logger().Success(stat)
	},
}

func init() {
	rootCmd.AddCommand(credCmd)
	credCmd.Flags().BoolP("verbose", "v", true, "for verbose output")

}
