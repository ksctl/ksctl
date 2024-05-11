package local

import (
	"fmt"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"os"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
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

func isPresent(storage types.StorageFactory, clusterName string) bool {
	err := storage.AlreadyCreated(consts.CloudLocal, "LOCAL", clusterName, consts.ClusterTypeMang)
	return err == nil
}

func createNecessaryConfigs(storeDir string) (string, error) {

	_, err := os.Create(storeDir + helpers.PathSeparator + "kubeconfig")
	if err != nil {
		return "", err
	}
	return storeDir + helpers.PathSeparator + "kubeconfig", nil
}

func loadStateHelper(storage types.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return log.NewError(err.Error())
	}
	*mainStateDocument = func(x *storageTypes.StorageDocument) storageTypes.StorageDocument {
		return *x
	}(raw)
	return nil
}
