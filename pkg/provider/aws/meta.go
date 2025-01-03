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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"io"
	"net/http"
	"slices"
)

type Region struct {
	Code        string
	Description string
}

type RegionCode string
type RegionDescription string

type Regions map[RegionCode]RegionDescription

func GetAllRegions(cfg aws.Config) (Regions, error) {
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

	fmt.Println("Generating AWS regions")

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

	ec2Svc := ec2.NewFromConfig(cfg)
	describeRegions, err := ec2Svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
		Filters: []types.Filter{
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

	var regions = make(Regions)
	for _, p := range e.Partitions {
		for r, v := range p.Regions {
			if o, ok := v.(map[string]interface{})["description"].(string); ok {
				if slices.Contains(availRegion, r) {
					regions[RegionCode(r)] = RegionDescription(o)
				}
			}
		}
	}
	return regions, nil
}

func getAccountId(cfg aws.Config) (*string, error) {
	svcP := sts.NewFromConfig(cfg)

	result, err := svcP.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to get caller identity, %v", err)
	}

	return result.Account, nil
}
