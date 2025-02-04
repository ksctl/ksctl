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
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingTypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

type AwsSDKPricing struct {
	Product struct {
		Sku        string         `json:"sku"`
		Attributes map[string]any `json:"attributes"`
	} `json:"product"`
	Terms struct {
		OnDemand map[string]struct {
			PriceDimensions map[string]struct {
				PricePerUnit struct {
					USD string `json:"USD"`
				} `json:"pricePerUnit"`
				Description       string `json:"description"`
				UnitOfMeasurement string `json:"unit"` // Hrs, GB-Mo, GB-month
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

func (m *AwsMeta) priceVMs(region provider.RegionOutput) (map[string]provider.InstanceRegionOutput, error) {
	// aws pricing get-products --service-code AmazonEC2 --filters Type=TERM_MATCH,Field=operatingSystem,Value=linux Type=TERM_MATCH,Field=tenancy,Value="Shared" Type=TERM_MATCH,Field=capacitystatus,Value="Used" Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" --output=json
	filters := []pricingTypes.Filter{
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("serviceCode"),
			Value: aws.String("AmazonEC2"),
		},
		//{
		//	Type:  pricingTypes.FilterTypeTermMatch,
		//	Field: aws.String("instanceType"),
		//	Value: aws.String(vmType),
		//},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("location"),
			Value: aws.String(region.Name),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("operatingSystem"),
			Value: aws.String("Linux"),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("preInstalledSw"),
			Value: aws.String("NA"),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("tenancy"),
			Value: aws.String("Shared"),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("capacitystatus"),
			Value: aws.String("Used"),
		},
	}

	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEC2"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
	}

	session, err := m.GetNewSession(region.Sku)
	if err != nil {
		return nil, err
	}

	priC := pricing.NewFromConfig(*session)
	var vmsPrice pricing.GetProductsOutput

	for {
		o, err := priC.GetProducts(m.ctx, input)
		if err != nil {
			// TODO: need to correctly handle the error
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "Error in fetching instance types", "Reason", err),
			)
		}
		vmsPrice.PriceList = append(vmsPrice.PriceList, o.PriceList...)
		if o.NextToken == nil {
			break
		}
		input.NextToken = o.NextToken
	}

	out := make(map[string]provider.InstanceRegionOutput, len(vmsPrice.PriceList))
	for _, p := range vmsPrice.PriceList {
		pricingResult := AwsSDKPricing{}
		err = json.Unmarshal([]byte(p), &pricingResult)
		if err != nil {
			return nil, err
		}

		instanceType := ""

		if o, ok := pricingResult.Product.Attributes["instanceType"]; ok {
			instanceType = o.(string)
		}

		//if o, ok := pricingResult.Product.Attributes["tenancy"]; ok {
		//	ee.Tenancy = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["currentGeneration"]; ok {
		//	ee.CurrentGeneration = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["physicalProcessor"]; ok {
		//	ee.PhysicalProcessor = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["processorFeatures"]; ok {
		//	ee.ProcessorFeatures = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["processorArchitecture"]; ok {
		//	ee.ProcessorArchitecture = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["vcpu"]; ok {
		//	ee.VCPU = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["memory"]; ok {
		//	ee.Memory = o.(string)
		//}
		//if o, ok := pricingResult.Product.Attributes["clockSpeed"]; ok {
		//	ee.ClockSpeed = o.(string)
		//}

		for _, onDemand := range pricingResult.Terms.OnDemand {
			for _, priceDimension := range onDemand.PriceDimensions {
				var monthlyPrice, hourlyPrice *float64

				cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
				if err != nil {
					log.Fatalf("Failed to parse hourly cost: %v", err)
				}
				if priceDimension.UnitOfMeasurement == "Hrs" {
					hourlyPrice = utilities.Ptr(cost)
				} else if priceDimension.UnitOfMeasurement == "GB-Mo" || priceDimension.UnitOfMeasurement == "GB-month" {
					monthlyPrice = utilities.Ptr(cost)
				}

				out[instanceType] = provider.InstanceRegionOutput{
					Price: provider.PriceOutput{
						HourlyPrice:  hourlyPrice,
						MonthlyPrice: monthlyPrice,
						Currency:     "USD",
					},
				}
				break
			}
			break
		}
	}

	return out, nil
}

func (m *AwsMeta) priceDisks(region provider.RegionOutput) (map[string]provider.StorageBlockRegionOutput, error) {
	volTypes := []string{"gp3", "gp2", "io1", "io2"}
	out := make(map[string]provider.StorageBlockRegionOutput, len(volTypes))

	for _, volumeType := range volTypes {
		v, err := m.priceSpecificEBS(region, volumeType, 30)
		if err != nil {
			return nil, err
		}
		out[volumeType] = *v
	}

	return out, nil
}

