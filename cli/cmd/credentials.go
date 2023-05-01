package cmd

/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

import (
	"fmt"

	log "github.com/kubesimplify/ksctl/api/logger"

	eks "github.com/kubesimplify/ksctl/api/aws"
	aks "github.com/kubesimplify/ksctl/api/azure"
	"github.com/kubesimplify/ksctl/api/civo"
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

		isSuccess = storeCredentials(cmd, choice)
		if !isSuccess {
			fmt.Println("Login Failed")
		}
		fmt.Println("Login Success")

	},
}

func storeCredentials(cmd *cobra.Command, provider int) bool {
	isSet := cmd.Flags().Lookup("verbose").Changed
	logger := log.Logger{Verbose: true}
	if !isSet {
		logger.Verbose = false
	}

	//TODO: Verify the Credentials
	switch provider {
	case AWS:
		return eks.Credentials(logger)
	case CIVO:
		return civo.Credentials(logger)
	case AZURE:
		return aks.Credentials(logger)
	default:
		return false
	}
}

func init() {
	rootCmd.AddCommand(credCmd)
	credCmd.Flags().BoolP("verbose", "v", true, "for verbose output")

}
