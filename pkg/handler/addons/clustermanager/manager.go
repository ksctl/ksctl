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

package clustermanager

import (
	"context"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/provider"
	"github.com/ksctl/ksctl/pkg/provider/aws"
	"github.com/ksctl/ksctl/pkg/provider/azure"
	"github.com/ksctl/ksctl/pkg/provider/civo"
	"github.com/ksctl/ksctl/pkg/provider/local"
	"github.com/ksctl/ksctl/pkg/statefile"
)

type Controller struct {
	ctx context.Context
	l   logger.Logger
	p   *controller.Client
	b   *controller.Controller
	s   *statefile.StorageDocument
}

// NewController intended to be used by the cli to enable or disable addon 'ksctl-clustermanager'
func NewController(ctx context.Context, log logger.Logger, controllerPayload *controller.Client) (*Controller, error) {

	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "addons-ksctl-clustermanager")
	cc.b = controller.NewBaseController(ctx, log)
	cc.p = controllerPayload
	cc.s = new(statefile.StorageDocument)
	cc.l = log

	if err := cc.b.ValidateMetadata(controllerPayload); err != nil {
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

func (kc *Controller) helper() (*provider.CloudResourceState, error) {
	if kc.b.IsLocalProvider(kc.p) {
		kc.p.Metadata.Region = "LOCAL"
	}

	clusterType := consts.ClusterTypeMang
	if kc.b.IsHA(kc.p) {
		clusterType = consts.ClusterTypeHa
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		clusterType); err != nil {

		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudCivo:
		kc.p.Cloud, err = civo.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, civo.ProvideClient)

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

	transferableInfraState, errState := kc.p.Cloud.GetStateForHACluster()
	if errState != nil {
		kc.l.Error("handled error", "catch", errState)
		return nil, err
	}

	return &transferableInfraState, nil
}
