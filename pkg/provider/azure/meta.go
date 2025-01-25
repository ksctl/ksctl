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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type Region struct {
	Code        string
	Description string
}

type RegionCode string
type RegionDescription string

type Regions map[RegionCode]RegionDescription

type ResourceDetails struct {
	ctx            context.Context
	regions        Regions
	cred           azcore.TokenCredential
	subscriptionId string
}

func NewResourceDetails(subscriptionId string) (*ResourceDetails, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	regions, err := getAllRegions(cred, subscriptionId)
	if err != nil {
		return nil, err
	}

	return &ResourceDetails{
		cred:           cred,
		regions:        regions,
		subscriptionId: subscriptionId,
	}, nil
}

func getAllRegions(cred azcore.TokenCredential, subscriptionId string) (Regions, error) {

	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return nil, err
	}

	pager := clientFactory.NewClient().NewListLocationsPager(
		subscriptionId,
		&armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil},
	)

	var regions = Regions{}
	ctx := context.TODO()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range page.Value {
			regions[RegionCode(*v.Name)] = RegionDescription(*v.DisplayName)
		}
	}
	return regions, nil
}
