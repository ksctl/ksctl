package local

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
	"sigs.k8s.io/kind/pkg/cluster"
)

func generateConfig(noWorker, noControl int, cni bool) ([]byte, error) {
	if noWorker >= 0 && noControl == 0 {
		return nil, log.NewError("invalid config request control node cannot be 0")
	}
	var config string
	config += `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
`
	for noControl > 0 {
		config += `- role: control-plane
`
		noControl--
	}

	for noWorker > 0 {
		config += `- role: worker
`
		noWorker--
	}

	config += fmt.Sprintf(`networking:
  disableDefaultCNI: %v
`, cni)

	config += `...`

	return []byte(config), nil
}

func configOption(noOfNodes int, cni bool) (cluster.CreateOption, error) {

	if noOfNodes < 1 {
		return nil, log.NewError("invalid config request control node cannot be 0")
	}
	if noOfNodes == 1 {
		var config string
		config += fmt.Sprintf(`---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
networking:
  disableDefaultCNI: %v
...`, cni)
		return cluster.CreateWithRawConfig([]byte(config)), nil
	}
	//control := noOfNodes / 2 // derive the math
	control := 1
	worker := noOfNodes - control
	raw, err := generateConfig(worker, control, cni)
	if err != nil {
		return nil, fmt.Errorf("ERR in node config generation")
	}

	log.Debug("Printing", "configCluster", string(raw))

	return cluster.CreateWithRawConfig(raw), nil
}

func isPresent(storage resources.StorageFactory, cluster string) bool {
	_, err := storage.Path(utils.GetPath(consts.UtilOtherPath, consts.CloudLocal, consts.ClusterTypeMang, cluster, STATE_FILE)).Load()
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func createNecessaryConfigs(storage resources.StorageFactory, clusterName string) (string, error) {

	var err error

	kpath := utils.GetPath(consts.UtilOtherPath, consts.CloudLocal, consts.ClusterTypeMang, clusterName, KUBECONFIG)

	err = storage.Permission(0755).
		Path(kpath).Save([]byte(""))
	if err != nil {
		return "", log.NewError(err.Error())
	}

	err = saveStateHelper(storage, utils.GetPath(consts.UtilOtherPath, consts.CloudLocal, consts.ClusterTypeMang, clusterName, STATE_FILE))
	if err != nil {
		return "", log.NewError(err.Error())
	}

	return kpath, nil
}

func printKubeconfig(storage resources.StorageFactory, operation consts.KsctlOperation, clustername string) {
	key := ""
	value := ""
	box := ""
	switch runtime.GOOS {
	case "windows":
		key = "$Env:KUBECONFIG"

		switch operation {
		case consts.OperationStateCreate:
			value = utils.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, clustername, KUBECONFIG)

		case consts.OperationStateDelete:
			value = ""
		}
		box = key + "=" + fmt.Sprintf("\"%s\"", value)
		log.Note("KUBECONFIG env var", key, value)

	case "linux", "macos":

		switch operation {
		case consts.OperationStateCreate:
			key = "export KUBECONFIG"
			value = utils.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, clustername, KUBECONFIG)
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

func saveStateHelper(storage resources.StorageFactory, path string) error {
	rawState, err := convertStateToBytes(*localState)
	if err != nil {
		return log.NewError(err.Error())
	}
	return storage.Path(path).Permission(0755).Save(rawState)
}

func loadStateHelper(storage resources.StorageFactory, path string) error {
	raw, err := storage.Path(path).Load()
	if err != nil {
		return log.NewError(err.Error())
	}

	return convertStateFromBytes(raw)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func convertStateFromBytes(raw []byte) error {
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		return log.NewError(err.Error())
	}
	localState = data
	return nil
}
