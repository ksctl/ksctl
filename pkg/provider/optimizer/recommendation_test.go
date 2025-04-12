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
	"context"
	"os"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"gotest.tools/v3/assert"
)

var (
	ctx = context.TODO()
	l   = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestEmissionCompatator(t *testing.T) {
	t.Run("Nil safety", func(t *testing.T) {
		a := RegionRecommendation{
			Emissions: nil,
		}
		b := RegionRecommendation{
			Emissions: nil,
		}
		assert.Equal(t, emissionComparator(a, b), 0)

		b.Emissions = nil
		a.Emissions = &provider.RegionalEmission{}
		assert.Equal(t, emissionComparator(a, b), 0)

		a.Emissions = nil
		b.Emissions = &provider.RegionalEmission{}
		assert.Equal(t, emissionComparator(a, b), 0)
	})

	t.Logf("Goal is to make `a` (lower direct, lower LCA, higher renewable, higher low-co2) better than `b`")

	t.Run("Direct Co2 diff", func(tt *testing.T) {
		a := RegionRecommendation{Emissions: &provider.RegionalEmission{}}
		b := RegionRecommendation{Emissions: &provider.RegionalEmission{}}

		a.Emissions.DirectCarbonIntensity = 100
		b.Emissions.DirectCarbonIntensity = 200
		assert.Equal(tt, emissionComparator(a, b), -1, "a should be less than b")

		a.Emissions.DirectCarbonIntensity = 200
		b.Emissions.DirectCarbonIntensity = 100
		assert.Equal(tt, emissionComparator(a, b), 1, "a should be greater than b")
	})

	t.Run("Renewable percentage diff", func(tt *testing.T) {
		a := RegionRecommendation{Emissions: &provider.RegionalEmission{
			DirectCarbonIntensity: 100,
		}}
		b := RegionRecommendation{Emissions: &provider.RegionalEmission{
			DirectCarbonIntensity: 100,
		}}

		a.Emissions.RenewablePercentage = 100
		b.Emissions.RenewablePercentage = 200
		assert.Equal(tt, emissionComparator(a, b), 1, "a should be lower than b")

		a.Emissions.RenewablePercentage = 200
		b.Emissions.RenewablePercentage = 100
		assert.Equal(tt, emissionComparator(a, b), -1, "a should be grater than b")
	})

	t.Run("Low carbon percentage diff", func(tt *testing.T) {
		a := RegionRecommendation{Emissions: &provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
		}}
		b := RegionRecommendation{Emissions: &provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
		}}

		a.Emissions.LowCarbonPercentage = 100
		b.Emissions.LowCarbonPercentage = 200
		assert.Equal(tt, emissionComparator(a, b), 1, "a should be lower than b")

		a.Emissions.LowCarbonPercentage = 200
		b.Emissions.LowCarbonPercentage = 100
		assert.Equal(tt, emissionComparator(a, b), -1, "a should be grater than b")
	})

	t.Run("LCA carbon intensity diff", func(tt *testing.T) {
		a := RegionRecommendation{Emissions: &provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
			LowCarbonPercentage:   50,
		}}
		b := RegionRecommendation{Emissions: &provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
			LowCarbonPercentage:   50,
		}}

		a.Emissions.LCACarbonIntensity = 100
		b.Emissions.LCACarbonIntensity = 200
		assert.Equal(tt, emissionComparator(a, b), -1, "a should be lower than b")

		a.Emissions.LCACarbonIntensity = 200
		b.Emissions.LCACarbonIntensity = 100
		assert.Equal(tt, emissionComparator(a, b), 1, "a should be grater than b")
	})

	t.Run("Should Right Go to Left ??", func(tt *testing.T) {
		a := RegionRecommendation{
			Emissions: &provider.RegionalEmission{
				DirectCarbonIntensity: 78.98,
				RenewablePercentage:   77.22,
				LowCarbonPercentage:   86.73,
				LCACarbonIntensity:    108.64,
			},
		}
		b := RegionRecommendation{
			Emissions: &provider.RegionalEmission{
				DirectCarbonIntensity: 678.58,
				RenewablePercentage:   6.1,
				LowCarbonPercentage:   9.4,
				LCACarbonIntensity:    736.62,
			},
		}

		assert.Equal(tt, emissionComparator(a, b), -1, "order is a < b")

		a, b = b, a
		assert.Equal(tt, emissionComparator(a, b), +1, "b should come before a")
	})
}

