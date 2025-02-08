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
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func (m *AzureMeta) listOfDisksStandardLRS_ESeries(region string) (out []provider.StorageBlockRegionOutput, _ error) {
	clientFactory, err := armcompute.NewClientFactory(m.subscriptionId, m.cred, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed in azure client", "Reason", err),
		)
	}
	pager := clientFactory.NewResourceSKUsClient().NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter:                   utilities.Ptr(fmt.Sprintf("location eq '%s'", region)),
		IncludeExtendedLocations: nil,
	})

	allowedStorageTierWithSizes := func(size, name string) bool {
		return strings.HasPrefix(size, "E") && name == "StandardSSD_LRS"
	}

	for pager.More() {
		page, err := pager.NextPage(m.ctx)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			if v.Size == nil {
				continue
			}
			if *v.ResourceType == "disks" && *v.Tier == "Standard" && allowedStorageTierWithSizes(*v.Size, *v.Name) {
				o := provider.StorageBlockRegionOutput{
					Tier:           v.Name,
					Sku:            v.Size,
					AttachmentType: provider.InstanceDiskNotIncluded,
				}

				if v.Capabilities != nil {
					for _, cap := range v.Capabilities {
						if *cap.Name == "MaxSizeGiB" {
							v, _ := strconv.Atoi(*cap.Value)
							o.Size = utilities.Ptr(int32(v))
						}
						if *cap.Name == "MaxIOps" {
							v, _ := strconv.Atoi(*cap.Value)
							o.MaxIOps = utilities.Ptr(int32(v))
						}
						if *cap.Name == "MinIOps" {
							v, _ := strconv.Atoi(*cap.Value)
							o.MinIOps = utilities.Ptr(int32(v))
						}
						if *cap.Name == "MaxBandwidthMBps" {
							v, _ := strconv.Atoi(*cap.Value)
							o.MaxThroughput = utilities.Ptr(int32(v))
						}
						if *cap.Name == "MinBandwidthMBps" {
							v, _ := strconv.Atoi(*cap.Value)
							o.MinThroughput = utilities.Ptr(int32(v))
						}
					}
				}
				out = append(out, o)
			}
		}
	}

	return out, nil
}

func (m *AzureMeta) listOfVms(region string) (out []provider.InstanceRegionOutput, _ error) {
	clientFactory, err := armcompute.NewClientFactory(m.subscriptionId, m.cred, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed in azure client", "Reason", err),
		)
	}
	pager := clientFactory.NewResourceSKUsClient().NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter:                   utilities.Ptr(fmt.Sprintf("location eq '%s'", region)),
		IncludeExtendedLocations: nil,
	})

	for pager.More() {
		page, err := pager.NextPage(m.ctx)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			if *v.ResourceType == "virtualMachines" && *v.Tier == "Standard" {
				o := provider.InstanceRegionOutput{
					Sku:         *v.Name,
					Description: *v.Name,
				}
				if v.Capabilities != nil {
					for _, cap := range v.Capabilities {
						if *cap.Name == "vCPUs" {
							v, _ := strconv.Atoi(*cap.Value)
							o.VCpus = int32(v)
						}
						if *cap.Name == "MemoryGB" {
							v, _ := strconv.Atoi(*cap.Value)
							o.Memory = int32(v)
						}
						if *cap.Name == "CpuArchitectureType" {
							if *cap.Value == "x64" {
								o.CpuArch = provider.ArchAmd64
							} else if *cap.Value == "Arm64" {
								o.CpuArch = provider.ArchArm64
							}
						}
						if *cap.Name == "EphemeralOSDiskSupported" {
							if *cap.Value == "true" {
								o.EphemeralOSDiskSupport = true
							} else {
								o.EphemeralOSDiskSupport = false
							}
						}
					}
				}
				out = append(out, o)
			}
		}
	}

	return out, nil
}
