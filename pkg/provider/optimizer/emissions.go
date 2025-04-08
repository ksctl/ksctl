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
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/provider"
)

func patchZones(emZone string) (updatedZone string) {
	// This is a patch for old mappings, Refer: https://github.com/Green-Software-Foundation/real-time-cloud/pull/85
	switch emZone {
	case "AUS-NSW":
		updatedZone = "AU-NSW"
	case "AUS-VIC":
		updatedZone = "AU-VIC"
	case "IT":
		updatedZone = "IT-NO"
	case "IN-MH":
		updatedZone = "IN-WE"
	case "IN-UP":
		updatedZone = "IN-NO"
	case "IND":
		updatedZone = "IN-SO"
	case "DK":
		updatedZone = "DK-DK2"
	default:
		updatedZone = emZone
	}
	return
}

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
		em = patchZones(em)

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

// AttachEmissionsToRegions attaches emissions data to the available regions.
//
//	Make sure the all availableregions data is populated first!
func (k *Optimizer) AttachEmissionsToRegions(cloudProvider consts.KsctlCloud) (provider.RegionsOutput, error) {
	emZones, err := k.GetMappingCloudRegionToElectricityMapsZones()
	if err != nil {
		return nil, err
	}

	resultChan := make(chan struct {
		idx      int
		emission *provider.RegionalEmission
		err      error
	}, len(k.AvailRegions))

	for idx, region := range k.AvailRegions {
		go func(idx int, sku string) {
			emission, err := k.getZonalEmissions(emZones, cloudProvider, sku)
			resultChan <- struct {
				idx      int
				emission *provider.RegionalEmission
				err      error
			}{idx, emission, err}
		}(idx, region.Sku)
	}

	for i := 0; i < len(k.AvailRegions); i++ {
		result := <-resultChan
		if result.err != nil {
			k.l.Error("Failed to get emissions for region", k.AvailRegions[result.idx].Sku, ":", result.err)
			return nil, result.err
		}
		k.AvailRegions[result.idx].Emission = result.emission
	}

	close(resultChan)

	return k.AvailRegions, nil
}

type zonalYearMonthlyEmission struct {
	Year  int
	Month string

	ZoneId string
	// DirectCarbonIntensity has unit gCO2eq/kWh
	DirectCarbonIntensity float64
	// LCACarbonIntensity has unit gCO2eq/kWh
	LCACarbonIntensity  float64
	LowCarbonPercentage float64
	RenewablePercentage float64
	DataSource          string
	Unit                string
}

func (k *Optimizer) getMonthlyPastData(zoneId string) ([]zonalYearMonthlyEmission, error) {
	url := fmt.Sprintf("https://data.electricitymaps.com/2025-01-27/%s_2024_monthly.csv", zoneId)

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

	if len(records) < 2 {
		return nil, fmt.Errorf("no data found in CSV")
	}

	var data []zonalYearMonthlyEmission
	for i, row := range records[1:] { // Skipping header row
		if len(row) < 9 {
			return nil, fmt.Errorf("unexpected number of columns in row %d", i+1)
		}

		timestamp, err := time.Parse("2006-01-02 15:04:05", row[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing date on row %d: %v", i+1, err)
		}

		directCI, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing direct carbon intensity on row %d: %v", i+1, err)
		}

		lcaCI, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing LCA carbon intensity on row %d: %v", i+1, err)
		}

		lowCarbonPercent, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing low carbon percentage on row %d: %v", i+1, err)
		}

		renewablePercent, err := strconv.ParseFloat(row[7], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing renewable percentage on row %d: %v", i+1, err)
		}

		data = append(data, zonalYearMonthlyEmission{
			Year:                  timestamp.Year(),
			Month:                 timestamp.Month().String(),
			ZoneId:                row[3],
			DirectCarbonIntensity: directCI,
			LCACarbonIntensity:    lcaCI,
			LowCarbonPercentage:   lowCarbonPercent,
			RenewablePercentage:   renewablePercent,
			DataSource:            row[8],
			Unit:                  "gCO2eq/kWh",
		})
	}

	return data, nil
}

