package civo

import (
	"encoding/json"
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
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

func saveStateHelper(state resources.StateManagementInfrastructure, path string) error {
	rawState, err := convertStateToBytes(*civoCloudState)
	if err != nil {
		return err
	}
	return state.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func loadStateHelper(state resources.StateManagementInfrastructure, path string) error {
	fmt.Println(path)
	raw, err := state.Path(path).Load()
	if err != nil {
		return err
	}

	return convertStateFromBytes(raw)
}

func saveKubeconfigHelper(state resources.StateManagementInfrastructure, path string, kubeconfig string) error {
	rawState := []byte(kubeconfig)

	return state.Path(path).Permission(FILE_PERM_CLUSTER_KUBECONFIG).Save(rawState)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func convertStateFromBytes(raw []byte) error {
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	civoCloudState = data
	return nil
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
