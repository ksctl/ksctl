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

package cni

import (
	"context"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

const (
	CiliumStandardStackID  stack.ID = "cilium"
	FlannelStandardStackID stack.ID = "flannel"
)

const (
	CiliumComponentID  stack.ComponentID = "cilium"
	FlannelComponentID stack.ComponentID = "flannel"
)

var appsManifests = map[stack.ID]func(stack.ApplicationParams) (stack.ApplicationStack, error){
	CiliumStandardStackID:  CiliumStandardCNI,
	FlannelStandardStackID: FlannelStandardCNI,
}

func getVersionIfItsNotNilAndLatest(ver *string, defaultVer string) string {
	if ver == nil {
		return defaultVer
	}
	if *ver == "latest" {
		return defaultVer
	}
	return *ver
}

func GetCNIs(isKindCluster bool) (addons.ClusterAddons, string) {
	if isKindCluster {
		return addons.ClusterAddons{
			{
				Name:   "cilium",
				Label:  string(consts.K8sKsctl),
				IsCNI:  true,
				Config: nil,
			},
		}, "cilium"
	}

	return addons.ClusterAddons{
		{
			Name:   "cilium",
			Label:  string(consts.K8sKsctl),
			IsCNI:  true,
			Config: nil,
		},
		{
			Name:   "flannel",
			Label:  string(consts.K8sKsctl),
			IsCNI:  true,
			Config: nil,
		},
	}, "cilium"
}

func FetchKsctlStack(ctx context.Context, log logger.Logger, stkID string) (func(stack.ApplicationParams) (stack.ApplicationStack, error), error) {
	fn, ok := appsManifests[stack.ID(stkID)]
	if !ok {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			log.NewError(ctx, "appStack not found", "stkId", stkID),
		)
	}
	return fn, nil
}

func DoesCNIExistOrNot(app stack.KsctlApp, state *statefile.StorageDocument) (isPresent bool) {
	installedApps := state.ProvisionerAddons.Cni
	if app.StackName == installedApps.Name {
		isPresent = true
		return
	}

	return
}
