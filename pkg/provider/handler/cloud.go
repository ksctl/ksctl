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

package handler

import (
	"context"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	awsPkg "github.com/ksctl/ksctl/v2/pkg/provider/aws"
	azurePkg "github.com/ksctl/ksctl/v2/pkg/provider/azure"
	localPkg "github.com/ksctl/ksctl/v2/pkg/provider/local"
)

type Controller struct {
	ctx context.Context
	ksc context.Context
	l   logger.Logger
	p   *controller.Client
	b   *controller.Controller
	s   *statefile.StorageDocument
}

func NewController(
	ctx context.Context,
	log logger.Logger,
	ksc context.Context,
	baseController *controller.Controller,
	state *statefile.StorageDocument,
	operation consts.KsctlOperation,
	controllerPayload *controller.Client,
) (*Controller, error) {

	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "pkg/providers/handler/cloud")
	cc.l = log
	cc.b = baseController
	cc.p = controllerPayload
	cc.s = state

	cc.ksc = ksc

	err := cc.setupInterfaces(operation)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func (kc *Controller) setupInterfaces(operation consts.KsctlOperation) error {

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		kc.p.Cloud, err = azurePkg.NewClient(kc.ctx, kc.l, kc.ksc, kc.p.Metadata, kc.s, kc.p.Storage, azurePkg.ProvideClient)

		if err != nil {
			return err
		}
	case consts.CloudAws:
		kc.p.Cloud, err = awsPkg.NewClient(kc.ctx, kc.l, kc.ksc, kc.p.Metadata, kc.s, kc.p.Storage, awsPkg.ProvideClient)

		if err != nil {
			return err
		}

	case consts.CloudLocal:
		kc.p.Cloud, err = localPkg.NewClient(kc.ctx, kc.l, kc.ksc, kc.p.Metadata, kc.s, kc.p.Storage, localPkg.ProvideClient)

		if err != nil {
			return err
		}
	}
	err = kc.p.Cloud.InitState(operation)
	if err != nil {
		return err
	}
	return nil
}

func pauseOperation(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}
