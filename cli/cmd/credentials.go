package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/aks"
	"github.com/kubesimplify/ksctl/api/civo"
	"github.com/kubesimplify/ksctl/api/eks"
	"github.com/spf13/cobra"
)

const (
	AWS   = 1
	AZURE = 2
	CIVO  = 3
)

// initCmd represents the init command
var credCmd = &cobra.Command{
	Use:   "cred",
	Short: "Login with your Cloud-provider Credentials",
	Long: `login with your cloud provider credentials
`,
	Run: func(cmd *cobra.Command, args []string) {
		isSuccess := false
		fmt.Println(`
1> AWS (EKS)
2> Azure (AKS)
3> Civo (K3s)`)
		choice := 0
		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			panic(err.Error())
		}
		switch choice {
		case 1, 2, 3:
			fmt.Println("Enter your credentials")
		default:
			err := fmt.Errorf("given Choice is Invalid")
			fmt.Println(err.Error())
			return
		}

		isSuccess = storeCredentials(choice)
		if !isSuccess {
			fmt.Println("Login Failed")
		}
		fmt.Println("Login Success")

	},
}

func storeCredentials(provider int) bool {

	//TODO: Verify the Credentials
	switch provider {
	case AWS:
		return eks.Credentials()
	case CIVO:
		return civo.Credentials()
	case AZURE:
		return aks.Credentials()
	default:
		return false
	}
}

func init() {
	rootCmd.AddCommand(credCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
