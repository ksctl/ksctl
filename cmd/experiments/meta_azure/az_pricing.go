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
	"log"
	"os"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/provider/azure"
)

func main() {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	cc, err := azure.NewResourceDetails(subscriptionID)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	v, err := cc.VMs(context.Background(), "eastus")
	if err != nil {
		log.Fatalf("failed to get VMs: %v", err)
	}

	dump.Println(v[azure.VMSku("Standard_F4s")])
	dump.Println(v[azure.VMSku("Standard_D2_v3")])

	pp, err := cc.Disks(context.Background(), "eastus")
	if err != nil {
		log.Fatalf("failed to get disks: %v", err)
	}
	dump.Println(pp)

	p, err := cc.AKS(
		context.Background(),
		"eastus",
		"Standard_D2_v3",
		2,
		azure.Standard,
	)
	if err != nil {
		log.Fatalf("failed to get AKS: %v", err)
	}
	dump.Println(p)
}
