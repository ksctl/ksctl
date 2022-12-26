/*
Kubesimplify
@maintainer:
*/

package eks

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	util "github.com/kubesimplify/ksctl/api/utils"
)

func getKubeconfig(params ...string) string {
	var ret string

	if runtime.GOOS == "windows" {
		ret = fmt.Sprintf("%s\\.ksctl\\config\\aws", util.GetUserName())
		for _, item := range params {
			ret += "\\" + item
		}
	} else {
		ret = fmt.Sprintf("%s/.ksctl/config/aws", util.GetUserName())
		for _, item := range params {
			ret += "/" + item
		}
	}
	return ret
}

func getCredentials() string {

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\aws", util.GetUserName())
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/aws", util.GetUserName())
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
	acckey := ""
	secacckey := ""
	func() {
		fmt.Println("Enter your ACCESS-KEY: ")
		_, err = fmt.Scan(&acckey)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Enter your SECRET-ACCESS-KEY: ")
		_, err = fmt.Scan(&secacckey)
		if err != nil {
			panic(err.Error())
		}
	}()

	_, err = file.Write([]byte(fmt.Sprintf(`Access-Key: %s
		Secret-Access-Key: %s`, acckey, secacckey)))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
