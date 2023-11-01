package azure

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func GetInputCredential(storage resources.StorageFactory, meta resources.Metadata) error {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(CloudAws))

	log.Print("Enter your SUBSCRIPTION ID")
	skey, err := utils.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your TENANT ID")
	tid, err := utils.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your CLIENT ID")
	cid, err := utils.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your CLIENT SECRET")
	cs, err := utils.UserInputCredentials(log)
	if err != nil {
		return err
	}

	apiStore := Credential{
		SubscriptionID: skey,
		TenantID:       tid,
		ClientID:       cid,
		ClientSecret:   cs,
	}

	// FIXME: add ping pong for validation of credentials
	//if err = os.Setenv("AZURE_SUBSCRIPTION_ID", skey); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_TENANT_ID", tid); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_CLIENT_ID", cid); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_CLIENT_SECRET", cs); err != nil {
	//	return err
	//}
	// ADD SOME PING method to validate credentials

	if err := utils.SaveCred(storage, log, apiStore, CloudAzure); err != nil {
		return err
	}

	return nil
}

func generatePath(flag KsctlUtilsConsts, clusterType KsctlClusterType, path ...string) string {
	return utils.GetPath(flag, CloudAzure, clusterType, path...)
}

func saveStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(UtilClusterPath, CloudAzure, clusterType, clusterDirName, STATE_FILE_NAME)
	rawState, err := convertStateToBytes(*azureCloudState)
	if err != nil {
		return err
	}
	log.Debug("Printing", "rawState", string(rawState), "path", path)

	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func loadStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(UtilClusterPath, CloudAzure, clusterType, clusterDirName, STATE_FILE_NAME)
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	log.Debug("Printing", "rawState", string(raw), "path", path)
	return convertStateFromBytes(raw)
}

func saveKubeconfigHelper(storage resources.StorageFactory, kubeconfig string) error {
	rawState := []byte(kubeconfig)
	path := utils.GetPath(UtilClusterPath, CloudAzure, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)

	log.Debug("Printing", "path", path, "kubeconfig", kubeconfig)

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
	azureCloudState = data
	return nil
}

func printKubeconfig(storage resources.StorageFactory, operation KsctlOperation) {
	key := ""
	value := ""
	box := ""
	switch runtime.GOOS {
	case "windows":
		key = "$Env:KUBECONFIG"

		switch operation {
		case OperationStateCreate:
			value = generatePath(UtilClusterPath, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)

		case OperationStateDelete:
			value = ""
		}
		box = key + "=" + fmt.Sprintf("\"%s\"", value)
		log.Note("KUBECONFIG env var", key, value)

	case "linux", "macos":

		switch operation {
		case OperationStateCreate:
			key = "export KUBECONFIG"
			value = generatePath(UtilClusterPath, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
			box = key + "=" + fmt.Sprintf("\"%s\"", value)
			log.Note("KUBECONFIG env var", key, value)

		case OperationStateDelete:
			key = "unset KUBECONFIG"
			box = key
			log.Note(key)
		}
	}

	log.Box("KUBECONFIG env var", box)
}

func validationOfArguments(obj *AzureProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	if err := utils.IsValidName(obj.clusterName); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *AzureProvider, ver string) error {
	res, err := obj.client.ListKubernetesVersions()
	if err != nil {
		return log.NewError("failed to finish the request: %v", err)
	}

	log.Debug("Printing", "ListKubernetesVersions", res)

	var vers []string
	for _, version := range res.Values {
		vers = append(vers, *version.Version)
	}
	for _, valver := range vers {
		if valver == ver {
			return nil
		}
	}
	return log.NewError("Invalid k8s version\nValid options: %v\n", vers)
}

func isValidRegion(obj *AzureProvider, reg string) error {
	validReg, err := obj.client.ListLocations()
	if err != nil {
		return err
	}
	log.Debug("Printing", "ListLocation", validReg)

	for _, valid := range validReg {
		if valid == reg {
			return nil
		}
	}
	return log.NewError("INVALID REGION\nValid options: %v\n", validReg)
}

func isValidVMSize(obj *AzureProvider, size string) error {

	validSize, err := obj.client.ListVMTypes()
	if err != nil {
		return err
	}
	log.Debug("Printing", "ListVMType", validSize)

	for _, valid := range validSize {
		if valid == size {
			return nil
		}
	}

	return log.NewError("INVALID VM SIZE\nValid options %v\n", validSize)
}
