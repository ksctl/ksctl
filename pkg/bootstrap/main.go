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

package bootstrap

import (
	"context"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
	"github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	log               types.LoggerFactory
	bootstrapCtx      context.Context
)

func NewPreBootStrap(parentCtx context.Context, parentLog types.LoggerFactory,
	state *storageTypes.StorageDocument) *PreBootstrap {

	bootstrapCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "bootstrap")
	log = parentLog

	mainStateDocument = state
	return &PreBootstrap{mu: &sync.Mutex{}}
}

func (p *PreBootstrap) Setup(cloudState cloud.CloudResourceState,
	storage types.StorageFactory, operation consts.KsctlOperation) error {

	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap = &storageTypes.KubernetesBootstrapState{}
		var err error
		mainStateDocument.K8sBootstrap.B.CACert,
			mainStateDocument.K8sBootstrap.B.EtcdCert,
			mainStateDocument.K8sBootstrap.B.EtcdKey,
			err = helpers.GenerateCerts(bootstrapCtx, log, cloudState.PrivateIPv4DataStores)
		if err != nil {
			return err
		}
	}

	mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes =
		utilities.DeepCopySlice[string](cloudState.IPv4ControlPlanes)

	mainStateDocument.K8sBootstrap.B.PrivateIPs.ControlPlanes =
		utilities.DeepCopySlice[string](cloudState.PrivateIPv4ControlPlanes)

	mainStateDocument.K8sBootstrap.B.PublicIPs.DataStores =
		utilities.DeepCopySlice[string](cloudState.IPv4DataStores)
	mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores =
		utilities.DeepCopySlice[string](cloudState.PrivateIPv4DataStores)

	mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes =
		utilities.DeepCopySlice[string](cloudState.IPv4WorkerPlanes)

	mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer =
		cloudState.IPv4LoadBalancer

	mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer =
		cloudState.PrivateIPv4LoadBalancer

	mainStateDocument.K8sBootstrap.B.SSHInfo = cloudState.SSHState

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(bootstrapCtx, "Initialized state from Cloud")
	return nil
}
