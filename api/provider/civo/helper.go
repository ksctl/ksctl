package civo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kubesimplify/ksctl/api/utils"
)

// fetchAPIKey returns the api_token from the cred/civo.json file store
func fetchAPIKey() string {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken
	}
	return ""
	// logger.Warn("environment vars not set: CIVO_TOKEN")

	// token, err := util.GetCred(logger, "civo")
	// if err != nil {
	// 	return ""
	// }
	// return token["token"]
}

func generatePath(flag int, path ...string) string {
	return utils.GetPath(flag, "civo", path...)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func validationOfArguments(name, region string) error {

	if !isValidRegionCIVO(region) {
		return fmt.Errorf("REGION")
	}

	if !utils.IsValidName(name) {
		return fmt.Errorf("NAME FORMAT")
	}

	return nil
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegionCIVO(reg string) bool {
	return strings.Compare(reg, "FRA1") == 0 ||
		strings.Compare(reg, "NYC1") == 0 ||
		strings.Compare(reg, "PHX1") == 0 ||
		strings.Compare(reg, "LON1") == 0
}
