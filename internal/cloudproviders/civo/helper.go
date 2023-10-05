package civo

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
)

// fetchAPIKey returns the api_token from the cred/civo.json file store
func fetchAPIKey(storage resources.StorageFactory) string {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken
	}
	storage.Logger().Warn("environment vars not set: `CIVO_TOKEN`")

	token, err := utils.GetCred(storage, CLOUD_CIVO)
	if err != nil {
		return ""
	}
	return token["token"]
}

func GetInputCredential(storage resources.StorageFactory) error {

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

	if err := utils.SaveCred(storage, Credential{token}, CLOUD_CIVO); err != nil {
		return err
	}
	return nil
}

func generatePath(flag KsctlUtilsConsts, clusterType KsctlClusterType, path ...string) string {
	return utils.GetPath(flag, CLOUD_CIVO, clusterType, path...)
}

func saveStateHelper(storage resources.StorageFactory, path string) error {
	rawState, err := convertStateToBytes(*civoCloudState)
	if err != nil {
		return err
	}
	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func loadStateHelper(storage resources.StorageFactory, path string) error {
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	return convertStateFromBytes(raw)
}

func saveKubeconfigHelper(storage resources.StorageFactory, path string, kubeconfig string) error {
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

// helper functions to get resources from civogo client
// seperation so that we can test logic by assert
func getValidK8sVersionClient(obj *CivoProvider) []string {
	vers, err := obj.client.ListAvailableKubernetesVersions()
	if err != nil {
		return nil
	}
	var val []string
	for _, ver := range vers {
		if ver.ClusterType == string(K8S_K3S) {
			val = append(val, ver.Label)
		}
	}
	return val
}

func getValidRegionsClient(obj *CivoProvider) []string {
	regions, err := obj.client.ListRegions()
	if err != nil {
		return nil
	}
	var val []string
	for _, region := range regions {
		val = append(val, region.Code)
	}
	return val
}

func getValidVMSizesClient(obj *CivoProvider) []string {
	nodeSizes, err := obj.client.ListInstanceSizes()
	if err != nil {
		return nil
	}
	var val []string
	for _, region := range nodeSizes {
		val = append(val, region.Name)
	}
	return val
}

func validationOfArguments(obj *CivoProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	if err := utils.IsValidName(obj.clusterName); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *CivoProvider, ver string) error {
	var valver []string = getValidK8sVersionClient(obj)
	for _, vver := range valver {
		if vver == ver {
			return nil
		}
	}
	return fmt.Errorf("Invalid k8s version\nValid options: %v\n", valver)
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegion(obj *CivoProvider, reg string) error {
	var validFromClient []string = getValidRegionsClient(obj)
	for _, region := range validFromClient {
		if region == reg {
			return nil
		}
	}
	return fmt.Errorf("INVALID REGION\nValid options: %v\n", validFromClient)
}

func isValidVMSize(obj *CivoProvider, size string) error {
	var validFromClient []string = getValidVMSizesClient(obj)
	for _, nodeSize := range validFromClient {
		if size == nodeSize {
			return nil
		}
	}
	return fmt.Errorf("INVALID VM SIZE\nValid options: %v\n", validFromClient)
}

func printKubeconfig(storage resources.StorageFactory, operation KsctlOperation) {
	env := ""
	storage.Logger().Note("KUBECONFIG env var")
	path := generatePath(CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
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
