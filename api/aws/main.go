/*
Kubesimplify
@maintainer:
*/

package eks

import (
	"os"

	log "github.com/kubesimplify/ksctl/api/logger"
	util "github.com/kubesimplify/ksctl/api/utils"
)

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	_, err := os.ReadFile(util.GetPath(0, "aws"))
	if err != nil {
		return ""
	}

	return ""
}

func Credentials(logger log.Logger) bool {

	logger.Print("Enter your ACCESS-KEY: ")
	accesKey, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	logger.Print("Enter your SECRET-KEY: ")
	secret, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	apiStore := util.AwsCredential{
		AccesskeyID: accesKey,
		Secret:      secret,
	}

	err = util.SaveCred(logger, apiStore, "aws")
	if err != nil {
		logger.Err(err.Error())
		return false
	}
	return true

}
