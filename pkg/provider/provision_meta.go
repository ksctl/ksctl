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

package provider

import (
	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/consts"
)

type RegionOutput struct {
	Sku  string
	Name string
}

type PriceOutput struct {
	Currency     string
	HourlyPrice  *float64
	MonthlyPrice *float64
}

type MachineCategory string

const (
// GeneralPurposeMachine MachineCategory = "general-purpose"
// ComputeOptimized      MachineCategory = "compute-optimized"
// GpuMachine            MachineCategory = "gpu"
// SpotMachine           MachineCategory = "spot"
)

type MachineArchitecture string

const (
	ArchArm64 MachineArchitecture = "arm64"
	ArchAmd64 MachineArchitecture = "amd64"
)

type InstanceRegionOutput struct {
	// Sku is the SKU of the instance
	Sku         string
	Description string

	// VCpus is the number of virtual CPUs
	VCpus int32

	// Memory is in GB
	Memory int32

	// CpuArch is the architecture of the CPU
	CpuArch MachineArchitecture

	Category MachineCategory
	// MostlyUsedFor is used what does the category tells about
	MostlyUsedFor []string

	// EphemeralOSDiskSupport is true if the instance supports ephemeral OS disk
	//  Currently being used for AKS to make cluster easier to update versions,...
	EphemeralOSDiskSupport bool

	Price PriceOutput

	Disk StorageBlockRegionOutput
}

func (I InstanceRegionOutput) GetCost() float64 {
	machineCostPerMonth := 0.0

	if I.Price.HourlyPrice != nil {
		machineCostPerMonth = *I.Price.HourlyPrice * 730
	}
	if I.Price.MonthlyPrice != nil {
		machineCostPerMonth = *I.Price.MonthlyPrice
	}

	return machineCostPerMonth + I.Disk.GetCost()
}

type DiskAttachmentType string

const (
	InstanceDiskIncluded    DiskAttachmentType = "included"
	InstanceDiskNotIncluded DiskAttachmentType = "not-included"
)

type StorageBlockRegionOutput struct {
	// DiskAttachmentType is the type of disk attachment
	//  If it is InstanceDiskIncluded then the disk is included in the instance and cost is included in the instance
	//  If it is InstanceDiskNotIncluded then the disk is not included in the instance and cost is not included in the instance
	AttachmentType DiskAttachmentType

	MaxIOps *int32

	MinIOps *int32

	// MaxThroughput in MBps
	MaxThroughput *int32

	// MinThroughput in MBps
	MinThroughput *int32

	// Size is in GB
	Size *int32

	// Sku in terms of Disk in AZ is E6
	Sku *string

	// Tier in terms of Disk in AZ is StandardSSD_LRS
	Tier *string

	// Price are included in the instance if its managed by the instance itself like For ex. Hetzner, DigitalOcean
	Price *PriceOutput
}

func (S StorageBlockRegionOutput) GetCost() float64 {
	if S.Price == nil {
		return 0.0
	}
	costPerMonth := 0.0

	if S.Price.HourlyPrice != nil {
		costPerMonth = *S.Price.HourlyPrice * 730
	}
	if S.Price.MonthlyPrice != nil {
		costPerMonth = *S.Price.MonthlyPrice
	}

	return costPerMonth
}

type ManagedClusterOutput struct {
	Sku         string
	Description string
	Tier        string
	Price       PriceOutput
}

func (M ManagedClusterOutput) GetCost() float64 {
	cost := 0.0

	if M.Price.HourlyPrice != nil {
		cost = *M.Price.HourlyPrice * 730
	}
	if M.Price.MonthlyPrice != nil {
		cost = *M.Price.MonthlyPrice
	}

	return cost
}

type ProvisionMetadata interface {
	GetAvailableRegions() ([]RegionOutput, error)

	GetAvailableInstanceTypes(regionSku string, clusterType consts.KsctlClusterType) ([]InstanceRegionOutput, error)

	GetAvailableManagedK8sManagementOfferings(regionSku string, vmType *string) ([]ManagedClusterOutput, error)

	GetAvailableManagedK8sVersions(regionSku string) ([]string, error)

	GetAvailableManagedCNIPlugins(regionSku string) (addons.ClusterAddons, string, error)
}
