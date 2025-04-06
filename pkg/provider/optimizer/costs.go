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
	"context"
	"fmt"
	"slices"

	"github.com/fatih/color"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
)

type RecommendationManagedCost struct {
	Region    string
	totalCost float64

	cpCost float64
	wpCost float64
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
			cpCost:    costForCP[region.Sku],
			wpCost:    costForWP[region.Sku],
			totalCost: totalCost,
		})
	}

	slices.SortFunc(costForCluster, func(a, b RecommendationManagedCost) int {
		return cmp.Compare(a.totalCost, b.totalCost)
	})

	return costForCluster
}

type RecommendationSelfManagedCost struct {
	Region    string
	totalCost float64

	cpCost   float64
	wpCost   float64
	etcdCost float64
	lbCost   float64
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
			cpCost:    costForCP[region.Sku],
			wpCost:    costForWP[region.Sku],
			etcdCost:  costForDS[region.Sku],
			lbCost:    costForLB[region.Sku],
			totalCost: totalCost,
		})
	}

	slices.SortFunc(costForCluster, func(a, b RecommendationSelfManagedCost) int {
		return cmp.Compare(a.totalCost, b.totalCost)
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

func (k *Optimizer) PrintRecommendationSelfManagedCost(
	ctx context.Context,
	l logger.Logger,
	costs []RecommendationSelfManagedCost,
	currRegion string,
	noOfCP int,
	noOfWP int,
	noOfDS int,
	instanceTypeCP string,
	instanceTypeWP string,
	instanceTypeDS string,
	instanceTypeLB string,
) {
	l.Print(ctx,
		"Here is your recommendation",
		"Parameter", "Region wise cost",
		"OptimizedRegion", color.HiCyanString(costs[0].Region),
	)

	headers := []string{
		"Region",
		"🏭 Direct Emission",
		fmt.Sprintf("ControlPlane (%s)", instanceTypeCP),
		fmt.Sprintf("WorkerPlane (%s)", instanceTypeWP),
		fmt.Sprintf("DatastorePlane (%s)", instanceTypeDS),
		fmt.Sprintf("LoadBalancer (%s)", instanceTypeLB),
		"Total Monthly Cost",
	}

	var data [][]string
	for _, cost := range costs {
		total := cost.cpCost*float64(noOfCP) + cost.wpCost*float64(noOfWP) + cost.etcdCost*float64(noOfDS) + cost.lbCost
		reg := cost.Region
		if reg == currRegion {
			reg += "*"
		}
		regEmissions := ""
		if v, ok := k.getRegionsInMapFormat()[reg]; ok && v.Emission != nil {
			regEmissions = fmt.Sprintf("%.2f %s",
				v.Emission.DirectCarbonIntensity,
				v.Emission.Unit,
			)
		} else if reg == currRegion {
			regEmissions = "*"
		} else {
			regEmissions = "NA"
		}

		data = append(data, []string{
			reg,
			regEmissions,
			fmt.Sprintf("$%.2f X %d", cost.cpCost, noOfCP),
			fmt.Sprintf("$%.2f X %d", cost.wpCost, noOfWP),
			fmt.Sprintf("$%.2f X %d", cost.etcdCost, noOfDS),
			fmt.Sprintf("$%.2f X 1", cost.lbCost),
			fmt.Sprintf("$%.2f", total),
		})
	}

	l.Table(ctx, headers, data)
}

func (k *Optimizer) PrintRecommendationManagedCost(
	ctx context.Context,
	l logger.Logger,
	costs []RecommendationManagedCost,
	currRegion string,
	noOfWP int,
	managedOfferingCP string,
	instanceTypeWP string,
) {
	l.Print(ctx,
		"Here is your recommendation",
		"Parameter", "Region wise cost",
		"OptimizedRegion", color.HiCyanString(costs[0].Region),
	)

	headers := []string{
		"Region",
		"🏭 Direct Emission",
		fmt.Sprintf("ControlPlane (%s)", managedOfferingCP),
		fmt.Sprintf("WorkerPlane (%s)", instanceTypeWP),
		"Total Monthly Cost",
	}

	var data [][]string
	for _, cost := range costs {
		total := cost.cpCost + cost.wpCost*float64(noOfWP)
		reg := cost.Region
		if reg == currRegion {
			reg += "*"
		}
		regEmissions := ""
		if v, ok := k.getRegionsInMapFormat()[reg]; ok && v.Emission != nil {
			regEmissions = fmt.Sprintf("%.2f %s",
				v.Emission.DirectCarbonIntensity,
				v.Emission.Unit,
			)
		} else if reg == currRegion {
			regEmissions = "*"
		} else {
			regEmissions = "NA"
		}
		data = append(data, []string{
			reg,
			regEmissions,
			fmt.Sprintf("$%.2f X 1", cost.cpCost),
			fmt.Sprintf("$%.2f X %d", cost.wpCost, noOfWP),
			fmt.Sprintf("$%.2f", total),
		})
	}

	l.Table(ctx, headers, data)
}
