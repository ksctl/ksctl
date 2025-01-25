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
	bootstrapHandler "github.com/ksctl/ksctl/v2/pkg/bootstrap/handler"
	"github.com/ksctl/ksctl/v2/pkg/consts"
)

func (kc *Controller) Disable() error {
	defer kc.b.PanicHandler(kc.l)

	transferableInfraState, err := kc.helper()
	if err != nil {
		return err
	}

	kubeconfig, err := kc.p.Cloud.GetKubeconfig()
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	kbc, err := bootstrapHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationGet,
		transferableInfraState,
		kc.p,
	)
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	if errAddon := kbc.DisableKsctlAddons(kubeconfig, "ksctl-system", "cluster-config"); errAddon != nil {
		kc.l.Error("handled error", "catch", errAddon)
		return errAddon
	}

	kc.l.Success(kc.ctx, "ksctl addons disabled")

	return nil
}
