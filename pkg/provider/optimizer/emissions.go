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

package optimizer

import "github.com/ksctl/ksctl/v2/pkg/provider"

func (k *Optimizer) AttachEmissionsToRegions() ([]provider.RegionOutput, error) {
	for idx, region := range k.AvailRegions {
		sku := region.Sku

		emission, err := k.GetZonalEmissions(sku)
		if err != nil {
			k.l.Error("Failed to get emissions for region", sku, ":", err)
			return nil, err
		}

		k.AvailRegions[idx].Emission = emission
	}
	return k.AvailRegions, nil
}

func (k *Optimizer) GetZonalEmissions(sku string) (*provider.ZonalEmission, error) {
	return nil, nil
}
