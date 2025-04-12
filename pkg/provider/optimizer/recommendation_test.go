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

package optimizer_test

import (
	"context"
	"os"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/provider/optimizer"
	"gotest.tools/v3/assert"
)

var (
	ctx = context.TODO()
	l   = logger.NewStructuredLogger(-1, os.Stdout)
	opt = optimizer.NewOptimizer(ctx, l, nil)
)

func TestInstanceTypeOptimizerAcrossRegionsSelfManaged(t *testing.T) {
	clusterType := consts.ClusterTypeSelfMang

	regions := map[string]provider.RegionOutput{
		"region1": {
			Sku:  "region1",
			Name: "region1",
			Emission: &provider.RegionalEmission{
				DirectCarbonIntensity: 50.0,
				RenewablePercentage:   74.4,
				LowCarbonPercentage:   5.5,
				Unit:                  "gCO2/kWh",
			},
		},
		"region2": {
			Sku:  "region2",
			Name: "region2",
			Emission: &provider.RegionalEmission{
				DirectCarbonIntensity: 100.0,
				RenewablePercentage:   94.4,
				LowCarbonPercentage:   9.5,
				Unit:                  "gCO2/kWh",
			},
		},
	}
	cpSku := "cpSku"
	wpSku := "wpSku"
	lbSku := "lbSku"
	etcdSku := "etcdSku"

	// we need to validate the no in the function
	noCP := 3
	noWP := 4
	noDS := 5

	t.Run("2 regions have diff costs no recommendation", func(t *testing.T) {
		costsSelfManaged := []optimizer.RecommendationSelfManagedCost{
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

		currReg := "region1"

		expectedResp := optimizer.RecommendationAcrossRegions{
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

		actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
		costsSelfManaged := []optimizer.RecommendationSelfManagedCost{
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

		currReg := "region2"

		expectedResp := optimizer.RecommendationAcrossRegions{
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
			RegionRecommendations: []optimizer.RegionRecommendation{
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

		actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
		costsSelfManaged := []optimizer.RecommendationSelfManagedCost{
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

		tI.Run("different emissions (all depths)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   15.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku: "region2",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 100.0,
						RenewablePercentage:   7.4,
						LowCarbonPercentage:   9.5,
						LCACarbonIntensity:    100.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region3": {
					Sku: "region3",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   15.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region4": {
					Sku: "region4",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   14.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   14.5,
						LCACarbonIntensity:    50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := optimizer.RecommendationAcrossRegions{
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
				RegionRecommendations: []optimizer.RegionRecommendation{
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
						Region:           "region3",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region3"].Emission,
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
						Region:           "region4",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region4"].Emission,
					},
				},
			}

			actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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

		tI.Run("different emissions (some are missing with current missing)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   5.5,
						LCACarbonIntensity:    60.0,
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
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   4.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   4.5,
						LCACarbonIntensity:    50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := optimizer.RecommendationAcrossRegions{
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
				RegionRecommendations: []optimizer.RegionRecommendation{
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
						Region:           "region5",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						DataStoreCost:    300.0,
						LoadBalancerCost: 20.0,
						TotalCost:        100.0*float64(noCP) + 200.0*float64(noWP) + 300.0*float64(noDS) + 20.0,
						Emissions:        regions["region5"].Emission,
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

			actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   5.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku: "region2",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 100.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   9.5,
						LCACarbonIntensity:    100.0,
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
						RenewablePercentage:   75.4,
						LowCarbonPercentage:   14.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   4.5,
						LCACarbonIntensity:    50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := optimizer.RecommendationAcrossRegions{
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
				RegionRecommendations: []optimizer.RegionRecommendation{
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

			actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
}

func TestInstanceTypeOptimizerAcrossRegionsManaged(t *testing.T) {
	clusterType := consts.ClusterTypeMang

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
	managedOfferSku := "managedOfferSku"
	wpSku := "wpSku"

	// we need to validate the no in the function
	noWP := 4
	noCP := 0
	noDS := 0

	t.Run("2 regions have diff costs no recommendation", func(t *testing.T) {
		costsManaged := []optimizer.RecommendationManagedCost{
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

		currReg := "region1"

		expectedResp := optimizer.RecommendationAcrossRegions{
			CurrentRegion:         currReg,
			CurrentEmissions:      regions[currReg].Emission,
			CurrentTotalCost:      100.0 + 200.0*float64(noWP),
			ManagedOffering:       managedOfferSku,
			InstanceTypeWP:        wpSku,
			WorkerPlaneCount:      noWP,
			RegionRecommendations: nil,
		}

		actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
		costsManaged := []optimizer.RecommendationManagedCost{
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

		currReg := "region2"

		expectedResp := optimizer.RecommendationAcrossRegions{
			CurrentRegion:    currReg,
			CurrentEmissions: regions[currReg].Emission,
			CurrentTotalCost: 150.0 + 250.0*float64(noWP),
			ManagedOffering:  managedOfferSku,
			InstanceTypeWP:   wpSku,
			WorkerPlaneCount: noWP,
			RegionRecommendations: []optimizer.RegionRecommendation{
				{
					Region:           "region1",
					ControlPlaneCost: 100.0,
					WorkerPlaneCost:  200.0,
					TotalCost:        100.0 + 200.0*float64(noWP),
					Emissions:        regions["region1"].Emission,
				},
			},
		}

		actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
		costsManaged := []optimizer.RecommendationManagedCost{
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

		tI.Run("different emissions (all depths)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   15.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku: "region2",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 100.0,
						RenewablePercentage:   7.4,
						LowCarbonPercentage:   9.5,
						LCACarbonIntensity:    100.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region3": {
					Sku: "region3",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   15.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region4": {
					Sku: "region4",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   14.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   14.5,
						LCACarbonIntensity:    50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := optimizer.RecommendationAcrossRegions{
				CurrentRegion:    currReg,
				CurrentEmissions: regions[currReg].Emission,
				CurrentTotalCost: 100.0 + 200.0*float64(noWP),
				ManagedOffering:  managedOfferSku,
				InstanceTypeWP:   wpSku,
				WorkerPlaneCount: noWP,
				RegionRecommendations: []optimizer.RegionRecommendation{
					{
						Region:           "region1",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region1"].Emission,
					},
					{
						Region:           "region3",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region3"].Emission,
					},
					{
						Region:           "region5",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region5"].Emission,
					},
					{
						Region:           "region4",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region4"].Emission,
					},
				},
			}

			actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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

		tI.Run("different emissions (some are missing with current missing)", func(tII *testing.T) {
			currReg := "region2"

			regions := map[string]provider.RegionOutput{
				"region1": {
					Sku: "region1",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   5.5,
						LCACarbonIntensity:    60.0,
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
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   4.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   4.5,
						LCACarbonIntensity:    50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := optimizer.RecommendationAcrossRegions{
				CurrentRegion:    currReg,
				CurrentEmissions: regions[currReg].Emission,
				CurrentTotalCost: 100.0 + 200.0*float64(noWP),
				ManagedOffering:  managedOfferSku,
				InstanceTypeWP:   wpSku,
				WorkerPlaneCount: noWP,
				RegionRecommendations: []optimizer.RegionRecommendation{
					{
						Region:           "region1",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region1"].Emission,
					},
					{
						Region:           "region5",
						ControlPlaneCost: 100.0,
						WorkerPlaneCost:  200.0,
						TotalCost:        100.0 + 200.0*float64(noWP),
						Emissions:        regions["region5"].Emission,
					},
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

			actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   5.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region2": {
					Sku: "region2",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 100.0,
						RenewablePercentage:   74.4,
						LowCarbonPercentage:   9.5,
						LCACarbonIntensity:    100.0,
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
						RenewablePercentage:   75.4,
						LowCarbonPercentage:   14.5,
						LCACarbonIntensity:    60.0,
						Unit:                  "gCO2/kWh",
					},
				},
				"region5": {
					Sku: "region5",
					Emission: &provider.RegionalEmission{
						DirectCarbonIntensity: 50.0,
						RenewablePercentage:   73.4,
						LowCarbonPercentage:   4.5,
						LCACarbonIntensity:    50.0,
						Unit:                  "gCO2/kWh",
					},
				},
			}

			expectedResp := optimizer.RecommendationAcrossRegions{
				CurrentRegion:    currReg,
				CurrentEmissions: regions[currReg].Emission,
				CurrentTotalCost: 100.0 + 200.0*float64(noWP),
				ManagedOffering:  managedOfferSku,
				InstanceTypeWP:   wpSku,
				WorkerPlaneCount: noWP,
				RegionRecommendations: []optimizer.RegionRecommendation{
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

			actualResp, err := opt.InstanceTypeOptimizerAcrossRegions(
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
}
