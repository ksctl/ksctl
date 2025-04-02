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

import (
	"encoding/csv"
	"fmt"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"io"
	"net/http"
)

func (k *Optimizer) GetMappingCloudRegionToElectricityMapsZones() (map[string]map[string]string, error) {
	url := "https://raw.githubusercontent.com/Green-Software-Foundation/real-time-cloud/3ba9accf352bcd081c3fe6b0789ee276d5286f82/Cloud_Region_Metadata.csv"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, response: %s",
			res.StatusCode, string(bodyBytes))
	}

	reader := csv.NewReader(res.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	resMapping := make(map[string]map[string]string, len(records))

	cpIdx, crIdx, emZone := -1, -1, -1
	for i, header := range records[0] {
		switch header {
		case "cloud-provider":
			cpIdx = i
		case "cloud-region":
			crIdx = i
		case "em-zone-id":
			emZone = i
		}
	}

	for _, row := range records[1:] {
		cp := row[cpIdx]
		cr := row[crIdx]
		em := row[emZone]

		key := ""

		if cp == "Amazon Web Services" {
			key = "aws"
		} else if cp == "Google Cloud" {
			key = "gcp"
		} else if cp == "Microsoft Azure" {
			key = "azure"
		}

		if len(em) == 0 {
			continue
		}

		if v, ok := resMapping[key]; ok {
			if _, ok := v[cr]; ok {
				v[cr] = em
			} else {
				v[cr] = em
			}
		} else {
			resMapping[key] = make(map[string]string)
			resMapping[key][cr] = em
		}
	}

	return resMapping, nil
}

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
