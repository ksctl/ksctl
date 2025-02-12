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

package addons

import (
	"context"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/handler/addons/kcm"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

type AddonController struct {
	ctx context.Context
	l   logger.Logger
	p   *controller.Client
	b   *controller.Controller
	s   *statefile.StorageDocument
}

// NewController intended to be used by the cli to enable or disable addon 'ksctl-clustermanager'
func NewController(ctx context.Context, log logger.Logger, controllerPayload *controller.Client) (*AddonController, error) {

	cc := new(AddonController)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-addons")
	cc.b = controller.NewBaseController(ctx, log)
	cc.p = controllerPayload
	cc.s = new(statefile.StorageDocument)
	cc.l = log

	if err := cc.b.ValidateMetadata(controllerPayload); err != nil {
		return nil, err
	}

	if err := cc.b.ValidateClusterType(controllerPayload.Metadata.ClusterType); err != nil {
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

func (ac *AddonController) ListAllAddons() ([]string, error) {
	return []string{kcm.Sku}, nil
}

func (ac *AddonController) ListAvailableVersions(addonSku string) ([]string, error) {
	switch addonSku {
	case kcm.Sku:
		return kcm.GetAvailableVersions()
	default:
		return nil, errors.WrapErrorf(
			errors.ErrInvalidKsctlClusterAddons,
			"addon %s is not supported", addonSku)
	}
}

type KsctlAddon interface {
	Install(version string) error
	Uninstall() error
}

func (ac *AddonController) GetAddon(sku string) (KsctlAddon, error) {
	switch sku {
	case kcm.Sku:
		return kcm.NewKcm(ac.ctx, ac.l, ac.p, ac.b, ac.s)
	default:
		return nil, errors.WrapErrorf(
			errors.ErrInvalidKsctlClusterAddons,
			"addon %s is not supported", sku)
	}
}
