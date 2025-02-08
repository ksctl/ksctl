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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	_ "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

type AwsMeta struct {
	ctx   context.Context
	l     logger.Logger
	creds statefile.CredentialsAws

	// cachedInstanceMapping used for calling pricing api
	cachedRegionMapping []provider.RegionOutput
}

func NewAwsMeta(ctx context.Context, l logger.Logger) (*AwsMeta, error) {
	v, ok := config.IsContextPresent(ctx, consts.KsctlAwsCredentials)
	if !ok {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			l.NewError(ctx, "missing aws credentials"),
		)
	}
	extractedCreds := statefile.CredentialsAws{}
	if err := json.Unmarshal([]byte(v), &extractedCreds); err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			l.NewError(ctx, "failed to get aws credentials", "reason", err),
		)
	}

	return &AwsMeta{
		ctx:   ctx,
		l:     l,
		creds: extractedCreds,
	}, nil
}

// GetNewSession returns a new aws session given you can "" or any valid region
func (m *AwsMeta) GetNewSession(regionSku string) (*aws.Config, error) {
	_session, err := awsConfig.LoadDefaultConfig(m.ctx,
		awsConfig.WithRegion(regionSku),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				m.creds.AccessKeyId,
				m.creds.SecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "Failed Init aws session", "Reason", err),
		)
	}

	return &_session, nil
}

func (m *AwsMeta) GetAvailableRegions() ([]provider.RegionOutput, error) {
	// https://github.com/aws/aws-sdk-go-v2/blob/main/codegen/smithy-aws-go-codegen/src/main/resources/software/amazon/smithy/aws/go/codegen/endpoints.json
	type Service struct {
		Endpoints map[string]interface{} `json:"endpoints"`
	}

	type Partition struct {
		PartitionName string                 `json:"partitionName"`
		Regions       map[string]interface{} `json:"regions"`
		Services      map[string]Service     `json:"services"`
	}

	type endpoints struct {
		Partitions []Partition `json:"partitions"`
	}

	resp, err := http.Get("https://raw.githubusercontent.com/aws/aws-sdk-go-v2/master/codegen/smithy-aws-go-codegen/src/main/resources/software/amazon/smithy/aws/go/codegen/endpoints.json")
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	e := endpoints{}
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		return nil, err
	}
	session, err := m.GetNewSession("")
	if err != nil {
		return nil, err
	}

	describeRegions, err := ec2.NewFromConfig(*session).
		DescribeRegions(m.ctx, &ec2.DescribeRegionsInput{
			AllRegions: aws.Bool(true),
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("opt-in-status"),
					Values: []string{"opt-in-not-required", "opted-in"},
				},
			},
		})
	if err != nil {
		return nil, err
	}
	availRegion := make([]string, 0, len(describeRegions.Regions))
	for _, r := range describeRegions.Regions {
		availRegion = append(availRegion, *r.RegionName)
	}

	var regions = make([]provider.RegionOutput, 0, len(availRegion))
	for _, p := range e.Partitions {
		for r, v := range p.Regions {
			if o, ok := v.(map[string]interface{})["description"].(string); ok {
				if slices.Contains(availRegion, r) {
					regions = append(regions, provider.RegionOutput{
						Sku:  r,
						Name: o,
					})
				}
			}
		}
	}
	m.cachedRegionMapping = regions

	return regions, nil
}

func (m *AwsMeta) searchRegionDescription(regionSku string) *provider.RegionOutput {
	for _, v := range m.cachedRegionMapping {
		if v.Sku == regionSku {
			return &v
		}
	}
	return nil
}

func (m *AwsMeta) GetAvailableInstanceTypes(regionSku string, _ consts.KsctlClusterType) ([]provider.InstanceRegionOutput, error) {
	reg := m.searchRegionDescription(regionSku)

	vms, err := m.listOfVms(regionSku, WithDefaultEC2())
	if err != nil {
		return nil, err
	}

	priceVMs, err := m.priceVMs(*reg)
	if err != nil {
		return nil, err
	}

	disks, err := m.priceDisks(*reg, WithDefaultEBSSize(), WithDefaultEBSVolumeType())
	if err != nil {
		return nil, err
	}

	var outVMs []provider.InstanceRegionOutput
	for i := range vms {
		if price, ok := priceVMs[vms[i].Sku]; ok {
			vms[i].Price = price.Price
			vms[i].Disk = disks[0]
			outVMs = append(outVMs, vms[i])
		}
	}

	return outVMs, nil
}

func (m *AwsMeta) GetAvailableManagedK8sManagementOfferings(regionSku string, vmType *string) ([]provider.ManagedClusterOutput, error) {
	reg := m.searchRegionDescription(regionSku)

	return m.priceEksManagement(*reg, vmType, WithDefaultEKS())
}

func (m *AwsMeta) GetAvailableManagedK8sVersions(regionSku string) ([]string, error) {
	v, err := m.listManagedOfferingsK8sVersions(regionSku)
	if err != nil {
		return nil, err
	}

	m.l.Debug(m.ctx, "Managed K8s versions", "EksVersions", v)
	return v, nil
}

//func getAccountId(cfg aws.Config) (*string, error) {
//	svcP := sts.NewFromConfig(cfg)
//
//	result, err := svcP.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
//	if err != nil {
//		return nil, fmt.Errorf("unable to get caller identity, %v", err)
//	}
//
//	return result.Account, nil
//}
