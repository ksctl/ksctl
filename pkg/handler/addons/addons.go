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
	"errors"
	"slices"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/handler/addons/kcm"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider/aws"
	"github.com/ksctl/ksctl/v2/pkg/provider/azure"
	"github.com/ksctl/ksctl/v2/pkg/provider/local"
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
func NewController(ctx context.Context, log logger.Logger, ksctlConfig controller.KsctlWorkerConfiguration, controllerPayload *controller.Client) (*AddonController, error) {

	cc := new(AddonController)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-addons")
	cc.b = controller.NewBaseController(ctx, log, ksctlConfig)
	cc.p = controllerPayload
	cc.s = new(statefile.StorageDocument)
	cc.l = log

	if err := cc.b.ValidateMetadata(controllerPayload); err != nil {
		return nil, err
	}

	if err := cc.b.ValidateClusterType(controllerPayload.Metadata.ClusterType); err != nil {
		return nil, err
	}

	if err := cc.b.InitStorage(controllerPayload, ksctlConfig.WorkerCtx); err != nil {
		return nil, err
	}

	if err := cc.b.StartPoller(cc.b.KsctlWorkloadConf.PollerCache); err != nil {
		return nil, err
	}

	return cc, nil
}

func (ac *AddonController) ListAllAddons() ([]string, error) {
	return []string{kcm.Sku}, nil
}

type InstalledAddon struct {
	Name    string
	Version string
}

func (kc *AddonController) ListInstalledAddons() (_ []InstalledAddon, errC error) {
	if kc.p.Metadata.Provider == consts.CloudLocal {
		kc.p.Metadata.Region = "LOCAL"
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		kc.p.Metadata.ClusterType); err != nil {

		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			if errC != nil {
				errC = errors.Join(errC, err)
			} else {
				errC = err
			}
		}
	}()

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.l, kc.b.KsctlWorkloadConf.WorkerCtx, kc.p.Metadata, kc.s, kc.p.Storage, azure.ProvideClient)

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.l, kc.b.KsctlWorkloadConf.WorkerCtx, kc.p.Metadata, kc.s, kc.p.Storage, aws.ProvideClient)
		if err != nil {
			break
		}

	case consts.CloudLocal:
		kc.p.Cloud, err = local.NewClient(kc.ctx, kc.l, kc.b.KsctlWorkloadConf.WorkerCtx, kc.p.Metadata, kc.s, kc.p.Storage, local.ProvideClient)

	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if errInit := kc.p.Cloud.InitState(consts.OperationGet); errInit != nil {
		kc.l.Error("handled error", "catch", errInit)
		return nil, errInit
	}

	addons := make([]InstalledAddon, 0)
	availableAddons, _ := kc.ListAllAddons()

	for _, app := range kc.s.ProvisionerAddons.Apps {
		if slices.Contains(availableAddons, app.Name) {
			a := InstalledAddon{
				Name: app.Name,
			}

			if app.Version != nil {
				a.Version = *app.Version
			}

			addons = append(addons, a)
		}
	}

	return addons, nil
}

func (ac *AddonController) ListAvailableVersions(addonSku string) ([]string, error) {
	switch addonSku {
	case kcm.Sku:
		return kcm.GetAvailableVersions()
	default:
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInvalidKsctlClusterAddons,
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
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInvalidKsctlClusterAddons,
			"addon %s is not supported", sku)
	}
}
