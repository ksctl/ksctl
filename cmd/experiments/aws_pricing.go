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

package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ksctl/ksctl/pkg/provider/aws"
	"log"
)

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	r, _ := aws.GetAllRegions()

	fmt.Printf("%#+v\n", r)

	dd := [...][2]string{
		{"us-east-1", "t3.micro"},
		{"ap-south-1", "t2.micro"},
		{"us-east-2", "t2.micro"},
		{"us-west-2", "t3.small"},
	}

	for _, d := range dd {
		if e := aws.PricingForEC2(cfg, string(r[aws.RegionCode(d[0])]), d[1]); e != nil {
			fmt.Println(">===========", e)
		}
	}
}
