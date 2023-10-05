package k3s

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
)

func saveStateHelper(storage resources.StorageFactory, path string) error {
	rawState, err := convertStateToBytes(*k8sState)
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

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func convertStateFromBytes(raw []byte) error {
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	k8sState = data
	return nil
}
func saveKubeconfigHelper(storage resources.StorageFactory, path string, kubeconfig string) error {
	rawState := []byte(kubeconfig)

	return storage.Path(path).Permission(FILE_PERM_CLUSTER_KUBECONFIG).Save(rawState)
}
func printKubeconfig(storage resources.StorageFactory, operation KsctlOperation) {
	env := ""
	storage.Logger().Note("KUBECONFIG env var")
	path := utils.GetPath(CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, KUBECONFIG_FILE_NAME)
	switch runtime.GOOS {
	case "windows":
		switch operation {
		case OPERATION_STATE_CREATE:
			env = fmt.Sprintf("$Env:KUBECONFIG=\"%s\"\n", path)
		case OPERATION_STATE_DELETE:
			env = fmt.Sprintf("$Env:KUBECONFIG=\"\"\n")
		}
	case "linux", "macos":
		switch operation {
		case OPERATION_STATE_CREATE:
			env = fmt.Sprintf("export KUBECONFIG=\"%s\"\n", path)
		case OPERATION_STATE_DELETE:
			env = "unset KUBECONFIG"
		}
	}
	storage.Logger().Note(env)
}

func isValidK3sVersion(ver string) bool {
	validVersion := []string{"1.27.4", "1.27.1", "1.26.7", "1.25.12"} // TODO: check

	for _, vver := range validVersion {
		if vver == ver {
			return true
		}
	}
	var log logger.LogFactory = &logger.Logger{}
	log.Err(strings.Join(validVersion, " "))
	return false
}