func TestIsEmissionLower(t *testing.T) {
	t.Run("a has lower emissions (direct)", func(tt *testing.T) {
		a := provider.RegionalEmission{}
		b := provider.RegionalEmission{}

		a.DirectCarbonIntensity = 100
		b.DirectCarbonIntensity = 200
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), true, "a should be less than b")

		a.DirectCarbonIntensity = 200
		b.DirectCarbonIntensity = 100
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), false, "a should be greater than b")
	})

	t.Run("a has lower emissions (renewable)", func(tt *testing.T) {
		a := provider.RegionalEmission{
			DirectCarbonIntensity: 100,
		}
		b := provider.RegionalEmission{
			DirectCarbonIntensity: 100,
		}

		a.RenewablePercentage = 200
		b.RenewablePercentage = 100
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), true, "a should be less than b")

		a.RenewablePercentage = 100
		b.RenewablePercentage = 200
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), false, "a should be greater than b")
	})

	t.Run("a has lower emissions (low carbon)", func(tt *testing.T) {
		a := provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
		}
		b := provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
		}

		a.LowCarbonPercentage = 200
		b.LowCarbonPercentage = 100
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), true, "a should be less than b")

		a.LowCarbonPercentage = 100
		b.LowCarbonPercentage = 200
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), false, "a should be greater than b")
	})

	t.Run("a has lower emissions (LCA)", func(tt *testing.T) {
		a := provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
			LowCarbonPercentage:   50,
		}
		b := provider.RegionalEmission{
			DirectCarbonIntensity: 100,
			RenewablePercentage:   50,
			LowCarbonPercentage:   50,
		}

		a.LCACarbonIntensity = 100
		b.LCACarbonIntensity = 200
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), true, "a should be less than b")

		a.LCACarbonIntensity = 200
		b.LCACarbonIntensity = 100
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), false, "a should be greater than b")
	})

	t.Run("a has lower emissions (all)", func(tt *testing.T) {
		a := provider.RegionalEmission{
			DirectCarbonIntensity: 78.98,
			RenewablePercentage:   77.22,
			LowCarbonPercentage:   86.73,
			LCACarbonIntensity:    108.64,
		}
		b := provider.RegionalEmission{
			DirectCarbonIntensity: 678.58,
			RenewablePercentage:   6.1,
			LowCarbonPercentage:   9.4,
			LCACarbonIntensity:    736.62,
		}

		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), true, "a should be less than b")

		a, b = b, a
		assert.Equal(tt, isA_sEmissionLowerOrEqual(a, b), false, "a should be greater than b")
	})
}

