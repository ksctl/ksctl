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
				Description string `json:"description"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

//type EC2DetailsOutput struct {
//	RegionDescription     string
//	VMType                string
//	HourlyCost            float64
//	PriceDescription      string
//	Tenancy               string
//	CurrentGeneration     string
//	PhysicalProcessor     string
//	ProcessorFeatures     string
//	ProcessorArchitecture string
//	VCPU                  string
//	Memory                string
//	ClockSpeed            string
//}
//
//type EBSVolumeType string
//
//const (
//	EBSGp3 EBSVolumeType = "gp3"
//	EBSGp2 EBSVolumeType = "gp2"
//	EBSIo1 EBSVolumeType = "io1"
//	EBSIo2 EBSVolumeType = "io2"
//)
//
//type EBSDetailsOutput struct {
//	RegionDescription string
//	VolumeType        string
//	VolumeSizeGB      int
//
//	AdditionalCostForThroughPutOrIOPS bool
//	PricePerGBMonth                   float64
//
//	MaxThroughputVol string // For gp3 (MB/s)
//	MaxIopsvolume    int    // For gp3
//
//	BaseIOPS      int // Base IOPS included for Gp3
//	BaseThoughput int // Base throughput included (MB/s) for GP3
//
//	TotalMonthlyCost float64
//	PriceDescription string
//}
//
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
//func (p *ResourceDetails) EC2(region, vmType string) (*EC2DetailsOutput, error) {
//
//	regionDescription := string(p.regions[RegionCode(region)])
//	ee := &EC2DetailsOutput{
//		RegionDescription: regionDescription,
//		VMType:            vmType,
//	}
//
//	filters := []types.Filter{
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("serviceCode"),
//			Value: aws.String("AmazonEC2"),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("instanceType"),
//			Value: aws.String(vmType),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("location"),
//			Value: aws.String(regionDescription),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("operatingSystem"),
//			Value: aws.String("Linux"),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("preInstalledSw"),
//			Value: aws.String("NA"),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("tenancy"),
//			Value: aws.String("Shared"),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("capacitystatus"),
//			Value: aws.String("Used"),
//		},
//	}
//
//	input := &pricing.GetProductsInput{
//		ServiceCode:   aws.String("AmazonEC2"),
//		Filters:       filters,
//		FormatVersion: aws.String("aws_v1"),
//		MaxResults:    aws.Int32(1),
//	}
//
//	result, err := p.svc.GetProducts(context.TODO(), input)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get products, %v", err)
//	}
//
//	if result != nil && len(result.PriceList) > 0 {
//
//		pricingResult := PricingResult{}
//		err = json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
//		if err != nil {
//			return nil, err
//		}
//
//		// TODO: need to add architecture of the machine type
//
//		if o, ok := pricingResult.Product.Attributes["tenancy"]; ok {
//			ee.Tenancy = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["currentGeneration"]; ok {
//			ee.CurrentGeneration = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["physicalProcessor"]; ok {
//			ee.PhysicalProcessor = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["processorFeatures"]; ok {
//			ee.ProcessorFeatures = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["processorArchitecture"]; ok {
//			ee.ProcessorArchitecture = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["vcpu"]; ok {
//			ee.VCPU = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["memory"]; ok {
//			ee.Memory = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["clockSpeed"]; ok {
//			ee.ClockSpeed = o.(string)
//		}
//
//		for _, onDemand := range pricingResult.Terms.OnDemand {
//			for _, priceDimension := range onDemand.PriceDimensions {
//				ee.PriceDescription = priceDimension.Description
//				ee.HourlyCost, err = strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
//				if err != nil {
//					log.Fatalf("Failed to parse hourly cost: %v", err)
//				}
//				break
//			}
//			break
//		}
//	}
//	return ee, nil
//}
//
//func (p *ResourceDetails) EBS(volumeType EBSVolumeType, sizeGB int, region string) (*EBSDetailsOutput, error) {
//
//	regionDescription := string(p.regions[RegionCode(region)])
//	ee := &EBSDetailsOutput{
//		RegionDescription: regionDescription,
//		VolumeType:        string(volumeType),
//		VolumeSizeGB:      sizeGB,
//	}
//
//	filterVolType := ""
//
//	switch volumeType {
//	case EBSGp3:
//		ee.BaseIOPS = 3000
//		ee.BaseThoughput = 125
//		ee.AdditionalCostForThroughPutOrIOPS = true
//		filterVolType = "General Purpose"
//	case EBSGp2:
//		ee.BaseIOPS = 0
//		ee.BaseThoughput = 0
//		ee.AdditionalCostForThroughPutOrIOPS = false
//		filterVolType = "General Purpose"
//	case EBSIo1:
//		ee.BaseIOPS = 0
//		ee.BaseThoughput = 0
//		ee.AdditionalCostForThroughPutOrIOPS = true
//		filterVolType = "Provisioned IOPS"
//	case EBSIo2:
//		ee.BaseIOPS = 0
//		ee.BaseThoughput = 0
//		ee.AdditionalCostForThroughPutOrIOPS = true
//		filterVolType = "Provisioned IOPS"
//	}
//
//	// Get volume pricing
//	// aws pricing get-products --service-code AmazonEC2 --filters Type=TERM_MATCH,Field=productFamily,Value=Storage Type=TERM_MATCH,Field=location,Value="US East (N. Virginia)" Type=TERM_MATCH,Field=volumeType,Value="Provisioned IOPS" Type=TERM_MATCH,Field=volumeType,Value="General Purpose" --output=json
//	filters := []types.Filter{
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("serviceCode"),
//			Value: aws.String("AmazonEC2"),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("volumeType"),
//			Value: aws.String(filterVolType),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("volumeApiName"),
//			Value: aws.String(string(volumeType)),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("location"),
//			Value: aws.String(regionDescription),
//		},
//		{
//			Type:  types.FilterTypeTermMatch,
//			Field: aws.String("productFamily"),
//			Value: aws.String("Storage"),
//		},
//	}
//
//	input := &pricing.GetProductsInput{
//		ServiceCode:   aws.String("AmazonEC2"),
//		Filters:       filters,
//		FormatVersion: aws.String("aws_v1"),
//		MaxResults:    aws.Int32(1),
//	}
//
//	result, err := p.svc.GetProducts(context.TODO(), input)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get EBS pricing: %v", err)
//	}
//
//	monthPerGB := 0.0
//	if result != nil && len(result.PriceList) > 0 {
//
//		pricingResult := PricingResult{}
//		err = json.Unmarshal([]byte(result.PriceList[0]), &pricingResult)
//		if err != nil {
//			return nil, err
//		}
//
//		if o, ok := pricingResult.Product.Attributes["maxThroughputvolume"]; ok {
//			ee.MaxThroughputVol = o.(string)
//		}
//		if o, ok := pricingResult.Product.Attributes["maxIopsvolume"]; ok {
//			ee.MaxIopsvolume, err = strconv.Atoi(o.(string))
//			if err != nil {
//				ee.MaxIopsvolume = -1 // failed to get it
//			}
//		}
//
//		for _, onDemand := range pricingResult.Terms.OnDemand {
//			for _, priceDimension := range onDemand.PriceDimensions {
//				ee.PriceDescription = priceDimension.Description
//				monthPerGB, err = strconv.ParseFloat(priceDimension.PricePerUnit.USD, 64)
//				if err != nil {
//					log.Fatalf("Failed to parse monthly cost: %v", err)
//				}
//				ee.PricePerGBMonth = monthPerGB
//				break
//			}
//			break
//		}
//
//		ee.TotalMonthlyCost = monthPerGB * float64(sizeGB)
//	}
//
//	return ee, nil
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
