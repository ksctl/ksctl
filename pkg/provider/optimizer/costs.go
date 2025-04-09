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

package optimizer

import (
	"cmp"
	"slices"

	"github.com/ksctl/ksctl/v2/pkg/consts"

	"github.com/ksctl/ksctl/v2/pkg/provider"
)

type RecommendationManagedCost struct {
	Region    string
	TotalCost float64

	CpCost float64
	WpCost float64
}

func (k *Optimizer) getBestRegionsWithTotalCostManaged(
	costForCP map[string]float64,
	costForWP map[string]float64,
) []RecommendationManagedCost {

	checkRegion := func(region string, m map[string]float64) bool {
		_, ok := m[region]
		return ok
	}

	var costForCluster []RecommendationManagedCost

	for _, region := range k.AvailRegions {
		if !checkRegion(region.Sku, costForCP) ||
			!checkRegion(region.Sku, costForWP) {
			continue
		}

		totalCost := costForCP[region.Sku] + costForWP[region.Sku]

		costForCluster = append(costForCluster, RecommendationManagedCost{
			Region:    region.Sku,
			CpCost:    costForCP[region.Sku],
			WpCost:    costForWP[region.Sku],
			TotalCost: totalCost,
		})
	}

	slices.SortFunc(costForCluster, func(a, b RecommendationManagedCost) int {
		return cmp.Compare(a.TotalCost, b.TotalCost)
	})

	return costForCluster
}

type RecommendationSelfManagedCost struct {
	Region    string
	TotalCost float64

	CpCost   float64
	WpCost   float64
	EtcdCost float64
	LbCost   float64
}

func (k *Optimizer) getBestRegionsWithTotalCostSelfManaged(
	costForCP map[string]float64,
	costForWP map[string]float64,
	costForDS map[string]float64,
	costForLB map[string]float64,
) []RecommendationSelfManagedCost {

	checkRegion := func(region string, m map[string]float64) bool {
		_, ok := m[region]
		return ok
	}

	var costForCluster []RecommendationSelfManagedCost

	for _, region := range k.AvailRegions {
		if !checkRegion(region.Sku, costForCP) ||
			!checkRegion(region.Sku, costForWP) ||
			!checkRegion(region.Sku, costForDS) ||
			!checkRegion(region.Sku, costForLB) {
			continue
		}

		totalCost := costForCP[region.Sku] + costForWP[region.Sku] + costForDS[region.Sku] + costForLB[region.Sku]

		costForCluster = append(costForCluster, RecommendationSelfManagedCost{
			Region:    region.Sku,
			CpCost:    costForCP[region.Sku],
			WpCost:    costForWP[region.Sku],
			EtcdCost:  costForDS[region.Sku],
			LbCost:    costForLB[region.Sku],
			TotalCost: totalCost,
		})
	}

	slices.SortFunc(costForCluster, func(a, b RecommendationSelfManagedCost) int {
		return cmp.Compare(a.TotalCost, b.TotalCost)
	})

	return costForCluster
}

// OptimizeSelfManagedInstanceTypesAcrossRegions it returns a sorted list of regions based on the total cost in ascending order across all the regions
//
//	It is a core function that is used to optimize the cost of the self-managed cluster instanceType across all the regions (Cost Optimization)
func (k *Optimizer) OptimizeSelfManagedInstanceTypesAcrossRegions(
	cp provider.InstanceRegionOutput,
	wp provider.InstanceRegionOutput,
	etcd provider.InstanceRegionOutput,
	lb provider.InstanceRegionOutput,

	instanceCostGetter func(allRegions []provider.RegionOutput, instanceSku string) (map[string]float64, error),
) []RecommendationSelfManagedCost {
	cpInstanceCosts, err := instanceCostGetter(k.AvailRegions, cp.Sku)
	if err != nil {
		k.l.Error("Failed to get the cost of control plane instances", "Reason", err)
	}

	wpInstanceCosts, err := instanceCostGetter(k.AvailRegions, wp.Sku)
	if err != nil {
		k.l.Error("Failed to get the cost of worker plane instances", "Reason", err)
	}

	etcdInstanceCosts, err := instanceCostGetter(k.AvailRegions, etcd.Sku)
	if err != nil {
		k.l.Error("Failed to get the cost of etcd instances", "Reason", err)
	}

	lbInstanceCosts, err := instanceCostGetter(k.AvailRegions, lb.Sku)
	if err != nil {
		k.l.Error("Failed to get the cost of load balancer instances", "Reason", err)
	}

	return k.getBestRegionsWithTotalCostSelfManaged(
		cpInstanceCosts,
		wpInstanceCosts,
		etcdInstanceCosts,
		lbInstanceCosts,
	)
}

func (k *Optimizer) OptimizeManagedOfferingsAcrossRegions(
	cp provider.ManagedClusterOutput,
	wp provider.InstanceRegionOutput,

	instanceCostGetter, managedOfferingCostGetter func(allRegions []provider.RegionOutput, instanceSku string) (map[string]float64, error),
) []RecommendationManagedCost {
	wpInstanceCosts, err := instanceCostGetter(k.AvailRegions, wp.Sku)
	if err != nil {
		k.l.Error("Failed to get the cost of worker plane instances", "Reason", err)
	}

	cpInstanceCosts, err := managedOfferingCostGetter(k.AvailRegions, cp.Sku)
	if err != nil {
		k.l.Error("Failed to get the cost of control plane managed offerings", "Reason", err)
	}

	return k.getBestRegionsWithTotalCostManaged(
		cpInstanceCosts,
		wpInstanceCosts,
	)
}