func TestCostPlusEmissionDecision(t *testing.T) {
	t.Run("Case when cost_r > cost_R", func(tt *testing.T) {
		cost_r := 100.0
		cost_R := 50.0
		assert.Equal(tt, costPlusEmissionDecision(cost_r, cost_R, nil, nil), 2, "cost_r should be greater than cost_R")
	})

	t.Run("Case Reject regions with higher emissions regardless of cost_r < or = cost_R", func(tt *testing.T) {

		cost_r := 10.0
		cost_R := 50.0

		tt.Logf("Goal: we remove the items which are in cost_r < cost_R but higher emissions")
		e_r := &provider.RegionalEmission{
			DirectCarbonIntensity: 101.0,
		}
		e_R := &provider.RegionalEmission{
			DirectCarbonIntensity: 100.0,
		}

		assert.Equal(tt, costPlusEmissionDecision(cost_r, cost_R, e_r, e_R), -2, "cost_r should be less than cost_R")

		tt.Logf("Goal: we remove the items which are in cost_r == cost_R but higher emissions")
		cost_r, cost_R = 5.0, 5.0

		assert.Equal(tt, costPlusEmissionDecision(cost_r, cost_R, e_r, e_R), -2, "cost_r == cost_R and e_r > e_R")
	})

	t.Run("Case when cost_r == cost_R and emission of either r is missing", func(tt *testing.T) {
		assert.Equal(tt, costPlusEmissionDecision(5.0, 5.0, nil, &provider.RegionalEmission{}), 0)
	})

	t.Run("Case when cost_r == cost_R and emission of either R is missing", func(tt *testing.T) {
		assert.Equal(tt, costPlusEmissionDecision(5.0, 5.0, &provider.RegionalEmission{}, nil), 1)
	})

	t.Run("Case when cost_r < cost_R and emission of r <= R", func(tt *testing.T) {
		cost_r, cost_R := 5.0, 6.0
		e_r := &provider.RegionalEmission{
			DirectCarbonIntensity: 100.0,
		}
		e_R := &provider.RegionalEmission{
			DirectCarbonIntensity: 101.0,
		}
		assert.Equal(tt, costPlusEmissionDecision(cost_r, cost_R, e_r, e_R), -1, "cost_r should be less than cost_R")
	})

	t.Run("Case when cost_r == cost_R and emission of r <= R", func(tt *testing.T) {
		cost_r, cost_R := 5.0, 5.0
		e_r := &provider.RegionalEmission{
			DirectCarbonIntensity: 100.0,
		}
		e_R := &provider.RegionalEmission{
			DirectCarbonIntensity: 101.0,
		}
		assert.Equal(tt, costPlusEmissionDecision(cost_r, cost_R, e_r, e_R), 1, "cost_r should be less than cost_R")
	})
}

