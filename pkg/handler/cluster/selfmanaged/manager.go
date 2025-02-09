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

package selfmanaged

import (
	"context"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

type Controller struct {
	ctx context.Context
	l   logger.Logger
	p   *controller.Client
	b   *controller.Controller
	s   *statefile.StorageDocument
}

func NewController(ctx context.Context, log logger.Logger, controllerPayload *controller.Client) (*Controller, error) {

	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "controller-selfmanaged")
	cc.b = controller.NewBaseController(ctx, log)
	cc.p = controllerPayload
	cc.s = new(statefile.StorageDocument)
	cc.l = log

	if err := cc.b.ValidateMetadata(controllerPayload); err != nil {
		return nil, err
	}

	if err := cc.b.ValidateName(controllerPayload.Metadata.ClusterName); err != nil {
		return nil, err
	}

	if cc.p.Metadata.ClusterType != consts.ClusterTypeSelfMang {
		err := cc.l.NewError(cc.ctx, "this feature is only for selfmanaged clusters")
		return nil, err
	}

	if err := cc.b.InitStorage(controllerPayload); err != nil {
		return nil, err
	}

	if err := cc.b.StartPoller(); err != nil {
		return nil, err
	}

	return cc, nil
}