func (m *AwsMeta) priceSpecificEBS(region provider.RegionOutput, volumeType string, volSize int) (*provider.StorageBlockRegionOutput, error) {

	filterVolType := ""
	var minIops, maxIops, minThroughput, maxThroughput *int32

	switch volumeType {
	case "gp3":
		minIops = utilities.Ptr(int32(3000))
		minThroughput = utilities.Ptr(int32(125))
		filterVolType = "General Purpose"
	case "gp2":
		filterVolType = "General Purpose"
	case "io1", "io2":
		filterVolType = "Provisioned IOPS"
	}

	// Get volume pricing
	// aws pricing get-products --service-code AmazonEC2 --filters Type=TERM_MATCH,Field=productFamily,Value=Storage Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" Type=TERM_MATCH,Field=volumeType,Value="Provisioned IOPS" Type=TERM_MATCH,Field=volumeType,Value="General Purpose" --output=json
	filters := []pricingTypes.Filter{
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("serviceCode"),
			Value: aws.String("AmazonEC2"),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("volumeType"),
			Value: aws.String(filterVolType),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("volumeApiName"),
			Value: aws.String(volumeType),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("location"),
			Value: aws.String(region.Name),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("productFamily"),
			Value: aws.String("Storage"),
		},
	}

	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEC2"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int32(1),
	}
	session, err := m.GetNewSession(region.Sku)
	if err != nil {
		return nil, err
	}

	priC := pricing.NewFromConfig(*session)

	result, err := priC.GetProducts(m.ctx, input)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed to get EBS pricing", "reason", err),
		)
	}

	var out *provider.StorageBlockRegionOutput
	if result != nil && len(result.PriceList) > 0 {

		pricingResult := AwsSDKPricing{}
		err = json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to unmarshal EBS pricing", "reason", err),
			)
		}

		if o, ok := pricingResult.Product.Attributes["maxThroughputvolume"]; ok {
			_o := o.(string)
			_o = strings.TrimSpace(strings.TrimSuffix(_o, "MiB/s"))
			v, err := strconv.Atoi(_o)
			if err != nil {
				return nil, ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					m.l.NewError(m.ctx, "failed to get max throughput", "reason", err),
				)
			}
			maxThroughput = utilities.Ptr(int32(v))
		}
		if o, ok := pricingResult.Product.Attributes["maxIopsvolume"]; ok {
			v, err := strconv.Atoi(o.(string))
			if err != nil {
				return nil, ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					m.l.NewError(m.ctx, "failed to get max Iops", "reason", err),
				)
			}
			maxIops = utilities.Ptr(int32(v))
		}
		//GB-month,GB-Mo
		for _, onDemand := range pricingResult.Terms.OnDemand {
			for _, priceDimension := range onDemand.PriceDimensions {
				var monthlyPrice, hourlyPrice *float64

				cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
				if err != nil {
					log.Fatalf("Failed to parse monthly cost: %v", err)
				}
				if priceDimension.UnitOfMeasurement == "Hrs" {
					hourlyPrice = utilities.Ptr(cost)
				} else if priceDimension.UnitOfMeasurement == "GB-Mo" || priceDimension.UnitOfMeasurement == "GB-month" {
					monthlyPrice = utilities.Ptr(cost)
				}
				out = &provider.StorageBlockRegionOutput{
					Sku: utilities.Ptr(volumeType),
					Price: &provider.PriceOutput{
						MonthlyPrice: monthlyPrice,
						HourlyPrice:  hourlyPrice,
						Currency:     "USD",
					},
					Tier:           utilities.Ptr(filterVolType),
					Size:           utilities.Ptr(int32(volSize)),
					AttachmentType: provider.InstanceDiskNotIncluded,
					MaxIOps:        maxIops,
					MinIOps:        minIops,
					MaxThroughput:  maxThroughput,
					MinThroughput:  minThroughput,
				}

				break
			}
			break
		}

	}
	return out, nil
}

