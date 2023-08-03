package civo

import (
	"encoding/json"
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
)

// fetchAPIKey returns the api_token from the cred/civo.json file store
func fetchAPIKey(storage resources.StateManagementInfrastructure) string {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken
	}
	storage.Logger().Warn("environment vars not set: `CIVO_TOKEN`")

	token, err := utils.GetCred(storage, "civo")
	if err != nil {
		return ""
	}
	return token["token"]
}

func GetInputCredential(storage resources.StateManagementInfrastructure) error {

	storage.Logger().Print("Enter CIVO TOKEN")
	token, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	if err := utils.SaveCred(storage, Credential{token}, "civo"); err != nil {
		return err
	}
	return nil
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

	if err := isValidRegion(region); err != nil {
		return err
	}

	if err := utils.IsValidName(name); err != nil {
		return err
	}

	return nil
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegion(reg string) error {
	regions, err := civoClient.ListRegions()
	if err != nil {
		return err
	}
	for _, region := range regions {
		if region.Name == reg {
			return nil
		}
	}
	return fmt.Errorf("INVALID REGION")
}

func isValidVMSize(size string) error {
	nodeSizes, err := civoClient.ListInstanceSizes()
	if err != nil {
		return err
	}
	for _, nodeSize := range nodeSizes {
		if size == nodeSize.Name {
			return nil
		}
	}
	return fmt.Errorf("INVALID VM SIZE")
}
