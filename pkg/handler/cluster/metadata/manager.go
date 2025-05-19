// Copyright 2025 Ksctl Authors
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

package metadata

import (
	"context"

	bootstrapMeta "github.com/ksctl/ksctl/v2/pkg/bootstrap/meta"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	awsMeta "github.com/ksctl/ksctl/v2/pkg/provider/aws/meta"
	azureMeta "github.com/ksctl/ksctl/v2/pkg/provider/azure/meta"
	localMeta "github.com/ksctl/ksctl/v2/pkg/provider/local/meta"
)

type Controller struct {
	ctx    context.Context
	l      logger.Logger
	b      *controller.Controller
	client *controller.Client
	cc     provider.ProvisionMetadata
	bb     *bootstrapMeta.BootstrapMetadata
}

func NewController(
	ctx context.Context,
	log logger.Logger,
	ksctlConfig controller.KsctlWorkerConfiguration,
	client *controller.Client,
) (*Controller, error) {

	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-metadata")
	cc.b = controller.NewBaseController(ctx, log, ksctlConfig)
	cc.l = log
	cc.client = client

	if err := cc.b.ValidateMetadata(client); err != nil {
		return nil, err
	}

	if err := cc.b.ValidateClusterType(client.Metadata.ClusterType); err != nil {
		return nil, err
	}

	if err := cc.b.StartPoller(cc.b.KsctlWorkloadConf.PollerCache); err != nil {
		return nil, err
	}

	var err error
	switch cc.client.Metadata.Provider {
	case consts.CloudAzure:
		cc.cc, err = azureMeta.NewAzureMeta(cc.ctx, cc.l)
	case consts.CloudAws:
		cc.cc, err = awsMeta.NewAwsMeta(cc.ctx, cc.l)
	case consts.CloudLocal:
		cc.cc, err = localMeta.NewLocalMeta(cc.ctx, cc.l)
	}
	if err != nil {
		return nil, err
	}

	if err := cc.cc.Connect(cc.b.KsctlWorkloadConf.WorkerCtx); err != nil {
		return nil, err
	}

	if client.Metadata.ClusterType == consts.ClusterTypeSelfMang {
		cc.bb = bootstrapMeta.NewBootstrapMetadata(client.Metadata.K8sDistro)
	}

	return cc, nil
}
