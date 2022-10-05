/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	credentials "github.com/dipankardas011/Kubesimpctl/src/api/payload"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
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
			err := fmt.Errorf("Given Choice is Invalid!")
			fmt.Println(err.Error())
			return
		}

		for !isSuccess {
			acckey := ""
			secacckey := ""
			fmt.Println("Enter your ACCESS-KEY: ")
			_, err := fmt.Scanf("%s", &acckey)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("Enter your SECRET-ACCESS-KEY: ")
			_, err = fmt.Scanf("%s", &secacckey)
			if err != nil {
				panic(err.Error())
			}
			credentialStore := credentials.Credential{AccessKey: acckey, Secret: secacckey}
			// store them ~/.kube/kubesimpctl/cred/<provider>/credential
			fmt.Println("Entered Cred")
			fmt.Printf(`
Access-Key: %s
Secret-Access-Key: %s`, credentialStore.AccessKey, credentialStore.Secret)
			isSuccess = true
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
