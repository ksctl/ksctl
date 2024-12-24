// Copyright 2024 Ksctl Authors
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
	"github.com/ksctl/ksctl/pkg/certs"
	"github.com/ksctl/ksctl/pkg/providers"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"
	"sync"

	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/utilities"
)

type PreBootstrap struct {
	l   logger.Logger
	ctx context.Context
	mu  *sync.Mutex

	state *statefile.StorageDocument
	store storage.Storage
}

func NewPreBootStrap(
	parentCtx context.Context,
	parentLog logger.Logger,
	state *statefile.StorageDocument,
	store storage.Storage,
) *PreBootstrap {

	p := &PreBootstrap{mu: &sync.Mutex{}}

	p.ctx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "bootstrap")
	p.l = parentLog
	p.state = state
	p.store = store

	return p
}

func (p *PreBootstrap) Setup(
	cloudState providers.CloudResourceState,
	operation consts.KsctlOperation,
) error {

	if operation == consts.OperationCreate {
		p.state.K8sBootstrap = &statefile.KubernetesBootstrapState{}
		var err error
		p.state.K8sBootstrap.B.CACert,
			p.state.K8sBootstrap.B.EtcdCert,
			p.state.K8sBootstrap.B.EtcdKey,
			err = certs.GenerateCerts(p.ctx, p.l, cloudState.PrivateIPv4DataStores)
		if err != nil {
			return err
		}
	}

	p.state.K8sBootstrap.B.PublicIPs.ControlPlanes =
		utilities.DeepCopySlice[string](cloudState.IPv4ControlPlanes)

	p.state.K8sBootstrap.B.PrivateIPs.ControlPlanes =
		utilities.DeepCopySlice[string](cloudState.PrivateIPv4ControlPlanes)

	p.state.K8sBootstrap.B.PublicIPs.DataStores =
		utilities.DeepCopySlice[string](cloudState.IPv4DataStores)
	p.state.K8sBootstrap.B.PrivateIPs.DataStores =
		utilities.DeepCopySlice[string](cloudState.PrivateIPv4DataStores)

	p.state.K8sBootstrap.B.PublicIPs.WorkerPlanes =
		utilities.DeepCopySlice[string](cloudState.IPv4WorkerPlanes)

	p.state.K8sBootstrap.B.PublicIPs.LoadBalancer =
		cloudState.IPv4LoadBalancer

	p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer =
		cloudState.PrivateIPv4LoadBalancer

	p.state.K8sBootstrap.B.SSHInfo = statefile.SSHInfo{
		PrivateKey: cloudState.SSHPrivateKey,
		UserName:   cloudState.SSHUserName,
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "Initialized state from Cloud")
	return nil
}