func TestInstanceTypeOptimizerAcrossRegionsSelfManaged(t *testing.T) {
	clusterType := consts.ClusterTypeSelfMang

	cpSku := "cpSku"
	wpSku := "wpSku"
	lbSku := "lbSku"
	etcdSku := "etcdSku"

	// we need to validate the no in the function
	noCP := 3
	noWP := 4
	noDS := 5

	t.Run("2 regions have diff costs no recommendation", func(t *testing.T) {
		costsSelfManaged := []RecommendationSelfManagedCost{
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region2",
				CpCost:    150.0,
				WpCost:    250.0,
				EtcdCost:  350.0,
				LbCost:    30.0,
				TotalCost: 780.0,
			},
		}
		regions := map[string]provider.RegionOutput{
			"region1": {
				Sku:  "region1",
				Name: "region1",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 150.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region2": {
				Sku:  "region2",
				Name: "region2",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 100.0,
					Unit:                  "gCO2/kWh",
				},
			},
		}

		currReg := "region1"

		expectedResp := RecommendationAcrossRegions{
			CurrentRegion:         currReg,
			CurrentEmissions:      regions[currReg].Emission,
			CurrentTotalCost:      100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
			InstanceTypeCP:        cpSku,
			InstanceTypeWP:        wpSku,
			InstanceTypeDS:        etcdSku,
			InstanceTypeLB:        lbSku,
			ControlPlaneCount:     noCP,
			WorkerPlaneCount:      noWP,
			DataStoreCount:        noDS,
			RegionRecommendations: nil,
		}

		actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
			regions,
			clusterType,
			nil,
			costsSelfManaged,
			currReg,
			noCP,
			noWP,
			noDS,
			"",
			cpSku,
			wpSku,
			etcdSku,
			lbSku,
		)

		assert.NilError(t, err, "error should be nil")
		assert.DeepEqual(t, expectedResp, actualResp)
	})

	t.Run("2 regions have diff costs with recommendation", func(t *testing.T) {
		costsSelfManaged := []RecommendationSelfManagedCost{
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region2",
				CpCost:    150.0,
				WpCost:    250.0,
				EtcdCost:  350.0,
				LbCost:    30.0,
				TotalCost: 780.0,
			},
		}
		regions := map[string]provider.RegionOutput{
			"region1": {
				Sku:  "region1",
				Name: "region1",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 50.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region2": {
				Sku:  "region2",
				Name: "region2",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 100.0,
					Unit:                  "gCO2/kWh",
				},
			},
		}

		currReg := "region2"

		expectedResp := RecommendationAcrossRegions{
			CurrentRegion:     currReg,
			CurrentEmissions:  regions[currReg].Emission,
			CurrentTotalCost:  150.0*float64(noCP) + 250.0*float64(noWP) + 350.0*float64(noDS) + 30.0,
			InstanceTypeCP:    cpSku,
			InstanceTypeWP:    wpSku,
			InstanceTypeDS:    etcdSku,
			InstanceTypeLB:    lbSku,
			ControlPlaneCount: noCP,
			WorkerPlaneCount:  noWP,
			DataStoreCount:    noDS,
			RegionRecommendations: []RegionRecommendation{
				{
					Region:           "region1",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  200.0,
					DataStoreCost:    300.0,
					LoadBalancerCost: 20.0,
					TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
					Emissions:        regions["region1"].Emission,
				},
			},
		}

		actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
			regions,
			clusterType,
			nil,
			costsSelfManaged,
			currReg,
			noCP,
			noWP,
			noDS,
			"",
			cpSku,
			wpSku,
			etcdSku,
			lbSku,
		)

		assert.NilError(t, err, "error should be nil")
		assert.DeepEqual(t, expectedResp, actualResp)
	})

	t.Run("2 regions have same costs recommendation", func(tI *testing.T) {
		costsSelfManaged := []RecommendationSelfManagedCost{
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region2",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region3",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region4",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region5",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
		}

		tI.Run("different emissions (some are missing with current missing)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku:      "region2",
					Emission: nil,
				},
				"region3": {
					Sku:      "region3",
					Emission: nil,
				},
				"region4": {
					Sku: "region4",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := RecommendationAcrossRegions{
				CurrentRegion:     currReg,
				CurrentEmissions:  regions[currReg].Emission,
				CurrentTotalCost:  100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
				InstanceTypeCP:    cpSku,
				InstanceTypeWP:    wpSku,
				InstanceTypeDS:    etcdSku,
				InstanceTypeLB:    lbSku,
				ControlPlaneCount: noCP,
				WorkerPlaneCount:  noWP,
				DataStoreCount:    noDS,
				RegionRecommendations: []RegionRecommendation{
					{
						Region:           "region1",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region1"].Emission,
					},
					{
						Region:           "region4",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region4"].Emission,
					},
					{
						Region:           "region5",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region5"].Emission,
					},
					{
						Region:           "region3",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region3"].Emission,
					},
				},
			}

			actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
				regions,
				clusterType,
				nil,
				costsSelfManaged,
				currReg,
				noCP,
				noWP,
				noDS,
				"", cpSku,
				wpSku, etcdSku, lbSku,
			)

			assert.NilError(tII, err, "error should be nil")
			assert.DeepEqual(tII, expectedResp, actualResp)
		})

		tI.Run("different emissions (some are missing with current not missing with some higher emissions)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 250.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku: "region2",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 100.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region3": {
					Sku:      "region3",
					Emission: nil,
				},
				"region4": {
					Sku: "region4",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 150.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := RecommendationAcrossRegions{
				CurrentRegion:     currReg,
				CurrentEmissions:  regions[currReg].Emission,
				CurrentTotalCost:  100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
				InstanceTypeCP:    cpSku,
				InstanceTypeWP:    wpSku,
				InstanceTypeDS:    etcdSku,
				InstanceTypeLB:    lbSku,
				ControlPlaneCount: noCP,
				WorkerPlaneCount:  noWP,
				DataStoreCount:    noDS,
				RegionRecommendations: []RegionRecommendation{
					{
						Region:           "region4",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region4"].Emission,
					},
					{
						Region:           "region3",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region3"].Emission,
					},
				},
			}

			actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
				regions,
				clusterType,
				nil,
				costsSelfManaged,
				currReg,
				noCP,
				noWP,
				noDS,
				"", cpSku,
				wpSku, etcdSku, lbSku,
			)

			assert.NilError(tII, err, "error should be nil")
			assert.DeepEqual(tII, expectedResp, actualResp)
		})
	})

	t.Run("regions have same costs and different costs", func(tI *testing.T) {
		costsSelfManaged := []RecommendationSelfManagedCost{
			{
				Region:    "region6",
				CpCost:    300.0,
				WpCost:    200.0,
				EtcdCost:  400.0,
				LbCost:    20.0,
				TotalCost: 920.0,
			},
			{
				Region:    "region10",
				CpCost:    301.0,
				WpCost:    200.0,
				EtcdCost:  400.0,
				LbCost:    20.0,
				TotalCost: 1020.0,
			},
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region2",
				CpCost:    100.0,
				WpCost:    350.0,
				EtcdCost:  350.0,
				LbCost:    20.0,
				TotalCost: 820.0,
			},
			{
				Region:    "region3",
				CpCost:    200.0,
				WpCost:    200.0,
				EtcdCost:  400.0,
				LbCost:    120.0,
				TotalCost: 920.0,
			},
			{
				Region:    "region4",
				CpCost:    110.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 630.0,
			},
			{
				Region:    "region5",
				CpCost:    100.0,
				WpCost:    200.0,
				EtcdCost:  300.0,
				LbCost:    20.0,
				TotalCost: 620.0,
			},
		}

		currReg := "region6"

		regions := map[string]provider.RegionOutput{
			"region1": {
				Sku: "region1",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 200.0,
					RenewablePercentage:   50.0,
					LowCarbonPercentage:   14.5,
					LCACarbonIntensity:    10.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region2": {
				Sku: "region2",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 120.0,
					RenewablePercentage:   55.0,
					LowCarbonPercentage:   20.0,
					LCACarbonIntensity:    5,
					Unit:                  "gCO2/kWh",
				},
			},
			"region3": {
				Sku: "region3",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 230,
					RenewablePercentage:   60,
					LowCarbonPercentage:   15,
					LCACarbonIntensity:    40,
					Unit:                  "gCO2/kWh",
				},
			},
			"region4": {
				Sku:      "region4",
				Emission: nil,
			},
			"region5": {
				Sku: "region5",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 200,
					RenewablePercentage:   51,
					LowCarbonPercentage:   15,
					LCACarbonIntensity:    5,
					Unit:                  "gCO2/kWh",
				},
			},
			"region6": {
				Sku: "region6",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 240.0,
					RenewablePercentage:   50.0,
					LowCarbonPercentage:   14.5,
					LCACarbonIntensity:    50.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region10": {
				Sku:      "region10",
				Emission: nil,
			},
		}

		expectedResp := RecommendationAcrossRegions{
			CurrentRegion:     currReg,
			CurrentEmissions:  regions[currReg].Emission,
			CurrentTotalCost:  300.0*float64(noCP) + 200.0*float64(noWP) + 400.0*float64(noDS) + 20.0,
			InstanceTypeCP:    cpSku,
			InstanceTypeWP:    wpSku,
			InstanceTypeDS:    etcdSku,
			InstanceTypeLB:    lbSku,
			ControlPlaneCount: noCP,
			WorkerPlaneCount:  noWP,
			DataStoreCount:    noDS,
			RegionRecommendations: []RegionRecommendation{
				{
					Region:           "region5",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  200.0,
					DataStoreCost:    300.0,
					LoadBalancerCost: 20.0,
					TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
					Emissions:        regions["region5"].Emission,
				},
				{
					Region:           "region1",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  200.0,
					DataStoreCost:    300.0,
					LoadBalancerCost: 20.0,
					TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
					Emissions:        regions["region1"].Emission,
				},
				{
					Region:           "region4",
					ControlPlaneCost: 110.0,
					WorkerPlaneCost:  200.0,
					DataStoreCost:    300.0,
					LoadBalancerCost: 20.0,
					TotalCost:        110.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
					Emissions:        regions["region4"].Emission,
				},
				{
					Region:           "region2",
					ControlPlaneCost: 100,
					WorkerPlaneCost:  350.0,
					DataStoreCost:    350.0,
					LoadBalancerCost: 20.0,
					TotalCost:        100.0*float64(noCP) + 350.0*float64(noWP) + 350.0*float64(noDS) + 20.0,
					Emissions:        regions["region2"].Emission,
				},
				{
					Region:           "region3",
					ControlPlaneCost: 200,
					WorkerPlaneCost:  200.0,
					DataStoreCost:    400.0,
					LoadBalancerCost: 120.0,
					TotalCost:        200.0*float64(noCP) + 200.0*float64(noWP) + 400.0*float64(noDS) + 120.0,
					Emissions:        regions["region3"].Emission,
				},
			},
		}

		actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
			regions,
			clusterType,
			nil,
			costsSelfManaged,
			currReg,
			noCP,
			noWP,
			noDS,
			"", cpSku,
			wpSku, etcdSku, lbSku,
		)

		assert.NilError(tI, err, "error should be nil")
		assert.DeepEqual(tI, expectedResp, actualResp)

	})
}

