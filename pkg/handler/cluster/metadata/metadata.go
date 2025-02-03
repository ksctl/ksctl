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
	"errors"

	"github.com/ksctl/ksctl/v2/pkg/provider"
)

func (kc *Controller) ListAllRegions() (
	_ []provider.RegionOutput,
	errC error,
) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.client) {
		return nil, nil
	}

	regions, err := kc.cc.GetAvailableRegions()
	if err != nil {
		return nil, err
	}

	return regions, nil
}

func (kc *Controller) ListAllInstances(region string) (
	out map[string]provider.InstanceRegionOutput,
	errC error,
) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.client) {
		return nil, nil
	}

	instances, err := kc.cc.GetAvailableInstanceTypes(region, kc.client.Metadata.ClusterType)
	if err != nil {
		return nil, err
	}

	out = make(map[string]provider.InstanceRegionOutput)
	for _, v := range instances {
		out[v.Sku] = v
	}

	return out, nil
}