func (k *Optimizer) getZonalEmissions(emZones map[string]map[string]string, cloudProvider consts.KsctlCloud, sku string) (*provider.RegionalEmission, error) {
	em := &provider.RegionalEmission{
		CalcMethod: "average",
		Unit:       "gCO2eq/kWh",
	}

	regMapping, ok := emZones[string(cloudProvider)]
	if !ok {
		return nil, fmt.Errorf("cloud provider %s not found in emissions mapping", cloudProvider)
	}

	emZone, ok := regMapping[sku]
	if !ok {
		k.l.Debug(k.ctx, "No emissions mapping found for region", "region", sku)
		return nil, nil
	}

	v, err := k.getMonthlyPastData(emZone)
	if err != nil {
		k.l.Debug(k.ctx, "Failed to get emissions data", "region", sku, "reason", err)
		return nil, nil
	}

	for _, _v := range v {
		if _v.ZoneId == emZone {
			em.DirectCarbonIntensity += _v.DirectCarbonIntensity
			em.LCACarbonIntensity += _v.LCACarbonIntensity
			em.LowCarbonPercentage += _v.LowCarbonPercentage
			em.RenewablePercentage += _v.RenewablePercentage
		}
	}

	// average the values
	em.DirectCarbonIntensity /= float64(len(v))
	em.LCACarbonIntensity /= float64(len(v))
	em.LowCarbonPercentage /= float64(len(v))
	em.RenewablePercentage /= float64(len(v))

	return em, nil
}

type embodiedCo2Emissions map[consts.KsctlCloud]map[string]provider.VMEmbodied

func (k *Optimizer) getEmbodiedEmissions() (embodiedCo2Emissions, error) {
	url := "https://raw.githubusercontent.com/ksctl/components/c4e4bf2a768acf497cf711e15fed985ebc870dfb/co2/embodied_emissions.csv"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
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
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	records = records[1:]

	type coordinates struct {
		start, end int
	}

	azureRecords := coordinates{0, 6}
	gcpRecords := coordinates{7, 13}
	awsRecords := coordinates{14, 20}

	dataAws := make(map[string]provider.VMEmbodied, 200)
	dataAzure := make(map[string]provider.VMEmbodied, 200)
	dataGcp := make(map[string]provider.VMEmbodied, 100)

	for _, record := range records {
		aws := record[awsRecords.start:awsRecords.end]
		azure := record[azureRecords.start:azureRecords.end]
		gcp := record[gcpRecords.start:gcpRecords.end]

		if len(gcp[1]) > 0 {
			gcpEmission, _ := strconv.ParseFloat(gcp[4], 64)
			dataGcp[gcp[1]] = provider.VMEmbodied{
				EmboddedCo2: gcpEmission,
				Co2Unit:     "kgCO2eq",
			}
		}

		if len(azure[1]) > 0 && (strings.HasPrefix(azure[1], "E") ||
			strings.HasPrefix(azure[1], "D") ||
			strings.HasPrefix(azure[1], "B") ||
			strings.HasPrefix(azure[1], "F")) {
			azureEmission, _ := strconv.ParseFloat(azure[4], 64)
			sku := "Standard_" + strings.ReplaceAll(azure[1], " ", "_")
			dataAzure[sku] = provider.VMEmbodied{
				EmboddedCo2: azureEmission,
				Co2Unit:     "kgCO2eq",
			}
		}

		if len(aws[1]) > 0 && (strings.HasPrefix(aws[1], "m5") ||
			strings.HasPrefix(aws[1], "c5") ||
			strings.HasPrefix(aws[1], "r5") ||
			strings.HasPrefix(aws[1], "t3") ||
			strings.HasPrefix(aws[1], "m7i") ||
			strings.HasPrefix(aws[1], "m7a") ||
			strings.HasPrefix(aws[1], "m8g")) {
			awsEmission, _ := strconv.ParseFloat(aws[4], 64)
			dataAws[aws[1]] = provider.VMEmbodied{
				EmboddedCo2: awsEmission,
				Co2Unit:     "kgCO2eq",
			}
		}
	}

	return embodiedCo2Emissions{
		consts.CloudAzure: dataAzure,
		consts.CloudAws:   dataAws,
		consts.CloudGcp:   dataGcp,
	}, nil
}

func (k *Optimizer) AttachEmbodiedToInstanceType(
	instances []provider.InstanceRegionOutput,
	cloudProvider consts.KsctlCloud,
) (provider.InstancesRegionOutput, error) {
	mappings, err := k.getEmbodiedEmissions()
	if err != nil {
		return nil, err
	}

	data, ok := mappings[cloudProvider]
	if !ok {
		return nil, fmt.Errorf("cloud provider %s not found in emissions mapping", cloudProvider)
	}
	for i, instance := range instances {
		if v, ok := data[instance.Sku]; ok {
			instances[i].EmboddedEmissions = &v
		} else {
			k.l.Debug(k.ctx, "No embodied emissions mapping found for instance", "instance", instance.Sku)
		}
	}

	return instances, nil
}
