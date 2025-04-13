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

func (k *Optimizer) getRegionsWithTotalCostManaged(
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

func (k *Optimizer) getRegionsWithTotalCostSelfManaged(
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

	return k.getRegionsWithTotalCostSelfManaged(
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

	return k.getRegionsWithTotalCostManaged(
		cpInstanceCosts,
		wpInstanceCosts,
	)
}

type RegionRecommendation struct {
	Region           provider.RegionOutput `json:"region"`
	ControlPlaneCost float64               `json:"controlPlaneCost"`
	WorkerPlaneCost  float64               `json:"workerPlaneCost"`
	DataStoreCost    float64               `json:"dataStoreCost,omitempty"`
	LoadBalancerCost float64               `json:"loadBalancerCost,omitempty"`
	TotalCost        float64               `json:"totalCost"`
}

type RecommendationAcrossRegions struct {
	RegionRecommendations []RegionRecommendation `json:"regionRecommendations"`
	CurrentRegion         provider.RegionOutput  `json:"current_region"`
	CurrentTotalCost      float64                `json:"current_total_cost"`

	InstanceTypeCP string `json:"instanceTypeCP,omitempty"`
	InstanceTypeWP string `json:"instanceTypeWP"`
	InstanceTypeDS string `json:"instanceTypeDS,omitempty"`
	InstanceTypeLB string `json:"instanceTypeLB,omitempty"`

	ManagedOffering string `json:"managedOffering,omitempty"`

	ControlPlaneCount int `json:"controlPlaneCount,omitempty"`
	WorkerPlaneCount  int `json:"workerPlaneCount"`
	DataStoreCount    int `json:"dataStoreCount,omitempty"`
}

func isA_sEmissionLowerOrEqual(a, b provider.RegionalEmission) bool {
	dco2Cmp := cmp.Compare(a.DirectCarbonIntensity, b.DirectCarbonIntensity)
	rpCmp := cmp.Compare(a.RenewablePercentage, b.RenewablePercentage)
	lco2Cmp := cmp.Compare(a.LowCarbonPercentage, b.LowCarbonPercentage)
	lcaco2Cmp := cmp.Compare(a.LCACarbonIntensity, b.LCACarbonIntensity)

	if dco2Cmp == 1 || rpCmp == -1 || lco2Cmp == -1 || lcaco2Cmp == 1 {
		return false
	}
	if dco2Cmp == 1 {
		return true
	}
	if rpCmp == 1 {
		return true
	}
	if lco2Cmp == -1 {
		return true
	}

	if lcaco2Cmp > 0 {
		return false
	}
	return true
}

func emissionComparator(a, b RegionRecommendation) int {
	if a.Region.Emission == nil || b.Region.Emission == nil {
		return 0
	}

	dco2Cmp := cmp.Compare(a.Region.Emission.DirectCarbonIntensity, b.Region.Emission.DirectCarbonIntensity)
	rpCmp := cmp.Compare(b.Region.Emission.RenewablePercentage, a.Region.Emission.RenewablePercentage)   // Higher is better
	lco2Cmp := cmp.Compare(b.Region.Emission.LowCarbonPercentage, a.Region.Emission.LowCarbonPercentage) // Higher is better
	lcaco2Cmp := cmp.Compare(a.Region.Emission.LCACarbonIntensity, b.Region.Emission.LCACarbonIntensity)

	if dco2Cmp != 0 {
		return dco2Cmp
	}

	if rpCmp != 0 {
		return rpCmp
	}

	if lco2Cmp != 0 {
		return lco2Cmp
	}

	return lcaco2Cmp
}

// costPlusEmissionDecision return
//
//	Refer to test case to understand more TestInstanceTypeOptimizerAcrossRegionsSelfManaged
func costPlusEmissionDecision(
	cost_r float64,
	cost_R float64,
	emission_r *provider.RegionalEmission,
	emission_R *provider.RegionalEmission,
) int {

	if cost_r > cost_R {
		return 2
	}

	if cost_r < cost_R {
		// Note: it removes regions with higher emissions but lower costs
		//addIt := true
		//if emission_r != nil && emission_R != nil {
		//	if !isA_sEmissionLowerOrEqual(*emission_r, *emission_R) {
		//		addIt = false
		//	}
		//}
		//
		//if addIt {
		//	return -1
		//}
		return -1
	} else if cost_r == cost_R {
		addIt := true
		if emission_r == nil {
			return 0 // no emissions
		}

		if emission_R == nil { // when selected region emissions are absent
			return 1
		} else {
			if !isA_sEmissionLowerOrEqual(*emission_r, *emission_R) {
				addIt = false
			}
		}

		if addIt {
			return 1
		}
	}

	return -2
}

// InstanceTypeOptimizerAcrossRegions is used to get the best regions for the given instance types across all the regions
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
		CurrentRegion:    regions[currRegion],
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

	var regWithlowerCost, regWithlowerEmission, regWithNoEmissions []RegionRecommendation

	if clusterType == consts.ClusterTypeMang {
		slices.SortStableFunc(costsManaged, func(a, b RecommendationManagedCost) int {
			return cmp.Compare(a.TotalCost, b.TotalCost)
		})

		for _, cost := range costsManaged {
			if cost.Region == currRegion {
				res.CurrentTotalCost = cost.CpCost + cost.WpCost*float64(noOfWP)
				break
			}
		}

		for _, cost := range costsManaged {
			if cost.Region == currRegion {
				continue
			}
			total := cost.CpCost + cost.WpCost*float64(noOfWP)

			var regionEmissions *provider.RegionalEmission
			if v, ok := regions[cost.Region]; ok && v.Emission != nil {
				regionEmissions = v.Emission
			}

			item := RegionRecommendation{
				Region:           regions[cost.Region],
				ControlPlaneCost: cost.CpCost,
				WorkerPlaneCost:  cost.WpCost,
				TotalCost:        total,
			}

			resCmp := costPlusEmissionDecision(
				total,
				res.CurrentTotalCost,
				regionEmissions,
				res.CurrentRegion.Emission,
			)
			if resCmp == 2 {
				break
			}
			if resCmp == -1 {
				regWithlowerCost = append(regWithlowerCost, item)
			}
			if resCmp == 0 {
				regWithNoEmissions = append(regWithNoEmissions, item)
			}
			if resCmp == 1 {
				regWithlowerEmission = append(regWithlowerEmission, item)
			}
		}
	} else if clusterType == consts.ClusterTypeSelfMang {
		slices.SortStableFunc(costsSelfManaged, func(a, b RecommendationSelfManagedCost) int {
			return cmp.Compare(a.TotalCost, b.TotalCost)
		})

		for _, cost := range costsSelfManaged {
			total := cost.CpCost*float64(noOfCP) + cost.WpCost*float64(noOfWP) + cost.EtcdCost*float64(noOfDS) + cost.LbCost
			if cost.Region == currRegion {
				res.CurrentTotalCost = total
				break
			}
		}

		for _, cost := range costsSelfManaged {
			if cost.Region == currRegion {
				continue
			}
			total := cost.CpCost*float64(noOfCP) + cost.WpCost*float64(noOfWP) + cost.EtcdCost*float64(noOfDS) + cost.LbCost

			var regionEmissions *provider.RegionalEmission
			if v, ok := regions[cost.Region]; ok && v.Emission != nil {
				regionEmissions = v.Emission
			}

			item := RegionRecommendation{
				Region:           regions[cost.Region],
				ControlPlaneCost: cost.CpCost,
				WorkerPlaneCost:  cost.WpCost,
				DataStoreCost:    cost.EtcdCost,
				LoadBalancerCost: cost.LbCost,
				TotalCost:        total,
			}

			resCmp := costPlusEmissionDecision(
				total,
				res.CurrentTotalCost,
				regionEmissions,
				res.CurrentRegion.Emission,
			)
			if resCmp == 2 {
				break
			}
			if resCmp == -1 {
				regWithlowerCost = append(regWithlowerCost, item)
			}
			if resCmp == 0 {
				regWithNoEmissions = append(regWithNoEmissions, item)
			}
			if resCmp == 1 {
				regWithlowerEmission = append(regWithlowerEmission, item)
			}
		}
	}

	slices.SortStableFunc(regWithlowerEmission, emissionComparator)

	slices.SortStableFunc(regWithlowerCost, func(a, b RegionRecommendation) int {
		_cmp := cmp.Compare(a.TotalCost, b.TotalCost)
		if _cmp == 0 {
			return emissionComparator(a, b)
		}
		return _cmp
	})

	res.RegionRecommendations = append(res.RegionRecommendations, regWithlowerCost...)
	res.RegionRecommendations = append(res.RegionRecommendations, regWithlowerEmission...)
	res.RegionRecommendations = append(res.RegionRecommendations, regWithNoEmissions...)

	return res, nil
}
