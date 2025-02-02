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

type RegionOutput struct {
	Sku  string
	Name string
}

type PriceOutput struct {
	HourlyPrice  *float64
	MonthlyPrice *float64
}

type InstanceRegionInput string

type MachineCategory string

const (
	GeneralPurposeMachine MachineCategory = "general-purpose"
	ComputeOptimized      MachineCategory = "compute-optimized"
	GpuMachine            MachineCategory = "gpu"
	SpotMachine           MachineCategory = "spot"
)

type InstanceRegionOutput struct {
	// Sku is the SKU of the instance
	Sku string

	// VCpu is the number of virtual CPUs
	VCpu uint8

	// Memory is in MB
	Memory uint32

	// CpuArch is the architecture of the CPU
	CpuArch string

	Category MachineCategory

	Price PriceOutput

	Disk StorageBlockRegionOutput
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

	// DiskSize is the size of the disk in GB
	DiskSize uint32

	// Sku is the SKU of the disk only gets populated if the disk is not included in the instance
	Sku *string

	// Price are included in the instance if its managed by the instance itself like For ex. Hetzner, DigitalOcean
	Price *PriceOutput
}

type ManagedClusterOutput struct {
	Sku   string
	Tier  string
	Price PriceOutput
}

type ProvisionMetadata interface {
	GetAvailableRegions() ([]RegionOutput, error)

	GetAvailableInstanceTypes(InstanceRegionInput) ([]InstanceRegionOutput, error)

	GetAvailableManagedK8sManagementOfferings() ([]ManagedClusterOutput, error)

	GetAvailableManagedK8sVersions() ([]string, error)
}
