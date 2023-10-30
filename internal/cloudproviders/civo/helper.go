package civo

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// fetchAPIKey returns the api_token from the cred/civo.json file store
func fetchAPIKey(storage resources.StorageFactory) string {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken
	}
	log.Warn("environment vars not set: `CIVO_TOKEN`")

	token, err := utils.GetCred(storage, log, CloudCivo)
	if err != nil {
		return ""
	}
	return token["token"]
}

func GetInputCredential(storage resources.StorageFactory) error {

	log.Print("Enter CIVO TOKEN")
	token, err := utils.UserInputCredentials(log)
	if err != nil {
		return err
	}
	client, err := civogo.NewClient(token, "LON1")
	if err != nil {
		return err
	}
	id := client.GetAccountID()

	if len(id) == 0 {
		return log.NewError("Invalid user")
	}
	log.Print(id)

	if err := utils.SaveCred(storage, log, Credential{token}, CloudCivo); err != nil {
		return err
	}
	return nil
}

func generatePath(flag KsctlUtilsConsts, clusterType KsctlClusterType, path ...string) string {
	p := utils.GetPath(flag, CloudCivo, clusterType, path...)
	log.Debug("Printing", "path", p)
	return p
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
		return log.NewError(err.Error())
	}

	return convertStateFromBytes(raw)
}

func saveKubeconfigHelper(storage resources.StorageFactory, path string, kubeconfig string) error {
	rawState := []byte(kubeconfig)
	log.Debug("Printing", "kubeconfig", kubeconfig, "path", path)

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
		log.Error("unable to get available k8s versions", "err", err)
		return nil
	}
	log.Debug("Printing", "ListAvailableKubernetesVersions", vers)
	var val []string
	for _, ver := range vers {
		if ver.ClusterType == string(K8sK3s) {
			val = append(val, ver.Label)
		}
	}
	return val
}

func getValidRegionsClient(obj *CivoProvider) []string {
	regions, err := obj.client.ListRegions()
	if err != nil {
		log.Error("unable to get available regions", "err", err)
		return nil
	}
	log.Debug("Printing", "ListRegions", regions)
	var val []string
	for _, region := range regions {
		val = append(val, region.Code)
	}
	return val
}

func getValidVMSizesClient(obj *CivoProvider) []string {
	nodeSizes, err := obj.client.ListInstanceSizes()
	if err != nil {
		log.Error("unable to fetch list of valid instance sizes", "err", err)
		return nil
	}
	log.Debug("Printing", "ListInstanceSizes", nodeSizes)
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
	return log.NewError("Invalid k8s version\nValid options: %v\n", valver)
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegion(obj *CivoProvider, reg string) error {
	var validFromClient []string = getValidRegionsClient(obj)
	for _, region := range validFromClient {
		if region == reg {
			return nil
		}
	}
	return log.NewError("INVALID REGION\nValid options: %v\n", validFromClient)
}

func isValidVMSize(obj *CivoProvider, size string) error {
	var validFromClient []string = getValidVMSizesClient(obj)
	for _, nodeSize := range validFromClient {
		if size == nodeSize {
			return nil
		}
	}
	return log.NewError("INVALID VM SIZE\nValid options: %v\n", validFromClient)
}

func printKubeconfig(storage resources.StorageFactory, operation KsctlOperation) {
	env := ""
	log.Note("KUBECONFIG env var")
	path := generatePath(UtilClusterPath, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
	log.Debug("Printing", "path", path)
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
	log.Note(env)
}
