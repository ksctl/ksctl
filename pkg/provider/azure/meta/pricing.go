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
	filterManagedDisk            *string
	filterInstanceType           *string
	filterAksOfferings           *string
}

type Option func(op *options) error

func IgnoreSpotAndLowPriMeterName() Option {
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

func fetchPrices(query string, opts ...Option) ([]AzPrice, error) {
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

func WithDefaultManagedDisk() Option {
	return func(op *options) error {
		op.filterManagedDisk = utilities.Ptr("skuName eq 'E6 LRS'")
		return nil
	}
}

func (m *AzureMeta) priceDisks(regionSku string, opts ...Option) ([]provider.StorageBlockRegionOutput, error) {

	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	filterForDisk := "endswith(skuName, 'LRS') and startswith(skuName, 'E')"
	if options.filterManagedDisk != nil {
		filterForDisk = *options.filterManagedDisk
	}

	filter := fmt.Sprintf("armRegionName eq '%s' and serviceFamily eq 'Storage' and productName eq 'Standard SSD Managed Disks' and %s and endswith(meterName, 'LRS Disk') and Type eq 'Consumption' and unitOfMeasure eq '1/Month'", regionSku, filterForDisk)

	m.l.Debug(m.ctx, "Fetching Disk prices", "Filter", filter)

	prices, err := fetchPrices(filter)
	if err != nil {
		return nil, err
	}

	m.l.Debug(m.ctx, "Disk prices", "Prices", prices)

	o := make([]provider.StorageBlockRegionOutput, 0, len(prices))
	for _, p := range prices {
		o = append(
			o,
			provider.StorageBlockRegionOutput{
				Sku: utilities.Ptr(strings.Split(p.SkuName, " ")[0]),
				Price: &provider.PriceOutput{
					MonthlyPrice: utilities.Ptr(p.UnitPrice),
					Currency:     p.CurrencyCode,
				},
			},
		)
	}

	return o, nil
}

func WithDefaultInstanceType() Option {
	return func(op *options) error {
		op.filterInstanceType = utilities.Ptr("(startswith(armSkuName, 'Standard_D2s') or startswith(armSkuName, 'Standard_D4s') or startswith(armSkuName, 'Standard_E2s') or startswith(armSkuName, 'Standard_E4s') or startswith(armSkuName, 'Standard_F2s') or startswith(armSkuName, 'Standard_F4s') or startswith(armSkuName, 'Standard_B2s') or startswith(armSkuName, 'Standard_B4s'))")
		return nil
	}
}

func (m *AzureMeta) priceVMs(regionSku string, opts ...Option) (map[string]provider.InstanceRegionOutput, error) {

	var options options

	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	filterInstance := "startswith(armSkuName, 'Standard_')"
	if options.filterInstanceType != nil {
		filterInstance = *options.filterInstanceType
	}

	filter := fmt.Sprintf("serviceName eq 'Virtual Machines' and armRegionName eq '%s' and serviceFamily eq 'Compute' and %s and type eq 'Consumption' and unitOfMeasure eq '1 Hour'", regionSku, filterInstance)

	m.l.Debug(m.ctx, "Fetching VM prices", "Filter", filter)

	prices, err := fetchPrices(filter, IgnoreSpotAndLowPriMeterName())
	if err != nil {
		return nil, err
	}

	m.l.Debug(m.ctx, "VM prices", "Prices", prices)

	o := make(map[string]provider.InstanceRegionOutput)
	for _, p := range prices {
		o[p.ArmSkuName] = provider.InstanceRegionOutput{
			Price: provider.PriceOutput{
				HourlyPrice: utilities.Ptr(p.UnitPrice),
				Currency:    p.CurrencyCode,
			},
		}
	}

	return o, nil
}

func WithDefaultAks() Option {
	return func(op *options) error {
		op.filterAksOfferings = utilities.Ptr("meterName eq 'Standard Uptime SLA'")
		return nil
	}
}

func (m *AzureMeta) priceAksManagement(regionSku string, opts ...Option) (out []provider.ManagedClusterOutput, _ error) {

	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	filterAks := ""
	if options.filterAksOfferings != nil {
		filterAks = "and " + *options.filterAksOfferings
	}

	filter := fmt.Sprintf("serviceName eq 'Azure Kubernetes Service' and armRegionName eq '%s' and unitOfMeasure eq '1 Hour' and skuName eq 'Standard' %s", regionSku, filterAks)

	m.l.Debug(m.ctx, "Fetching AKS prices", "Filter", filter)

	prices, err := fetchPrices(filter)
	if err != nil {
		return nil, err
	}

	m.l.Debug(m.ctx, "AKS prices", "Prices", prices)

	// out = append(out, provider.ManagedClusterOutput{
	// 	Sku:         "Standard Free",
	// 	Description: "Aks Free Management",
	// 	Price: provider.PriceOutput{
	// 		HourlyPrice: utilities.Ptr(0.0),
	// 		Currency:    prices[0].CurrencyCode,
	// 	},
	// 	Tier: "Free",
	// })

	for _, p := range prices {
		tier := ""
		if p.MeterName == "Standard Long Term Support" {
			tier = "Premium"
		} else if p.MeterName == "Standard Uptime SLA" {
			tier = "Standard"
		}

		o := provider.ManagedClusterOutput{
			Sku:         p.MeterName,
			Description: p.MeterName + " " + p.ProductName,
			Tier:        tier,
			Price: provider.PriceOutput{
				HourlyPrice: utilities.Ptr(prices[0].UnitPrice),
				Currency:    p.CurrencyCode,
			},
		}

		out = append(out, o)
	}

	return out, nil
}
