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

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
)

type ComputeDetailsOutput struct {
	VCPU          string
	Memory        string
	OSDiskSize    string
	Region        string
	Price         float64
	Name          string
	MeterName     string
	UnitOfMeasure string
}
type VMSku string
type VMsOutput map[VMSku]ComputeDetailsOutput

type Price struct {
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

type PricingResponse struct {
	Items        []Price `json:"Items"`
	NextPageLink string  `json:"NextPageLink"`
}

func fetchPrices(query string) ([]Price, error) {
	apiURL := "https://prices.azure.com/api/retail/prices"
	v := make([]Price, 0, 100)

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

		var jsonData PricingResponse
		if err := json.NewDecoder(resp.Body).Decode(&jsonData); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return nil, err
		}
		for _, p := range jsonData.Items {
			if !strings.HasSuffix(p.MeterName, "Spot") &&
				!strings.HasSuffix(p.MeterName, "Low Priority") &&
				!strings.HasSuffix(p.ProductName, "Windows") {
				v = append(v, p)
			}
		}

		nextPage = jsonData.NextPageLink
	}

	return v, nil
}

type DiskDetailsOutput struct {
	Cost          float64
	DiskSizeGB    int
	ServiceName   string
	ProductName   string
	MeterName     string
	UnitOfMeasure string
	Region        string
}

type DiskSku string

type DisksOutput map[DiskSku]DiskDetailsOutput

func (p *ResourceDetails) Disks(ctx context.Context, region string) (DisksOutput, error) {
	ee := DisksOutput{}

	// NOTE: This query gives the pricing for All StandardSSDLRS
	// "armRegionName eq 'eastus' and serviceFamily eq 'Storage' and productName eq 'Standard SSD Managed Disks' and endswith(skuName, 'LRS') and startswith(skuName, 'E') and Type eq 'Consumption' and unitOfMeasure eq '1/Month'"

	// TODO: need to make the query more specific to get the pricing for the specific disk as we are unable to fetch the disk Sizes for all the StandardSSDLRS disks from the API

	// filter NOTE: This query gives the pricing for StandardSSDLRU with disk size 64 GB only which is E6 StandardSSDLRS managed Disk
	filter := fmt.Sprintf("armRegionName eq '%s' and serviceFamily eq 'Storage' and productName eq 'Standard SSD Managed Disks' and skuName eq 'E6 LRS' and Type eq 'Consumption' and unitOfMeasure eq '1/Month'", region)

	prices, err := fetchPrices(filter)
	if err != nil {
		return nil, err
	}

	for _, p := range prices {
		ee[DiskSku(p.SkuName)] = DiskDetailsOutput{
			Cost:          p.UnitPrice,
			ServiceName:   p.ServiceName,
			ProductName:   p.ProductName,
			UnitOfMeasure: p.UnitOfMeasure,
			MeterName:     p.MeterName,
			DiskSizeGB:    64,
			Region:        p.ArmRegionName,
		}
	}

	return ee, nil
}

func (p *ResourceDetails) VMs(ctx context.Context, region string) (VMsOutput, error) {
	ee := VMsOutput{}

	client, err := armcompute.NewVirtualMachineSizesClient(p.subscriptionId, p.cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	pager := client.NewListPager(region, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, vmSize := range page.Value {
			ee[VMSku(*vmSize.Name)] = ComputeDetailsOutput{
				VCPU:       strconv.FormatInt(int64(*vmSize.NumberOfCores), 10),
				Memory:     strconv.FormatInt(int64(*vmSize.MemoryInMB), 10) + " MB",
				OSDiskSize: strconv.FormatInt(int64(*vmSize.OSDiskSizeInMB), 10) + " MB",
			}
		}
	}

	filter := fmt.Sprintf("serviceName eq 'Virtual Machines' and armRegionName eq '%s' and serviceFamily eq 'Compute' and type eq 'Consumption'", region)
	prices, err := fetchPrices(filter)
	if err != nil {
		return nil, err
	}

	for _, p := range prices {
		if v, ok := ee[VMSku(p.ArmSkuName)]; ok {
			v.Price = p.UnitPrice
			v.Name = p.ProductName
			v.UnitOfMeasure = p.UnitOfMeasure
			v.MeterName = p.MeterName
			v.Region = p.ArmRegionName
			ee[VMSku(p.ArmSkuName)] = v
		}
	}

	return ee, nil
}

type AksDetailsOutput struct {
	Region        string
	Price         float64
	ServiceName   string
	ProductName   string
	UnitOfMeasure string

	Compute    ComputeDetailsOutput
	TotalPrice float64
}

type AKSTier string

const (
	Standard AKSTier = "Standard"
	Premium  AKSTier = "Premium"
	Free     AKSTier = "Free"
)

func (p *ResourceDetails) AKS(
	ctx context.Context,
	region string,
	vmType string,
	vmCount int,
	tier AKSTier,
) (*AksDetailsOutput, error) {
	ee := &AksDetailsOutput{}

	// By default the azure is free / standard it is user choice so $0 or $0.1/hr
	// For the NodeSize the user selects it
	// as the default AKS Disk is not managed but emphermal storage no need to consider the disk cost

	totalCost := 0.0
	if tier == Premium || tier == Standard {
		query := ""
		if tier == Premium {
			query = fmt.Sprintf("serviceName eq 'Azure Kubernetes Service' and armRegionName eq '%s' and meterName eq 'Standard Long Term Support'", region)
		} else {
			query = fmt.Sprintf("serviceName eq 'Azure Kubernetes Service' and armRegionName eq '%s' and meterName eq 'Standard Uptime SLA'", region)
		}

		prices, err := fetchPrices(query)
		if err != nil {
			return nil, err
		}

		if len(prices) == 0 {
			return nil, fmt.Errorf("no pricing found for AKS for region %s, tier %s", region, tier)
		}

		totalCost += prices[0].UnitPrice * 730
		ee.ServiceName = prices[0].ServiceName
		ee.ProductName = prices[0].ProductName
		ee.Region = prices[0].ArmRegionName
		ee.Price = prices[0].UnitPrice
		ee.UnitOfMeasure = prices[0].UnitOfMeasure
	}

	// Get the VMs
	vms, err := p.VMs(ctx, region)
	if err != nil {
		return nil, err
	}
	if v, ok := vms[VMSku(vmType)]; ok {
		ee.Compute = v
	}

	totalCost += ee.Compute.Price * float64(vmCount) * 730
	ee.TotalPrice = totalCost

	return ee, nil
}
