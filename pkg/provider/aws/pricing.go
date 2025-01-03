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
	"github.com/gookit/goutil/dump"
	"log"
	"strconv"
)

type PricingResult struct {
	Product struct {
		Sku string `json:"sku"`
	} `json:"product"`
	Terms struct {
		OnDemand map[string]struct {
			PriceDimensions map[string]struct {
				PricePerUnit struct {
					USD string `json:"USD"`
				} `json:"pricePerUnit"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

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

	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEC2"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int32(1),
	}

	result, err := svc.GetProducts(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to get products, %v", err)
	}

	hourlyCost := 0.0
	if result != nil && len(result.PriceList) > 0 {

		pricingResult := PricingResult{}
		err := json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		for _, onDemand := range pricingResult.Terms.OnDemand {
			for _, priceDimension := range onDemand.PriceDimensions {
				hourlyCost, err = strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
				if err != nil {
					log.Fatalf("Failed to parse hourly cost: %v", err)
				}
				break
			}
			break
		}
		fmt.Printf("Instance Type: %s in Region: %s, Price per Hour (USD): %f\n", vmType, regionDescription, hourlyCost)
	}
	return nil
}

func PricingForEKS(cfg aws.Config, regionDescription string, vmType string) error {

	svc := pricing.NewFromConfig(cfg)

	filters := []types.Filter{
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("serviceCode"),
			Value: aws.String("AmazonEKS"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("location"),
			Value: aws.String(regionDescription),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("instancetype"),
			Value: aws.String(vmType),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("operation"),
			Value: aws.String("EKSAutoUsage"), // Other options: EKSAutoUsage,CreateOperation,HybridNodesUsage,ExtendedSupport
		},
	}

	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEKS"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
		//MaxResults:    aws.Int32(1),
	}

	//result, err := svc.GetProducts(context.TODO(), input)
	//if err != nil {
	//	return fmt.Errorf("failed to get products, %v", err)
	//}
	//dump.NewWithOptions(dump.SkipPrivate(), dump.SkipNilField()).Println(result)
	//
	//hourlyCost := 0.0
	//if result != nil && len(result.PriceList) > 0 {
	//
	//	pricingResult := PricingResult{}
	//	err := json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
	//	if err != nil {
	//		log.Fatalf("Failed to unmarshal JSON: %v", err)
	//	}
	//
	//	for _, onDemand := range pricingResult.Terms.OnDemand {
	//		for _, priceDimension := range onDemand.PriceDimensions {
	//			hourlyCost, err = strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
	//			if err != nil {
	//				log.Fatalf("Failed to parse hourly cost: %v", err)
	//			}
	//			break
	//		}
	//		break
	//	}
	//	fmt.Printf("EKS in Region: %s, Price per Hour (USD): %f\n", regionDescription, hourlyCost)
	//}
	//

	paginator := pricing.NewGetProductsPaginator(svc, input)

	for paginator.HasMorePages() {
		result, err := paginator.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to get products, %v", err)
		}

		for _, priceList := range result.PriceList {
			v := map[string]any{}
			_ = json.Unmarshal([]byte(priceList), &v)
			dump.NewWithOptions(dump.SkipPrivate(), dump.SkipNilField()).Println(v)

			pricingResult := PricingResult{}
			err := json.Unmarshal([]byte(priceList), &pricingResult)
			if err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for _, onDemand := range pricingResult.Terms.OnDemand {
				for _, priceDimension := range onDemand.PriceDimensions {
					hourlyCost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
					if err != nil {
						log.Fatalf("Failed to parse hourly cost: %v", err)
					}
					fmt.Printf("EKS in Region: %s, Price per Hour (USD): %f\n", regionDescription, hourlyCost)
				}
			}
		}
	}

	return nil
}
