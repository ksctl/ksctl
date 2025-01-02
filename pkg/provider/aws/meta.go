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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Region struct {
	Code        string
	Description string
}

type RegionCode string
type RegionDescription string

type Regions map[RegionCode]RegionDescription

// TODO: use the svc.Ec2.DescribeRegions with the default regions and filter out the below regioncodes!!!
func GetAllRegions() (Regions, error) {
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

	var regions = make(Regions)
	for _, p := range e.Partitions {
		for r, v := range p.Regions {
			if o, ok := v.(map[string]interface{})["description"].(string); ok {
				regions[RegionCode(r)] = RegionDescription(o)
			}
		}
	}
	return regions, nil
}