func TestInstanceTypeOptimizerAcrossRegionsManaged(t *testing.T) {
	clusterType := consts.ClusterTypeMang

	managedOfferSku := "managedOfferSku"
	wpSku := "wpSku"

	// we need to validate the no in the function
	noWP := 4
	noCP := 0
	noDS := 0

	t.Run("2 regions have diff costs no recommendation", func(t *testing.T) {
		costsManaged := []RecommendationManagedCost{
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
			{
				Region:    "region2",
				CpCost:    150.0,
				WpCost:    250.0,
				TotalCost: 400.0,
			},
		}

		regions := map[string]provider.RegionOutput{
			"region1": {
				Sku:  "region1",
				Name: "region1",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 150.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region2": {
				Sku:  "region2",
				Name: "region2",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 100.0,
					Unit:                  "gCO2/kWh",
				},
			},
		}

		currReg := "region1"

		expectedResp := RecommendationAcrossRegions{
			CurrentRegion:         currReg,
			CurrentEmissions:      regions[currReg].Emission,
			CurrentTotalCost:      100.0 + 200.0*float64(noWP),
			ManagedOffering:       managedOfferSku,
			InstanceTypeWP:        wpSku,
			WorkerPlaneCount:      noWP,
			RegionRecommendations: nil,
		}

		actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
			regions,
			clusterType,
			costsManaged,
			nil,
			currReg,
			noCP,
			noWP,
			noDS,
			managedOfferSku,
			"",
			wpSku,
			"", "",
		)

		assert.NilError(t, err, "error should be nil")
		assert.DeepEqual(t, expectedResp, actualResp)
	})

	t.Run("2 regions have diff costs with recommendation", func(t *testing.T) {
		costsManaged := []RecommendationManagedCost{
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
			{
				Region:    "region2",
				CpCost:    150.0,
				WpCost:    250.0,
				TotalCost: 400.0,
			},
		}
		regions := map[string]provider.RegionOutput{
			"region1": {
				Sku:  "region1",
				Name: "region1",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 50.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region2": {
				Sku:  "region2",
				Name: "region2",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 100.0,
					Unit:                  "gCO2/kWh",
				},
			},
		}

		currReg := "region2"

		expectedResp := RecommendationAcrossRegions{
			CurrentRegion:    currReg,
			CurrentEmissions: regions[currReg].Emission,
			CurrentTotalCost: 150.0 + 250.0*float64(noWP),
			ManagedOffering:  managedOfferSku,
			InstanceTypeWP:   wpSku,
			WorkerPlaneCount: noWP,
			RegionRecommendations: []RegionRecommendation{
				{
					Region:           "region1",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  200.0,
					TotalCost:        100.0 + 200.0*float64(noWP),
					Emissions:        regions["region1"].Emission,
				},
			},
		}

		actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
			regions,
			clusterType,
			costsManaged,
			nil,
			currReg,
			noCP,
			noWP,
			noDS,
			managedOfferSku,
			"",
			wpSku, "", "",
		)

		assert.NilError(t, err, "error should be nil")
		assert.DeepEqual(t, expectedResp, actualResp)
	})

	t.Run("2 regions have same costs recommendation", func(tI *testing.T) {
		costsManaged := []RecommendationManagedCost{
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
			{
				Region:    "region2",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
			{
				Region:    "region3",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
			{
				Region:    "region4",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
			{
				Region:    "region5",
				CpCost:    100.0,
				WpCost:    200.0,
				TotalCost: 300.0,
			},
		}

		tI.Run("different emissions (some are missing with current missing)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku:      "region2",
					Emission: nil,
				},
				"region3": {
					Sku:      "region3",
					Emission: nil,
				},
				"region4": {
					Sku: "region4",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := RecommendationAcrossRegions{
				CurrentRegion:    currReg,
				CurrentEmissions: regions[currReg].Emission,
				CurrentTotalCost: 100.0 + 200.0*float64(noWP),
				ManagedOffering:  managedOfferSku,
				InstanceTypeWP:   wpSku,
				WorkerPlaneCount: noWP,
				RegionRecommendations: []RegionRecommendation{
					{
						Region:           "region1",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region1"].Emission,
					},
					{
						Region:           "region4",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region4"].Emission,
					},
					{
						Region:           "region5",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region5"].Emission,
					},
					{
						Region:           "region3",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region3"].Emission,
					},
				},
			}

			actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
				regions,
				clusterType,
				costsManaged,
				nil,
				currReg,
				noCP,
				noWP,
				noDS,
				managedOfferSku,
				"",
				wpSku, "", "",
			)

			assert.NilError(tII, err, "error should be nil")
			assert.DeepEqual(tII, expectedResp, actualResp)
		})

		tI.Run("different emissions (some are missing with current not missing)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 250.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku: "region2",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 100.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region3": {
					Sku:      "region3",
					Emission: nil,
				},
				"region4": {
					Sku: "region4",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 150.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := RecommendationAcrossRegions{
				CurrentRegion:    currReg,
				CurrentEmissions: regions[currReg].Emission,
				CurrentTotalCost: 100.0 + 200.0*float64(noWP),
				ManagedOffering:  managedOfferSku,
				InstanceTypeWP:   wpSku,
				WorkerPlaneCount: noWP,
				RegionRecommendations: []RegionRecommendation{
					{
						Region:           "region4",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region4"].Emission,
					},
					{
						Region:           "region3",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region3"].Emission,
					},
				},
			}

			actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
				regions,
				clusterType,
				costsManaged,
				nil,
				currReg,
				noCP,
				noWP,
				noDS,
				managedOfferSku,
				"",
				wpSku, "", "",
			)

			assert.NilError(tII, err, "error should be nil")
			assert.DeepEqual(tII, expectedResp, actualResp)
		})
	})

	t.Run("regions have same costs and different costs", func(tI *testing.T) {
		costsManaged := []RecommendationManagedCost{
			{
				Region:    "region6",
				CpCost:    100.0,
				WpCost:    820.0,
				TotalCost: 920.0,
			},
			{
				Region:    "region10",
				CpCost:    101.0,
				WpCost:    920.0,
				TotalCost: 1021.0,
			},
			{
				Region:    "region1",
				CpCost:    100.0,
				WpCost:    520.0,
				TotalCost: 620.0,
			},
			{
				Region:    "region2",
				CpCost:    100.0,
				WpCost:    720.0,
				TotalCost: 820.0,
			},
			{
				Region:    "region3",
				CpCost:    200.0,
				WpCost:    620.0,
				TotalCost: 920.0,
			},
			{
				Region:    "region4",
				CpCost:    110.0,
				WpCost:    520.0,
				TotalCost: 630.0,
			},
			{
				Region:    "region5",
				CpCost:    100.0,
				WpCost:    520.0,
				TotalCost: 620.0,
			},
		}

		currReg := "region6"

		regions := map[string]provider.RegionOutput{
			"region1": {
				Sku: "region1",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 200.0,
					RenewablePercentage:   50.0,
					LowCarbonPercentage:   14.5,
					LCACarbonIntensity:    10.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region2": {
				Sku: "region2",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 120.0,
					RenewablePercentage:   55.0,
					LowCarbonPercentage:   20.0,
					LCACarbonIntensity:    5,
					Unit:                  "gCO2/kWh",
				},
			},
			"region3": {
				Sku: "region3",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 230,
					RenewablePercentage:   60,
					LowCarbonPercentage:   15,
					LCACarbonIntensity:    40,
					Unit:                  "gCO2/kWh",
				},
			},
			"region4": {
				Sku:      "region4",
				Emission: nil,
			},
			"region5": {
				Sku: "region5",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 200,
					RenewablePercentage:   51,
					LowCarbonPercentage:   15,
					LCACarbonIntensity:    5,
					Unit:                  "gCO2/kWh",
				},
			},
			"region6": {
				Sku: "region6",
				Emission: &provider.RegionalEmission{
					DirectCarbonIntensity: 240.0,
					RenewablePercentage:   50.0,
					LowCarbonPercentage:   14.5,
					LCACarbonIntensity:    50.0,
					Unit:                  "gCO2/kWh",
				},
			},
			"region10": {
				Sku:      "region10",
				Emission: nil,
			},
		}

		expectedResp := RecommendationAcrossRegions{
			CurrentRegion:    currReg,
			CurrentEmissions: regions[currReg].Emission,
			CurrentTotalCost: 100.0 + 820*float64(noWP),
			ManagedOffering:  managedOfferSku,
			InstanceTypeWP:   wpSku,
			WorkerPlaneCount: noWP,
			RegionRecommendations: []RegionRecommendation{
				{
					Region:           "region5",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  520.0,
					TotalCost:        100.0 + 520.0*float64(noWP),
					Emissions:        regions["region5"].Emission,
				},
				{
					Region:           "region1",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  520.0,
					TotalCost:        100.0 + 520.0*float64(noWP),
					Emissions:        regions["region1"].Emission,
				},
				{
					Region:           "region4",
					ControlPlaneCost: 110.0,
					WorkerPlaneCost:  520.0,
					TotalCost:        110.0 + 520.0*float64(noWP),
					Emissions:        regions["region4"].Emission,
				},
				{
					Region:           "region3",
					ControlPlaneCost: 200,
					WorkerPlaneCost:  620.0,
					TotalCost:        200.0 + 620.0*float64(noWP),
					Emissions:        regions["region3"].Emission,
				},
				{
					Region:           "region2",
					ControlPlaneCost: 100,
					WorkerPlaneCost:  720.0,
					TotalCost:        100.0 + 720.0*float64(noWP),
					Emissions:        regions["region2"].Emission,
				},
			},
		}

		actualResp, err := NewOptimizer(ctx, l, nil).InstanceTypeOptimizerAcrossRegions(
			regions,
			clusterType,
			costsManaged, nil,
			currReg,
			noCP,
			noWP,
			noDS,
			managedOfferSku,
			"",
			wpSku, "", "",
		)

		assert.NilError(tI, err, "error should be nil")
		assert.DeepEqual(tI, expectedResp, actualResp)

	})
}
