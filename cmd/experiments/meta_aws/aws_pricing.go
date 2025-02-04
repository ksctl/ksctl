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

//import (
//	"context"
//	"fmt"
//	"log"
//	"strings"
//
//	"github.com/aws/aws-sdk-go-v2/config"
//	"github.com/gookit/goutil/dump"
//	"github.com/ksctl/ksctl/v2/pkg/provider/aws"
//)
//
//func main() {
//	// Load the Shared AWS Configuration (~/.aws/config)
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
//	if err != nil {
//		log.Fatalf("unable to load SDK config, %v", err)
//	}
//
//	client, err := aws.NewResourceDetails(cfg)
//	if err != nil {
//		log.Fatalf("unable to load SDK config, %v", err)
//	}
//
//	summary := make([]string, 0, 5)
//	for _, d := range [...][2]string{
//		{"ap-south-1", "t2.large"},
//		{"ap-south-1", "t3.medium"},
//		{"ap-south-1", "t3.large"},
//		{"ap-south-1", "t4g.medium"},
//		{"ap-south-1", "t4g.large"},
//		{"ap-south-1", "m6g.large"},
//		// {"ap-south-1", "m8g.medium"},
//		// {"ap-south-1", "m8g.large"},
//	} {
//
//		fmt.Println("$>", d)
//		if v, e := client.EC2(d[0], d[1]); e != nil {
//			fmt.Println(">===========", e)
//		} else {
//			dump.Println(v)
//			summary = append(summary, v.PriceDescription+" "+v.VCPU+" "+v.Memory)
//		}
//	}
//
//	fmt.Println(strings.Join(summary, "\n"))
//
//	for _, d := range [...][5]any{
//		{"ap-south-1", "t2.micro", 2, false, aws.EksTypeStandard},
//		{"ap-south-1", "t2.medium", 2, true, aws.EksTypeStandard},
//	} {
//		fmt.Println("$>", d)
//		if v, e := client.EKS(d[0].(string), d[1].(string), d[2].(int), d[3].(bool), d[4].(aws.EksType)); e != nil {
//			fmt.Println(">===========", e)
//		} else {
//			dump.Println(v)
//		}
//	}
//
//	for _, d := range [...][3]any{
//		{aws.EBSGp3, 1000, "ap-south-1"},
//	} {
//		fmt.Println("$>", d)
//		if v, e := client.EBS(d[0].(aws.EBSVolumeType), d[1].(int), d[2].(string)); e != nil {
//			fmt.Println(">===========", e)
//		} else {
//			dump.Println(v)
//		}
//	}
//}
