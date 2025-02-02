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
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	azureMeta "github.com/ksctl/ksctl/v2/pkg/provider/azure/meta"
)

type Controller struct {
	ctx    context.Context
	l      logger.Logger
	b      *controller.Controller
	client *controller.Client
	cc     provider.ProvisionMetadata
}

func NewController(ctx context.Context, log logger.Logger, client *controller.Client) (*Controller, error) {

	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-metadata")
	cc.b = controller.NewBaseController(ctx, log)
	cc.l = log
	cc.client = client

	if err := cc.b.ValidateMetadata(client); err != nil {
		return nil, err
	}

	if err := cc.b.ValidateClusterType(client.Metadata.ClusterType); err != nil {
		return nil, err
	}

	if err := cc.b.StartPoller(); err != nil {
		return nil, err
	}

	var err error
	switch cc.client.Metadata.Provider {
	case consts.CloudAzure:
		cc.cc, err = azureMeta.NewAzureMeta(cc.ctx, cc.l)
	}
	if err != nil {
		return nil, err
	}

	return cc, nil
}