type RegionRecommendation struct {
	Region           string                     `json:"region"`
	Emissions        *provider.RegionalEmission `json:"emissions,omitempty"`
	ControlPlaneCost float64                    `json:"controlPlaneCost"`
	WorkerPlaneCost  float64                    `json:"workerPlaneCost"`
	DataStoreCost    float64                    `json:"dataStoreCost,omitempty"`
	LoadBalancerCost float64                    `json:"loadBalancerCost,omitempty"`
	TotalCost        float64                    `json:"totalCost"`
}

type RecommendationAcrossRegions struct {
	RegionRecommendations []RegionRecommendation     `json:"regionRecommendations"`
	CurrentRegion         string                     `json:"current_region"`
	CurrentTotalCost      float64                    `json:"current_total_cost"`
	CurrentEmissions      *provider.RegionalEmission `json:"current_emissions"`

	InstanceTypeCP string `json:"instanceTypeCP,omitempty"`
	InstanceTypeWP string `json:"instanceTypeWP"`
	InstanceTypeDS string `json:"instanceTypeDS,omitempty"`
	InstanceTypeLB string `json:"instanceTypeLB,omitempty"`

	ManagedOffering string `json:"managedOffering,omitempty"`

	ControlPlaneCount int `json:"controlPlaneCount,omitempty"`
	WorkerPlaneCount  int `json:"workerPlaneCount"`
	DataStoreCount    int `json:"dataStoreCount,omitempty"`
}

// InstanceTypeOptimizerAcrossRegions is used to get the best regions for the given instance types across all the regions
//
//	TODO: also wrt to the emissions as well!!!
func (k *Optimizer) InstanceTypeOptimizerAcrossRegions(
	regions map[string]provider.RegionOutput,
	clusterType consts.KsctlClusterType,
	costsManaged []RecommendationManagedCost,
	costsSelfManaged []RecommendationSelfManagedCost,
	currRegion string,
	noOfCP int,
	noOfWP int,
	noOfDS int,
	managedOfferingCP string,
	instanceTypeCP string,
	instanceTypeWP string,
	instanceTypeDS string,
	instanceTypeLB string,
) (res RecommendationAcrossRegions, errC error) {
	res = RecommendationAcrossRegions{
		CurrentRegion:    currRegion,
		CurrentEmissions: nil,
		CurrentTotalCost: 0.0,

		InstanceTypeCP:  instanceTypeCP,
		ManagedOffering: managedOfferingCP,

		InstanceTypeWP: instanceTypeWP,

		InstanceTypeDS: instanceTypeDS,
		InstanceTypeLB: instanceTypeLB,

		ControlPlaneCount: noOfCP,
		WorkerPlaneCount:  noOfWP,
		DataStoreCount:    noOfDS,
	}

	var (
		lowerCostReg []RegionRecommendation

		// lowerEmissionReg we use it when the costs are same as the current region
		lowerEmissionReg []RegionRecommendation
	)

	if clusterType == consts.ClusterTypeMang {
		for _, cost := range costsManaged {
			if cost.Region == currRegion {
				res.CurrentTotalCost = cost.CpCost + cost.WpCost*float64(noOfWP)
				if v, ok := regions[cost.Region]; ok && v.Emission != nil {
					res.CurrentEmissions = v.Emission
				}
				break
			}
		}

		for _, cost := range costsManaged {
			total := cost.CpCost + cost.WpCost*float64(noOfWP)

			var regionEmissions *provider.RegionalEmission
			if v, ok := regions[cost.Region]; ok && v.Emission != nil {
				regionEmissions = v.Emission
			}

			if total > res.CurrentTotalCost {
				break
			}

			if total == res.CurrentTotalCost {
				// we need to compute something else aka emissions!!

			} else {
				lowerCostReg = append(lowerCostReg, RegionRecommendation{
					Region:           cost.Region,
					ControlPlaneCost: cost.CpCost,
					WorkerPlaneCost:  cost.WpCost,
					TotalCost:        total,
					Emissions:        regionEmissions,
				})
			}
		}
	} else if clusterType == consts.ClusterTypeSelfMang {
		for _, cost := range costsSelfManaged {
			total := cost.CpCost*float64(noOfCP) + cost.WpCost*float64(noOfWP) + cost.EtcdCost*float64(noOfDS) + cost.LbCost
			if cost.Region == currRegion {
				res.CurrentTotalCost = total
				if v, ok := regions[cost.Region]; ok && v.Emission != nil {
					res.CurrentEmissions = v.Emission
				}
				break
			}
		}

		for _, cost := range costsSelfManaged {
			total := cost.CpCost*float64(noOfCP) + cost.WpCost*float64(noOfWP) + cost.EtcdCost*float64(noOfDS) + cost.LbCost

			var regionEmissions *provider.RegionalEmission
			if v, ok := regions[cost.Region]; ok && v.Emission != nil {
				regionEmissions = v.Emission
			}

			if total > res.CurrentTotalCost {
				break
			}

			if total == res.CurrentTotalCost {
				// we need to compute something else aka emissions!!
			} else {
				lowerCostReg = append(lowerCostReg, RegionRecommendation{
					Region:           cost.Region,
					ControlPlaneCost: cost.CpCost,
					WorkerPlaneCost:  cost.WpCost,
					DataStoreCost:    cost.EtcdCost,
					LoadBalancerCost: cost.LbCost,
					TotalCost:        total,
					Emissions:        regionEmissions,
				})
			}
		}
	}

	res.RegionRecommendations = append(res.RegionRecommendations, lowerCostReg...)
	res.RegionRecommendations = append(res.RegionRecommendations, lowerEmissionReg...)

	return res, nil
}
