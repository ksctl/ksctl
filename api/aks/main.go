/*
Kubesimplify
@maintainer:
*/

package aks

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	util "github.com/kubesimplify/ksctl/api/utils"
)

func getKubeconfig(params ...string) string {
	var ret strings.Builder

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\.ksctl\\config\\azure", util.GetUserName()))
		for _, item := range params {
			ret.WriteString("\\" + item)
		}
	} else {
		ret.WriteString(fmt.Sprintf("%s/.ksctl/config/azure", util.GetUserName()))
		for _, item := range params {
			ret.WriteString("/" + item)
		}
	}
	return ret.String()
}

func getCredentials() string {

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\azure", util.GetUserName())
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/azure", util.GetUserName())
	}
}

// GetPath use this in every function and differentiate the logic by using if-else
// flag is used to indicate 1 -> KUBECONFIG, 0 -> CREDENTIALS
func GetPath(flag int8, params ...string) string {
	switch flag {
	case 1:
		return getKubeconfig(params...)
	case 0:
		return getCredentials()
	default:
		return ""
	}
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(GetPath(0))
	if err != nil {
		return ""
	}
	if len(file) == 0 {
		return ""
	}

	return strings.Split(string(file), " ")[1]
}

func Credentials() bool {
	file, err := os.OpenFile(GetPath(0), os.O_WRONLY, 0640)
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
		_, err = fmt.Scan(&skey)
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

	_, err = file.Write([]byte(fmt.Sprintf(`Subscription-ID: %s
		Tenant-ID: %s
		Service-Principal-ID: %s
		Service Principal-Key: %s`, skey, tid, pi, pk)))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
