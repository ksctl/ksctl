package cmd

/*
Kubesimplify (c)
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

const (
	AWS   = 1
	AZURE = 2
	CIVO  = 3
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
			err := fmt.Errorf("given Choice is Invalid")
			fmt.Println(err.Error())
			return
		}

		for !isSuccess {
			isSuccess = storeCredentials(choice)
		}
	},
}

func storeCredentials(provider int) bool {

	usrCmd := exec.Command("whoami")

	output, err := usrCmd.Output()
	if err != nil {
		return false
	}
	userName := strings.Trim(string(output), "\n")

	switch provider {
	case AWS:
		file, err := os.OpenFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/cred/aws", userName), os.O_WRONLY, 0640)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		acckey := ""
		secacckey := ""
		func() {
			fmt.Println("Enter your ACCESS-KEY: ")
			_, err = fmt.Scanf("%s", &acckey)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("Enter your SECRET-ACCESS-KEY: ")
			_, err = fmt.Scanf("%s", &secacckey)
			if err != nil {
				panic(err.Error())
			}
		}()

		_, err = file.Write([]byte(fmt.Sprintf(`Access-Key: %s
Secret-Access-Key: %s\n`, acckey, secacckey)))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	case CIVO:
		file, err := os.OpenFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/cred/civo", userName), os.O_WRONLY, 0640)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		apikey := ""
		func() {
			fmt.Println("Enter your API-TOKEN-KEY: ")
			_, err = fmt.Scanf("%s", &apikey)
			if err != nil {
				panic(err.Error())
			}
		}()

		_, err = file.Write([]byte(fmt.Sprintf("API-TOKEN-Key: %s\n", apikey)))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	case AZURE:
		file, err := os.OpenFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/cred/azure", userName), os.O_WRONLY, 0640)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		skey := ""
		tid := ""
		pi := ""
		pk := ""
		func() {
			fmt.Println("Enter your SUBSCRIPTION ID: ")
			_, err = fmt.Scanf("%s", &skey)
			if err != nil {
				panic(err.Error())
			}

			fmt.Println("Enter your TENANT ID: ")
			_, err = fmt.Scanf("%s", &tid)
			if err != nil {
				panic(err.Error())
			}

			fmt.Println("Enter your SERVICE PRINCIPAL ID: ")
			_, err = fmt.Scanf("%s", &pi)
			if err != nil {
				panic(err.Error())
			}

			fmt.Println("Enter your : SERVICE PRINCIPAL KEY")
			_, err = fmt.Scanf("%s", &pk)
			if err != nil {
				panic(err.Error())
			}
		}()

		_, err = file.Write([]byte(fmt.Sprintf(`Subscription-ID: %s
Tenant-ID: %s
Service-Principal-ID: %s
Service Principal-Key: %s\n`, skey, tid, pi, pk)))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}

	//TODO: Verify the Credentials

	fmt.Println("Login Success")
	return true
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
