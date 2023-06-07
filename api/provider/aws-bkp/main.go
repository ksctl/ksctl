/*
Kubesimplify
@maintainer:
*/

package eks

import (
	"fmt"
	"os"

	util "github.com/kubesimplify/ksctl/api/provider/utils"
)

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	_, err := os.ReadFile(util.GetPath(0, "aws"))
	if err != nil {
		return ""
	}

	return ""
}

func Credentials() bool {
	// _, err := os.OpenFile(util.GetPath(0, "aws"), os.O_WRONLY, 0640)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	acckey := ""
	secacckey := ""
	func() {
		fmt.Println("Enter your ACCESS-KEY: ")
		_, err := fmt.Scan(&acckey)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Enter your SECRET-ACCESS-KEY: ")
		_, err = fmt.Scan(&secacckey)
		if err != nil {
			panic(err.Error())
		}
	}()

	// _, err = file.Write([]byte(fmt.Sprintf(`Access-Key: %s
	// 	Secret-Access-Key: %s`, acckey, secacckey)))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	return true
}
