package civo

import (
	"encoding/json"
	"fmt"
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
	"runtime"
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
	client, err := civogo.NewClient(token, "LON1")
	if err != nil {
		return err
	}
	id := client.GetAccountID()

	if len(id) == 0 {
		return fmt.Errorf("Invalid user")
	}
	fmt.Println(id)

	if err := utils.SaveCred(storage, Credential{token}, "civo"); err != nil {
		return err
	}
	return nil
}

func generatePath(flag int, path ...string) string {
	return utils.GetPath(flag, "civo", path...)
}

func saveStateHelper(storage resources.StateManagementInfrastructure, path string) error {
	rawState, err := convertStateToBytes(*civoCloudState)
	if err != nil {
		return err
	}
	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func loadStateHelper(storage resources.StateManagementInfrastructure, path string) error {
	fmt.Println(path)
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	return convertStateFromBytes(raw)
}

func saveKubeconfigHelper(storage resources.StateManagementInfrastructure, path string, kubeconfig string) error {
	rawState := []byte(kubeconfig)

	return storage.Path(path).Permission(FILE_PERM_CLUSTER_KUBECONFIG).Save(rawState)
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
		if region.Code == reg {
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

func printKubeconfig(storage resources.StateManagementInfrastructure, operation string) {
	env := ""
	storage.Logger().Note("KUBECONFIG env var")
	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
	switch runtime.GOOS {
	case "windows":
		switch operation {
		case "create":
			env = fmt.Sprintf("$Env:KUBECONFIG=\"%s\"\n", path)
		case "delete":
			env = fmt.Sprintf("$Env:KUBECONFIG=\"\"\n")
		}
	case "linux", "macos":
		switch operation {
		case "create":
			env = fmt.Sprintf("export KUBECONFIG=\"%s\"\n", path)
		case "delete":
			env = "unset KUBECONFIG"
		}
	}
	storage.Logger().Note(env)
}
