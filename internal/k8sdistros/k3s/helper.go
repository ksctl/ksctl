package k3s

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func saveStateHelper(storage resources.StorageFactory, path string) error {
	rawState, err := convertStateToBytes(*k8sState)
	if err != nil {
		return err
	}
	log.Debug("Printing", "k3sState", string(rawState))
	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func loadStateHelper(storage resources.StorageFactory, path string) error {
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	log.Debug("Printing", "rawState", string(raw))

	return convertBytesToState(raw)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func convertBytesToState(raw []byte) error {
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	k8sState = data

	log.Debug("Printing", "k3sState", k8sState)
	return nil
}
func saveKubeconfigHelper(storage resources.StorageFactory, path string, kubeconfig string) error {
	rawKubeconfig := []byte(kubeconfig)

	log.Debug("Printing", "kubeconfig", kubeconfig)
	return storage.Path(path).Permission(FILE_PERM_CLUSTER_KUBECONFIG).Save(rawKubeconfig)
}
func printKubeconfig(storage resources.StorageFactory, operation consts.KsctlOperation) {
	key := ""
	value := ""
	box := ""
	switch runtime.GOOS {
	case "windows":
		key = "$Env:KUBECONFIG"

		switch operation {
		case consts.OperationStateCreate:
			value = helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)

		case consts.OperationStateDelete:
			value = ""
		}
		box = key + "=" + fmt.Sprintf("\"%s\"", value)
		log.Note("KUBECONFIG env var", key, value)

	case "linux", "darwin":

		switch operation {
		case consts.OperationStateCreate:
			key = "export KUBECONFIG"
			value = helpers.GetPath(consts.UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)
			box = key + "=" + fmt.Sprintf("\"%s\"", value)
			log.Note("KUBECONFIG env var", key, value)

		case consts.OperationStateDelete:
			key = "unset KUBECONFIG"
			box = key
			log.Note(key)
		}
	}

	log.Box("KUBECONFIG env var", box)
}

func isValidK3sVersion(ver string) bool {
	validVersion := []string{"1.27.4", "1.27.1", "1.26.7", "1.25.12"} // TODO: check

	for _, vver := range validVersion {
		if vver == ver {
			return true
		}
	}
	log.Error(strings.Join(validVersion, " "))
	return false
}
