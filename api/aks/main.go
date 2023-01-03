/*
Kubesimplify
@maintainer:
*/

package aks

import (
	"fmt"
	"os"

	util "github.com/kubesimplify/ksctl/api/utils"
)

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	_, err := os.ReadFile(util.GetPath(0, "aks"))
	if err != nil {
		return ""
	}
	return ""
}

func Credentials() bool {
	// _, err := os.OpenFile(util.GetPath(0, "azure"), os.O_WRONLY, 0640)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }

	skey := ""
	tid := ""
	pi := ""
	pk := ""
	func() {
		fmt.Println("Enter your SUBSCRIPTION ID: ")
		_, err := fmt.Scan(&skey)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("Enter your TENANT ID: ")
		_, err = fmt.Scan(&tid)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("Enter your SERVICE PRINCIPAL ID: ")
		_, err = fmt.Scan(&pi)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("Enter your : SERVICE PRINCIPAL KEY")
		_, err = fmt.Scan(&pk)
		if err != nil {
			panic(err.Error())
		}
	}()

	return true
	// _, err = file.Write([]byte(fmt.Sprintf(`Subscription-ID: %s
	// 	Tenant-ID: %s
	// 	Service-Principal-ID: %s
	// 	Service Principal-Key: %s`, skey, tid, pi, pk)))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	// return true
}