//type EksType string
//
//const (
//	EksTypeStandard EksType = "Standard"
//	EksTypeExtended EksType = "Extended"
//)
//
//type EksDetailsOutput struct {
//	RegionDescription string
//	VMType            string
//	EksType           string
//	MonthlyCost       float64
//	PriceDescription  string
//	EC2DetailsOutput  *EC2DetailsOutput
//	EBSDetailsOutput  *EBSDetailsOutput
//	VMCount           int
//	EnabledAutoUsage  bool
//}
//
//func (p *ResourceDetails) EKS(
//	region string,
//	vmType string,
//	vmCount int,
//	enabledEksAutoUsage bool,
//	eksType EksType) (*EksDetailsOutput, error) {
//
//	//So the total cost would be the sum of:
//	//EKS Auto Mode management fee
//	//EKS cluster control plane cost
//	//EC2 instance costs for nodes
//	//EBS volume costs for persistent storage
//
//	regionDesc := string(p.regions[RegionCode(region)])
//
//	ee := &EksDetailsOutput{
//		RegionDescription: regionDesc,
//		VMType:            vmType,
//		EksType:           string(eksType),
//		VMCount:           vmCount,
//		EnabledAutoUsage:  enabledEksAutoUsage,
//	}
//
//	filters := []types.Filter{
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("serviceCode"),
//			Value: aws.String("AmazonEKS"),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("location"),
//			Value: aws.String(regionDesc),
//		},
//	}
//
//	if enabledEksAutoUsage {
//		autoMode := append(filters, []types.Filter{
//			{
//				Type:  types.FilterTypeTermMatch,
//				Field: aws.String("instancetype"),
//				Value: aws.String(vmType),
//			},
//			{
//				Type:  types.FilterTypeTermMatch,
//				Field: aws.String("operation"),
//				Value: aws.String("EKSAutoUsage"),
//			}}...,
//		)
//
//		cost, desc, err := p.pricingForEKS(autoMode)
//		if err != nil {
//			return nil, err
//		}
//		ee.MonthlyCost = cost
//		ee.PriceDescription = desc
//	}
//
//	switch eksType {
//	case EksTypeStandard:
//		filters = append(filters, []types.Filter{
//			{
//				Type:  types.FilterTypeTermMatch,
//				Field: aws.String("operation"),
//				Value: aws.String("CreateOperation"),
//			},
//		}...,
//		)
//
//	case EksTypeExtended:
//		filters = append(filters, []types.Filter{
//			{
//				Type:  types.FilterTypeTermMatch,
//				Field: aws.String("operation"),
//				Value: aws.String("ExtendedSupport"),
//			},
//		}...,
//		)
//	}
//
//	cost, desc, err := p.pricingForEKS(filters)
//	if err != nil {
//		return nil, err
//	}
//	ee.MonthlyCost += cost
//	if len(ee.PriceDescription) != 0 {
//		ee.PriceDescription = ee.PriceDescription + "\n" + desc
//	} else {
//		ee.PriceDescription = desc
//	}
//
//	costEc2, err := p.EC2(region, vmType)
//	if err != nil {
//		return nil, err
//	}
//	ee.EC2DetailsOutput = costEc2
//	ee.PriceDescription = ee.PriceDescription + "\n" + costEc2.PriceDescription
//
//	costEbs, err := p.EBS(EBSGp3, 30, region)
//	if err != nil {
//		return nil, err
//	}
//	ee.EBSDetailsOutput = costEbs
//	ee.PriceDescription = ee.PriceDescription + "\n" + costEbs.PriceDescription
//
//	ee.MonthlyCost += (costEc2.HourlyCost) * float64(vmCount)
//
//	ee.MonthlyCost *= 730
//
//	ee.MonthlyCost += costEbs.TotalMonthlyCost * float64(vmCount)
//
//	return ee, nil
//}
//
//func (p *ResourceDetails) pricingForEKS(filters []types.Filter) (hourlyCost float64, desc string, err error) {
//
//	input := &pricing.GetProductsInput{
//		ServiceCode:   aws.String("AmazonEKS"),
//		Filters:       filters,
//		FormatVersion: aws.String("aws_v1"),
//		MaxResults:    aws.Int32(1),
//	}
//
//	result, err := p.svc.GetProducts(context.TODO(), input)
//	if err != nil {
//		return 0.0, "", err
//	}
//
//	if result != nil && len(result.PriceList) > 0 {
//		pricingResult := PricingResult{}
//		err = json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
//		if err != nil {
//			return 0.0, "", err
//		}
//
//		for _, onDemand := range pricingResult.Terms.OnDemand {
//			for _, priceDimension := range onDemand.PriceDimensions {
//				desc = priceDimension.Description
//				hourlyCost, err = strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
//				if err != nil {
//					return 0.0, "", fmt.Errorf("failed to parse hourly cost: %v", err)
//				}
//				return hourlyCost, desc, nil
//			}
//			break
//		}
//	}
//
//	err = fmt.Errorf("failed to get EKS pricing, no results")
//	return
//}
