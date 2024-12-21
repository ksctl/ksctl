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

package k3s

import (
	"context"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	log               types.LoggerFactory
	k3sCtx            context.Context
)

type K3s struct {
	Cni string
	mu  *sync.Mutex
}

func NewClient(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	state *storageTypes.StorageDocument) *K3s {
	k3sCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.K8sK3s))
	log = parentLog

	mainStateDocument = state
	return &K3s{mu: &sync.Mutex{}}
}

func (k3s *K3s) Setup(storage types.StorageFactory, operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap.K3s = &storageTypes.StateConfigurationK3s{}
		mainStateDocument.BootstrapProvider = consts.K8sK3s
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	return nil
}

func scriptKUBECONFIG() types.ScriptCollection {
	collection := helpers.NewScriptCollection()
	collection.Append(types.Script{
		Name:           "k3s kubeconfig",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo cat /etc/rancher/k3s/k3s.yaml
`,
	})

	return collection
}

func (k3s *K3s) K8sVersion(ver string) types.KubernetesBootstrap {
	if v, err := isValidK3sVersion(ver); err == nil {
		mainStateDocument.K8sBootstrap.K3s.K3sVersion = v
		log.Debug(k3sCtx, "Printing", "k3s.K3sVer", v)
		return k3s
	} else {
		log.Error(err.Error())
		return nil
	}
}

func (k3s *K3s) CNI(cni string) (externalCNI bool) {
	log.Debug(k3sCtx, "Printing", "cni", cni)
	switch consts.KsctlValidCNIPlugin(cni) {
	case consts.CNIFlannel, "":
		k3s.Cni = string(consts.CNIFlannel)
		return false

	default:
		// this tells us that CNI should be installed via the k8s client
		k3s.Cni = string(consts.CNINone)
		return true
	}
}
