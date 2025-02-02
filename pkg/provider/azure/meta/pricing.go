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

package meta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

type AzPrice struct {
	CurrencyCode  string  `json:"currencyCode"`
	UnitPrice     float64 `json:"unitPrice"`
	ArmRegionName string  `json:"armRegionName"`
	ProductName   string  `json:"productName"`
	MeterName     string  `json:"meterName"`
	ArmSkuName    string  `json:"armSkuName"`
	SkuName       string  `json:"skuName"`
	ServiceName   string  `json:"serviceName"`
	UnitOfMeasure string  `json:"unitOfMeasure"`
}

type APIPricingResponse struct {
	Items        []AzPrice `json:"Items"`
	NextPageLink string    `json:"NextPageLink"`
}

type options struct {
	ignoreSpotAndLowPriMeterName bool
}

type option func(op *options) error

func IgnoreSpotAndLowPriMeterName() option {
	return func(op *options) error {
		op.ignoreSpotAndLowPriMeterName = true
		return nil
	}
}

func doesMeterNameContainSpotOrLowPri(meterName string) bool {
	return strings.HasSuffix(meterName, "Spot") || strings.HasSuffix(meterName, "Low Priority")
}

func doesProductNameContainWindows(productName string) bool {
	return strings.HasSuffix(productName, "Windows")
}

func fetchPrices(query string, opts ...option) ([]AzPrice, error) {
	apiURL := "https://prices.azure.com/api/retail/prices"
	v := make([]AzPrice, 0, 100)
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	client := &http.Client{}
	nextPage := fmt.Sprintf("%s?$filter=%s", apiURL, query)

	for nextPage != "" {
		req, err := http.NewRequest("GET", nextPage, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			return nil, err
		}

		defer resp.Body.Close()

		var jsonData APIPricingResponse
		if err := json.NewDecoder(resp.Body).Decode(&jsonData); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return nil, err
		}
		for _, p := range jsonData.Items {
			if doesProductNameContainWindows(p.ProductName) {
				continue
			}
			if options.ignoreSpotAndLowPriMeterName &&
				doesMeterNameContainSpotOrLowPri(p.MeterName) {
				continue
			}
			v = append(v, p)
		}

		nextPage = jsonData.NextPageLink
	}

	return v, nil
}

func (m *AzureMeta) priceDisksStandardLRS_ESeries(regionSku string) (map[string]provider.StorageBlockRegionOutput, error) {
	filter := fmt.Sprintf("armRegionName eq '%s' and serviceFamily eq 'Storage' and productName eq 'Standard SSD Managed Disks' and endswith(skuName, 'LRS') and startswith(skuName, 'E') and endswith(meterName, 'LRS Disk') and Type eq 'Consumption' and unitOfMeasure eq '1/Month'", regionSku)

	prices, err := fetchPrices(filter)
	if err != nil {
		return nil, err
	}
	o := make(map[string]provider.StorageBlockRegionOutput)
	for _, p := range prices {
		o[strings.Split(p.SkuName, " ")[0]] = provider.StorageBlockRegionOutput{
			Price: &provider.PriceOutput{
				MonthlyPrice: utilities.Ptr(prices[0].UnitPrice),
				Currency:     p.CurrencyCode,
			},
		}
	}

	return o, nil
}

func (m *AzureMeta) priceVMs(regionSku string) (map[string]provider.InstanceRegionOutput, error) {
	filter := fmt.Sprintf("serviceName eq 'Virtual Machines' and armRegionName eq '%s' and serviceFamily eq 'Compute' and type eq 'Consumption' and unitOfMeasure eq '1 Hour' and startswith(armSkuName, 'Standard_')", regionSku)

	prices, err := fetchPrices(filter, IgnoreSpotAndLowPriMeterName())
	if err != nil {
		return nil, err
	}

	o := make(map[string]provider.InstanceRegionOutput)
	for _, p := range prices {
		o[p.ArmSkuName] = provider.InstanceRegionOutput{
			Price: provider.PriceOutput{
				HourlyPrice: utilities.Ptr(prices[0].UnitPrice),
				Currency:    p.CurrencyCode,
			},
		}
	}

	return o, nil
}

func (m *AzureMeta) priceAksManagement(regionSku string) (out []provider.ManagedClusterOutput, _ error) {
	filter := fmt.Sprintf("serviceName eq 'Azure Kubernetes Service' and armRegionName eq '%s' and unitOfMeasure eq '1 Hour' and skuName eq 'Standard'", regionSku)

	prices, err := fetchPrices(filter)
	if err != nil {
		return nil, err
	}

	out = append(out, provider.ManagedClusterOutput{
		Sku: "Standard Free",
		Price: provider.PriceOutput{
			HourlyPrice: utilities.Ptr(0.0),
			Currency:    prices[0].CurrencyCode,
		},
		Tier: "Free",
	})

	for _, p := range prices {
		tier := ""
		if p.MeterName == "Standard Long Term Support" {
			tier = "Premium"
		} else if p.MeterName == "Standard Uptime SLA" {
			tier = "Standard"
		}

		o := provider.ManagedClusterOutput{
			Sku:  p.MeterName,
			Tier: tier,
			Price: provider.PriceOutput{
				HourlyPrice: utilities.Ptr(prices[0].UnitPrice),
				Currency:    p.CurrencyCode,
			},
		}

		out = append(out, o)
	}

	return out, nil
}

// type AksDetailsOutput struct {
// 	Region        string
// 	Price         float64
// 	ServiceName   string
// 	ProductName   string
// 	UnitOfMeasure string

// 	Compute    ComputeDetailsOutput
// 	TotalPrice float64
// }

// type AKSTier string

// const (
// 	Standard AKSTier = "Standard"
// 	Premium  AKSTier = "Premium"
// 	Free     AKSTier = "Free"
// )

// func (p *ResourceDetails) AKS(
// 	ctx context.Context,
// 	region string,
// 	vmType string,
// 	vmCount int,
// 	tier AKSTier,
// ) (*AksDetailsOutput, error) {
// 	ee := &AksDetailsOutput{}

// 	// By default the azure is free / standard it is user choice so $0 or $0.1/hr
// 	// For the NodeSize the user selects it
// 	// as the default AKS Disk is not managed but emphermal storage no need to consider the disk cost

// 	totalCost := 0.0
// 	if tier == Premium || tier == Standard {
// 		query := ""
// 		if tier == Premium {
// 			query = fmt.Sprintf("serviceName eq 'Azure Kubernetes Service' and armRegionName eq '%s' and meterName eq 'Standard Long Term Support'", region)
// 		} else {
// 			query = fmt.Sprintf("serviceName eq 'Azure Kubernetes Service' and armRegionName eq '%s' and meterName eq 'Standard Uptime SLA'", region)
// 		}

// 		prices, err := fetchPrices(query)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if len(prices) == 0 {
// 			return nil, fmt.Errorf("no pricing found for AKS for region %s, tier %s", region, tier)
// 		}

// 		totalCost += prices[0].UnitPrice * 730
// 		ee.ServiceName = prices[0].ServiceName
// 		ee.ProductName = prices[0].ProductName
// 		ee.Region = prices[0].ArmRegionName
// 		ee.Price = prices[0].UnitPrice
// 		ee.UnitOfMeasure = prices[0].UnitOfMeasure
// 	}

// 	// Get the VMs
// 	vms, err := p.VMs(ctx, region)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if v, ok := vms[VMSku(vmType)]; ok {
// 		ee.Compute = v
// 	}

// 	totalCost += ee.Compute.Price * float64(vmCount) * 730
// 	ee.TotalPrice = totalCost

// 	return ee, nil
// }
