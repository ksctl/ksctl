// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"fmt"
	"os"
	"path/filepath"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlError "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
)

func generateConfig(noWorker, noControl int, cni bool) ([]byte, error) {
	if noWorker >= 0 && noControl == 0 {
		return nil, ksctlError.ErrInvalidUserInput.Wrap(
			log.NewError(localCtx, "invalid config request control node cannot be 0"),
		)
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
		return nil, ksctlError.ErrInvalidUserInput.Wrap(
			log.NewError(localCtx, "invalid config request control node cannot be 0"),
		)
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
		return nil, err
	}

	log.Debug(localCtx, "Printing", "configCluster", string(raw))

	return cluster.CreateWithRawConfig(raw), nil
}

func isPresent(storage types.StorageFactory, clusterName string) error {
	return storage.AlreadyCreated(consts.CloudLocal, "LOCAL", clusterName, consts.ClusterTypeMang)
}

func createNecessaryConfigs(storeDir string) (string, error) {
	_path := filepath.Join(storeDir, "kubeconfig")

	_, err := os.Create(_path)
	if err != nil {
		return "", ksctlError.ErrInternal.Wrap(
			log.NewError(localCtx, "failed to create file to store kubeconfig", "Reason", err),
		)
	}
	return _path, nil
}

func loadStateHelper(storage types.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return err
	}
	*mainStateDocument = func(x *storageTypes.StorageDocument) storageTypes.StorageDocument {
		return *x
	}(raw)
	return nil
}
