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
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
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

type options struct {
	ebsType *string
	ebsSize *int32
	eksType *string
	ec2Type []string
}

type Option func(*options) error

func (m *AwsMeta) priceVMs(region provider.RegionOutput) (map[string]provider.InstanceRegionOutput, error) {
	// aws pricing get-products --service-code AmazonEC2 --filters Type=TERM_MATCH,Field=operatingSystem,Value=linux Type=TERM_MATCH,Field=tenancy,Value="Shared" Type=TERM_MATCH,Field=capacitystatus,Value="Used" Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" --output=json
	filters := []pricingTypes.Filter{
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("serviceCode"),
			Value: aws.String("AmazonEC2"),
		},
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

		for _, onDemand := range pricingResult.Terms.OnDemand {
			for _, priceDimension := range onDemand.PriceDimensions {
				var monthlyPrice, hourlyPrice *float64

				cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
				if err != nil {
					log.Fatalf("Failed to parse hourly cost: %v", err)
				}
				switch priceDimension.UnitOfMeasurement {
				case "Hrs", "hours", "Hours":
					hourlyPrice = utilities.Ptr(cost)
				case "GB-Mo", "GB-month":
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

func WithDefaultEBSVolumeType() Option {
	return func(o *options) error {
		o.ebsType = utilities.Ptr("gp3")
		return nil
	}
}

func WithDefaultEBSSize() Option {
	return func(o *options) error {
		o.ebsSize = utilities.Ptr(int32(30))
		return nil
	}
}

func (m *AwsMeta) priceDisks(region provider.RegionOutput, opts ...Option) ([]provider.StorageBlockRegionOutput, error) {
	options := options{}
	for _, o := range opts {
		if err := o(&options); err != nil {
			return nil, err
		}
	}
	if options.ebsSize == nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			m.l.NewError(m.ctx, "ebsSize is required", "reason", "Please pass on the selected ebsSize"),
		)
	}

	volTypes := []string{"gp3", "gp2", "io1", "io2"}
	out := make([]provider.StorageBlockRegionOutput, 0, len(volTypes))

	if options.ebsType != nil {
		v, err := m.priceSpecificEBS(region, *options.ebsType, *options.ebsSize)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	} else {
		for _, volumeType := range volTypes {
			v, err := m.priceSpecificEBS(region, volumeType, *options.ebsSize)
			if err != nil {
				return nil, err
			}
			out = append(out, *v)
		}
	}

	return out, nil
}

func (m *AwsMeta) priceSpecificEBS(region provider.RegionOutput, volumeType string, volSize int32) (*provider.StorageBlockRegionOutput, error) {

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
		for _, onDemand := range pricingResult.Terms.OnDemand {
			for _, priceDimension := range onDemand.PriceDimensions {
				var monthlyPrice, hourlyPrice *float64

				cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
				if err != nil {
					log.Fatalf("Failed to parse monthly cost: %v", err)
				}

				switch priceDimension.UnitOfMeasurement {
				case "Hrs", "hours", "Hours":
					hourlyPrice = utilities.Ptr(cost)
				case "GB-Mo", "GB-month":
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
					Size:           utilities.Ptr(volSize),
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

func WithDefaultEKS() Option {
	return func(o *options) error {
		o.eksType = utilities.Ptr("Standard")
		return nil
	}
}

func (m *AwsMeta) priceEksManagement(region provider.RegionOutput, vmType *string, opts ...Option) ([]provider.ManagedClusterOutput, error) {

	options := options{}
	for _, o := range opts {
		if err := o(&options); err != nil {
			return nil, err
		}
	}

	eksTypes := []string{"Standard", "Extended", "AutoNode Standard", "AutoNode Extended"}
	out := make([]provider.ManagedClusterOutput, 0, len(eksTypes))

	if options.eksType != nil {
		v, err := m.priceSpeficEks(region, *options.eksType, vmType)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	} else {
		for _, eksType := range eksTypes {
			if eksType == "AutoNode Standard" || eksType == "AutoNode Extended" {
				if vmType == nil {
					return nil, ksctlErrors.WrapError(
						ksctlErrors.ErrInvalidUserInput,
						m.l.NewError(m.ctx, "vmType is required for AutoNode EKS calculation", "reason", "Please pass on the selected vmType SKU"),
					)
				}
			}
			v, err := m.priceSpeficEks(region, eksType, vmType)
			if err != nil {
				return nil, err
			}
			out = append(out, *v)
		}
	}

	return out, nil
}

func (m *AwsMeta) priceSpeficEks(region provider.RegionOutput, eksType string, vmType *string) (*provider.ManagedClusterOutput, error) {

	filters := []pricingTypes.Filter{
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("serviceCode"),
			Value: aws.String("AmazonEKS"),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("location"),
			Value: aws.String(region.Name),
		},
		{
			Type:  pricingTypes.FilterTypeTermMatch,
			Field: aws.String("locationType"),
			Value: aws.String("AWS Region"),
		},
	}
	var autoNode *provider.ManagedClusterOutput
	if strings.HasPrefix(eksType, "AutoNode") {
		autonodeFilters := append(filters, []types.Filter{
			{
				Type:  pricingTypes.FilterTypeTermMatch,
				Field: aws.String("instancetype"),
				Value: aws.String(*vmType),
			},
			{
				Type:  pricingTypes.FilterTypeTermMatch,
				Field: aws.String("operation"),
				Value: aws.String("EKSAutoUsage"),
			}}...,
		)

		if v, err := m._pricingForEKS(true, eksType, autonodeFilters, region); err != nil {
			return nil, err
		} else {
			autoNode = v
		}
	}

	if strings.HasSuffix(eksType, "Standard") {
		filters = append(filters, []types.Filter{
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("operation"),
				Value: aws.String("CreateOperation"),
			},
		}...,
		)
	}
	if strings.HasSuffix(eksType, "Extended") {
		filters = append(filters, []types.Filter{
			{
				Type:  types.FilterTypeTermMatch,
				Field: aws.String("operation"),
				Value: aws.String("ExtendedSupport"),
			},
		}...,
		)
	}

	var normalOffering *provider.ManagedClusterOutput
	if v, err := m._pricingForEKS(false, eksType, filters, region); err != nil {
		return nil, err
	} else {
		normalOffering = v
	}

	if autoNode == nil {
		// no autonode
		return normalOffering, nil
	}

	// with autonode
	hourlyRate, monthlyRate := 0.0, 0.0
	if normalOffering.Price.HourlyPrice != nil {
		hourlyRate = *normalOffering.Price.HourlyPrice
	}
	if normalOffering.Price.MonthlyPrice != nil {
		monthlyRate = *normalOffering.Price.MonthlyPrice
	}
	if autoNode.Price.HourlyPrice != nil {
		hourlyRate += *autoNode.Price.HourlyPrice
	}
	if autoNode.Price.MonthlyPrice != nil {
		monthlyRate += *autoNode.Price.MonthlyPrice
	}

	return &provider.ManagedClusterOutput{
		Sku:         normalOffering.Sku + " " + autoNode.Sku,
		Description: autoNode.Description,
		Tier:        normalOffering.Tier,
		Price: provider.PriceOutput{
			HourlyPrice:  utilities.Ptr(hourlyRate),
			MonthlyPrice: utilities.Ptr(monthlyRate),
			Currency:     "USD",
		},
	}, nil
}

// Standard $ aws pricing get-products --service-code AmazonEKS --filters Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" Type=TERM_MATCH,Field=operation,Value="CreateOperation" Type=TERM_MATCH,Field=locationType,Value="AWS Region" --output=json
// Extended $ aws pricing get-products --service-code AmazonEKS --filters Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" Type=TERM_MATCH,Field=operation,Value="ExtendedSupport" Type=TERM_MATCH,Field=locationType,Value="AWS Region" --output=json
// AutoNode $ aws pricing get-products --service-code AmazonEKS --filters Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" Type=TERM_MATCH,Field=operation,Value="EKSAutoUsage" Type=TERM_MATCH,Field=locationType,Value="AWS Region" Type=TERM_MATCH,Field=instancetype,Value="t2.micro" --output=json  ~~~> if the result is empty then the instancetype isn't compatable

func (m *AwsMeta) _pricingForEKS(
	isAutoNode bool, eksType string,
	filters []pricingTypes.Filter,
	region provider.RegionOutput) (*provider.ManagedClusterOutput, error) {

	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEKS"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int32(1),
	}

	session, err := m.GetNewSession(region.Sku)
	if err != nil {
		return nil, err
	}

	result, err := pricing.NewFromConfig(*session).GetProducts(m.ctx, input)
	if err != nil {
		return nil, err
	}

	var out *provider.ManagedClusterOutput
	if result != nil && len(result.PriceList) > 0 {
		pricingResult := AwsSDKPricing{}
		err = json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to unmarshal EKS pricing", "reason", err),
			)
		}

		tierType := ""

		if v, ok := pricingResult.Product.Attributes["tiertype"]; !ok && !isAutoNode {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to get tier type", "reason", err),
			)
		} else {
			tierType = v.(string)
		}

		sku := ""

		if v, ok := pricingResult.Product.Attributes["operation"]; !ok {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to get operation", "reason", err),
			)
		} else {
			sku = v.(string)
		}

		for _, onDemand := range pricingResult.Terms.OnDemand {
			for _, priceDimension := range onDemand.PriceDimensions {
				var monthlyPrice, hourlyPrice *float64

				cost, err := strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
				if err != nil {
					log.Fatalf("Failed to parse monthly cost: %v", err)
				}

				switch priceDimension.UnitOfMeasurement {
				case "Hrs", "hours", "Hours":
					hourlyPrice = utilities.Ptr(cost)
				case "GB-Mo", "GB-month":
					monthlyPrice = utilities.Ptr(cost)
				}

				out = &provider.ManagedClusterOutput{
					Price: provider.PriceOutput{
						HourlyPrice:  hourlyPrice,
						MonthlyPrice: monthlyPrice,
						Currency:     "USD",
					},
					Tier:        tierType,
					Description: eksType,
					Sku:         sku,
				}
				break
			}
			break
		}
	}

	return out, nil
}
