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

package kcm

import (
	"context"
	"errors"
	"fmt"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"

	bootstrapHandler "github.com/ksctl/ksctl/v2/pkg/bootstrap/handler"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/provider/aws"
	"github.com/ksctl/ksctl/v2/pkg/provider/azure"
	"github.com/ksctl/ksctl/v2/pkg/provider/local"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

const (
	Sku = "kcm"
)

func GetAvailableVersions() ([]string, error) {
	return poller.GetSharedPoller().Get("ksctl", Sku)
}

type Kcm struct {
	ctx context.Context
	l   logger.Logger
	p   *controller.Client
	b   *controller.Controller
	s   *statefile.StorageDocument
}

func NewKcm(
	ctx context.Context,
	log logger.Logger,
	controllerPayload *controller.Client,
	b *controller.Controller,
	s *statefile.StorageDocument,
) (*Kcm, error) {
	cc := new(Kcm)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "addon-kcm")
	cc.l = log
	cc.p = controllerPayload
	cc.b = b
	cc.s = s

	return cc, nil
}

func (k *Kcm) Install(version string) (errC error) {

	defer func() {
		if errC != nil {
			v := k.b.PanicHandler(k.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	transferableInfraState, err := k.prepareClient()
	if err != nil {
		return err
	}

	s := k.s.ProvisionerAddons.Apps
	for _, addon := range s {
		if addon.Name == Sku && addon.For == consts.K8sKsctl {
			k.l.Box(k.ctx, "Addon", fmt.Sprintf("Addon %s@%s already installed\nIf you want to install a newer version you need to perform reinstall", Sku, *addon.Version))
			return nil
		}
	}

	ca := statefile.SlimProvisionerAddon{
		Name:    Sku,
		For:     consts.K8sKsctl,
		Version: &version,
	}

	kubeconfig, err := k.p.Cloud.GetKubeconfig()
	if err != nil {
		return err
	}
	kbc, err := bootstrapHandler.NewController(
		k.ctx,
		k.l,
		k.b,
		k.s,
		consts.OperationGet,
		transferableInfraState,
		k.p,
	)
	if err != nil {
		return err
	}

	return kbc.InstallKcm(kubeconfig, "ksctl-system", "cluster-config", ca)
}

func (k *Kcm) Uninstall() (errC error) {
	defer func() {
		if errC != nil {
			v := k.b.PanicHandler(k.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	transferableInfraState, err := k.prepareClient()
	if err != nil {
		return err
	}

	s := k.s.ProvisionerAddons.Apps
	var app statefile.SlimProvisionerAddon
	found := false
	for _, addon := range s {
		if addon.Name == Sku && addon.For == consts.K8sKsctl {
			found = true
			app = addon
			break
		}
	}
	if !found {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			fmt.Errorf("Addon %s not found", Sku),
		)
	}

	k.l.Box(k.ctx, "Addon", fmt.Sprintf("Addon %s@%s is going to be uninstalled.\n Make sure you cleanup before performing this action", Sku, *app.Version))

	kubeconfig, err := k.p.Cloud.GetKubeconfig()
	if err != nil {
		return err
	}
	kbc, err := bootstrapHandler.NewController(
		k.ctx,
		k.l,
		k.b,
		k.s,
		consts.OperationGet,
		transferableInfraState,
		k.p,
	)
	if err != nil {
		return err
	}

	return kbc.UninstallKcm(kubeconfig, "ksctl-system", "cluster-config", app)
}

func (kc *Kcm) prepareClient() (_ *provider.CloudResourceState, errC error) {
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
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, azure.ProvideClient)

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, aws.ProvideClient)
		if err != nil {
			break
		}

	case consts.CloudLocal:
		kc.p.Cloud, err = local.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, local.ProvideClient)

	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if errInit := kc.p.Cloud.InitState(consts.OperationGet); errInit != nil {
		kc.l.Error("handled error", "catch", errInit)
		return nil, errInit
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if kc.p.Metadata.ClusterType == consts.ClusterTypeSelfMang {
		transferableInfraState, errState := kc.p.Cloud.GetStateForHACluster()
		if errState != nil {
			kc.l.Error("handled error", "catch", errState)
			return nil, err
		}

		return &transferableInfraState, nil
	}
	return nil, nil
}
