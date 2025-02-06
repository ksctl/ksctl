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

package metadata

import (
	"errors"
	"sort"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"golang.org/x/mod/semver"

	"github.com/ksctl/ksctl/v2/pkg/provider"
)

func (kc *Controller) ListAllRegions() (
	_ []provider.RegionOutput,
	errC error,
) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.client) {
		return nil, nil
	}

	regions, err := kc.cc.GetAvailableRegions()
	if err != nil {
		return nil, err
	}

	return regions, nil
}

type PriceCalculatorInput struct {
	// NoOfWorkerNodes this is used for both managed as managedNodes and self managed as workerNodes
	NoOfWorkerNodes       int
	NoOfControlPlaneNodes int
	NoOfEtcdNodes         int
	ControlPlaneMachine   provider.InstanceRegionOutput

	// WorkerMachine this is used for both managed as managedNodes and self managed as workerNodes
	WorkerMachine       provider.InstanceRegionOutput
	EtcdMachine         provider.InstanceRegionOutput
	LoadBalancerMachine provider.InstanceRegionOutput

	ManagedControlPlaneMachine provider.ManagedClusterOutput
}

func (kc *Controller) PriceCalculator(inp PriceCalculatorInput) (float64, error) {
	if kc.client.Metadata.ClusterType == consts.ClusterTypeSelfMang {
		return kc.priceCalculatorForSelfManagedCluster(inp)
	} else {
		return kc.priceCalculatorForManagedCluster(inp)
	}
}

func (kc *Controller) priceCalculatorForSelfManagedCluster(inp PriceCalculatorInput) (float64, error) {
	workerCost := float64(inp.NoOfWorkerNodes) * inp.WorkerMachine.GetCost()
	controlPlaneCost := float64(inp.NoOfControlPlaneNodes) * inp.ControlPlaneMachine.GetCost()
	etcdCost := float64(inp.NoOfEtcdNodes) * inp.EtcdMachine.GetCost()
	lbCost := inp.LoadBalancerMachine.GetCost()

	return workerCost + controlPlaneCost + etcdCost + lbCost, nil
}

func (kc *Controller) priceCalculatorForManagedCluster(inp PriceCalculatorInput) (float64, error) {
	managedNodeCost := float64(inp.NoOfWorkerNodes) * inp.WorkerMachine.GetCost()

	return managedNodeCost + inp.ManagedControlPlaneMachine.GetCost(), nil
}

// ListAllManagedClusterManagementOfferings you can pass choosenInstanceType as nil if you are not using EKS AutoNode mode
func (kc *Controller) ListAllManagedClusterManagementOfferings(region string, choosenInstanceType *string) (
	out map[string]provider.ManagedClusterOutput,
	errC error,
) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.client) {
		return nil, nil
	}

	offerings, err := kc.cc.GetAvailableManagedK8sManagementOfferings(region, choosenInstanceType)
	if err != nil {
		return nil, err
	}

	out = make(map[string]provider.ManagedClusterOutput)
	for _, v := range offerings {
		out[v.Sku] = v
	}

	return out, nil
}

func (kc *Controller) ListAllInstances(region string) (
	out map[string]provider.InstanceRegionOutput,
	errC error,
) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.client) {
		return nil, nil
	}

	instances, err := kc.cc.GetAvailableInstanceTypes(region, kc.client.Metadata.ClusterType)
	if err != nil {
		return nil, err
	}

	out = make(map[string]provider.InstanceRegionOutput)
	for _, v := range instances {
		out[v.Sku] = v
	}

	return out, nil
}

func (kc *Controller) ListAllManagedClusterK8sVersions(region string) (_ []string, errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	vers, err := kc.cc.GetAvailableManagedK8sVersions(region)
	if err != nil {
		return nil, err
	}

	isRepoRespectSemver := true
	for i := range vers {
		if !semver.IsValid(vers[i]) {
			isRepoRespectSemver = false
			vers[i] = semver.Canonical("v" + vers[i]) // WARN: this is adding patch version to the version aka .0 to the end
		}
	}

	sort.Slice(vers, func(i, j int) bool {
		return semver.Compare(vers[i], vers[j]) > 0
	})

	tags := make([]string, 0, len(vers))

	for _, r := range vers {
		if !isRepoRespectSemver {
			r = strings.TrimPrefix(r, "v")
		}
		tags = append(tags, r)
	}

	return tags, nil
}

func (kc *Controller) ListAllEtcdVersions() (_ []string, errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	return kc.bb.GetAvailableEtcdVersions()
}

func (kc *Controller) ListAllBootstrapVersions() (_ []string, errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	return kc.bb.D.GetBootstrapedDistributionVersions()
}
