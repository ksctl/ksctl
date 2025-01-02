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

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
)

func PricingForEC2(cfg aws.Config, regionDescription, vmType string) error {

	svc := pricing.NewFromConfig(cfg)

	filters := []types.Filter{
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("serviceCode"),
			Value: aws.String("AmazonEC2"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("instanceType"),
			Value: aws.String(vmType),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("location"),
			Value: aws.String(regionDescription),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("operatingSystem"),
			Value: aws.String("Linux"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("preInstalledSw"),
			Value: aws.String("NA"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("tenancy"),
			Value: aws.String("Shared"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("capacitystatus"),
			Value: aws.String("Used"),
		},
	}

	// Retrieve products
	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEC2"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
	}

	result, err := svc.GetProducts(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to get products, %v", err)
	}

	// Parse and display the pricing information
	for _, priceItem := range result.PriceList {
		var product map[string]interface{}
		if err := json.Unmarshal([]byte(priceItem), &product); err != nil {
			return fmt.Errorf("failed to unmarshal price item, %v", err)
		}

		// Navigate through the JSON structure to find the price
		terms, ok := product["terms"].(map[string]interface{})
		if !ok {
			continue
		}
		onDemand, ok := terms["OnDemand"].(map[string]interface{})
		if !ok {
			continue
		}
		for _, term := range onDemand {
			termAttributes, ok := term.(map[string]interface{})
			if !ok {
				continue
			}
			priceDimensions, ok := termAttributes["priceDimensions"].(map[string]interface{})
			if !ok {
				continue
			}
			for _, dimension := range priceDimensions {
				dimensionAttributes, ok := dimension.(map[string]interface{})
				if !ok {
					continue
				}
				pricePerUnit, ok := dimensionAttributes["pricePerUnit"].(map[string]interface{})
				if !ok {
					continue
				}
				usdPrice, ok := pricePerUnit["USD"].(string)
				if !ok {
					continue
				}
				fmt.Printf("Instance Type: %s in Region: %s, Price per Hour (USD): %s\n", vmType, regionDescription, usdPrice)
			}
		}
	}

	return nil
}
