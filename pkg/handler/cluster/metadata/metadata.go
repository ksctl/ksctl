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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/provider/optimizer"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/bootstrap/handler/cni"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"golang.org/x/mod/semver"

	"github.com/ksctl/ksctl/v2/pkg/provider"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
)

func (kc *Controller) ListAllRegions() (
	_ provider.RegionsOutput,
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

	o := optimizer.NewOptimizer(kc.ctx, kc.l, regions)
	res, err := o.AttachEmissionsToRegions(kc.client.Metadata.Provider)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type PriceCalculatorInput struct {
	Currency string

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

func convertToHumanReadable(price float64, currency string) string {
	symbol := map[string]rune{
		"USD": '$',
		"EUR": '€',
		"INR": '₹',
	}

	sign := ""
	if v, ok := symbol[currency]; ok {
		sign = string(v)
	} else {
		sign = string(symbol["USD"])
	}

	return sign + strconv.FormatFloat(price, 'f', 2, 64)
}

func (kc *Controller) priceCalculatorForSelfManagedCluster(inp PriceCalculatorInput) (float64, error) {
	workerCost := float64(inp.NoOfWorkerNodes) * inp.WorkerMachine.GetCost()
	controlPlaneCost := float64(inp.NoOfControlPlaneNodes) * inp.ControlPlaneMachine.GetCost()
	etcdCost := float64(inp.NoOfEtcdNodes) * inp.EtcdMachine.GetCost()
	lbCost := inp.LoadBalancerMachine.GetCost()
	currency := inp.Currency

	total := workerCost + controlPlaneCost + etcdCost + lbCost

	headers := []string{"Resource", "UnitCost", "Quantity", "Cost"}
	rows := [][]string{
		{
			"Control Plane",
			convertToHumanReadable(inp.ControlPlaneMachine.GetCost(), currency),
			strconv.Itoa(inp.NoOfControlPlaneNodes),
			convertToHumanReadable(controlPlaneCost, currency),
		},
		{
			"Worker Node(s)",
			convertToHumanReadable(inp.WorkerMachine.GetCost(), currency),
			strconv.Itoa(inp.NoOfWorkerNodes),
			convertToHumanReadable(workerCost, currency),
		},
		{
			"Etcd Nodes",
			convertToHumanReadable(inp.EtcdMachine.GetCost(), currency),
			strconv.Itoa(inp.NoOfEtcdNodes),
			convertToHumanReadable(etcdCost, currency),
		},
		{
			"LoadBalancer Node",
			convertToHumanReadable(inp.LoadBalancerMachine.GetCost(), currency),
			"1",
			convertToHumanReadable(lbCost, currency),
		},
		{"", "", "", ""},
		{
			"Total", "", "",
			convertToHumanReadable(total, currency),
		},
	}

	kc.l.Table(kc.ctx, headers, rows)

	return total, nil
}

func (kc *Controller) priceCalculatorForManagedCluster(inp PriceCalculatorInput) (float64, error) {
	managedNodeCost := float64(inp.NoOfWorkerNodes) * inp.WorkerMachine.GetCost()

	total := managedNodeCost + inp.ManagedControlPlaneMachine.GetCost()

	currency := inp.Currency

	headers := []string{"Resource", "UnitCost", "Quantity", "Cost"}
	rows := [][]string{
		{
			"Managed Node(s)",
			convertToHumanReadable(inp.WorkerMachine.GetCost(), currency),
			strconv.Itoa(inp.NoOfWorkerNodes),
			convertToHumanReadable(managedNodeCost, currency),
		},
		{
			"Cloud-Managed Control Plane",
			convertToHumanReadable(inp.ManagedControlPlaneMachine.GetCost(), currency),
			"1",
			convertToHumanReadable(inp.ManagedControlPlaneMachine.GetCost(), currency),
		},
		{"", "", "", ""},
		{
			"Total", "", "",
			convertToHumanReadable(total, currency),
		},
	}

	kc.l.Table(kc.ctx, headers, rows)

	return total, nil
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

func (kc *Controller) GetPriceForInstance(region string, instanceType string) (_ float64, errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.client) {
		return 0.0, nil
	}

	prices, err := kc.cc.GetPriceInstanceType(region, instanceType)
	if err != nil {
		return 0.0, err
	}

	return prices.GetCost(), nil
}

func (kc *Controller) ListAllInstances(region string) (
	out provider.InstancesRegionOutput,
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

	o := optimizer.NewOptimizer(kc.ctx, kc.l, nil)
	_instances, err := o.AttachEmbodiedToInstanceType(instances, kc.client.Metadata.Provider)
	if err != nil {
		kc.l.Debug(kc.ctx, "Failed to attach embodied to instance type", "reason", err)
		return nil, nil
	}

	return _instances, nil
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
	oldVer := 3
	for i := range vers {
		if !semver.IsValid(vers[i]) {
			isRepoRespectSemver = false
			oldVer = len(strings.Split(vers[i], "."))
			vers[i] = semver.Canonical("v" + vers[i])
		}
	}

	sort.Slice(vers, func(i, j int) bool {
		return semver.Compare(vers[i], vers[j]) > 0
	})

	tags := make([]string, 0, len(vers))

	for _, r := range vers {
		if !isRepoRespectSemver {
			r = strings.TrimPrefix(r, "v")
			_v := strings.Split(r, ".")
			if oldVer < len(_v) {
				r = strings.Join(_v[:oldVer], ".")
			}
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

func (kc *Controller) ListManagedCNIs() (
	_ addons.ClusterAddons, defaultOptionManaged string,
	_ addons.ClusterAddons, defaultOptionKsctl string,
	errC error) {

	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.client.Metadata.ClusterType != consts.ClusterTypeMang {
		return nil, "", nil, "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInvalidUserInput,
			"Only supported for managed clusters",
		)
	}

	c, d, err := kc.cc.GetAvailableManagedCNIPlugins(kc.client.Metadata.Region)
	if err != nil {
		return nil, "", nil, "", err
	}

	a, b := cni.GetCNIs()

	return c, d, a, b, nil
}

func (kc *Controller) ListBootstrapCNIs() (
	_ addons.ClusterAddons, defaultOptionManaged string,
	_ addons.ClusterAddons, defaultOptionKsctl string,
	errC error) {

	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if kc.client.Metadata.ClusterType != consts.ClusterTypeSelfMang {
		return nil, "", nil, "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInvalidUserInput,
			"Only supported for self-managed clusters",
		)
	}

	c, d, err := kc.bb.D.GetAvailableCNIPlugins()
	if err != nil {
		return nil, "", nil, "", err
	}
	a, b := cni.GetCNIs()

	return c, d, a, b, nil
}

func (kc *Controller) findManagedOfferingCostAcrossRegions(
	availRegions []provider.RegionOutput,
	managedOfferingSku string,
) (map[string]float64, error) {
	resultChan := make(chan struct {
		region string
		price  float64
		err    error
	}, len(availRegions))

	for _, region := range availRegions {
		regSku := region.Sku
		go func(sku string) {
			_price, err := kc.ListAllManagedClusterManagementOfferings(regSku, nil)
			if err == nil {
				v, ok := _price[managedOfferingSku]
				if ok {
					resultChan <- struct {
						region string
						price  float64
						err    error
					}{sku, v.GetCost(), nil}
				} else {
					resultChan <- struct {
						region string
						price  float64
						err    error
					}{sku, 0.0, fmt.Errorf("managed offering not found")}
				}

			} else {
				resultChan <- struct {
					region string
					price  float64
					err    error
				}{sku, 0.0, err}
			}
		}(regSku)
	}

	cost := make(map[string]float64, len(availRegions))
	for i := 0; i < len(availRegions); i++ {
		result := <-resultChan
		if result.err == nil {
			cost[result.region] = result.price
		}
	}
	close(resultChan)

	return cost, nil
}

// findInstanceCostAcrossRegions it returns a map of K[V] where K is the region and V is the cost of the instance
func (kc *Controller) findInstanceCostAcrossRegions(
	availRegions []provider.RegionOutput,
	instanceSku string,
) (map[string]float64, error) {
	resultChan := make(chan struct {
		region string
		price  float64
		err    error
	}, len(availRegions))

	for _, region := range availRegions {
		regSku := region.Sku
		go func(sku string) {
			price, err := kc.GetPriceForInstance(sku, instanceSku)
			resultChan <- struct {
				region string
				price  float64
				err    error
			}{sku, price, err}
		}(regSku)
	}

	cost := make(map[string]float64, len(availRegions))
	for i := 0; i < len(availRegions); i++ {
		result := <-resultChan
		if result.err == nil {
			cost[result.region] = result.price
		}
	}

	close(resultChan)

	return cost, nil
}

type CostOptimizerInput struct {
	// ManagedOffering needs to be added for managed cluster it denotes managed control plane
	ManagedOffering provider.ManagedClusterOutput

	// ManagedPlane needs to be added for managed cluster it denotes managed plane
	ManagedPlane provider.InstanceRegionOutput

	// WorkerPlane needs to be added for self managed cluster it denotes worker plane
	WorkerPlane provider.InstanceRegionOutput

	// DataStorePlane needs to be added for self managed cluster it denotes DataStorePlane plane
	DataStorePlane provider.InstanceRegionOutput

	// ControlPlane needs to be added for self managed cluster it denotes control plane
	ControlPlane provider.InstanceRegionOutput

	// LoadBalancer needs to be added for self managed cluster it denotes load balancer
	LoadBalancer provider.InstanceRegionOutput

	CountOfWorkerNodes       int
	CountOfControlPlaneNodes int
	CountOfEtcdNodes         int

	CountOfManagedNodes int
}

func (kc *Controller) CostOptimizeAcrossRegions(
	regions []provider.RegionOutput,
	currentRegion string,
	req CostOptimizerInput,
) (_ *optimizer.RecommendationAcrossRegions, _ error) {

	o := optimizer.NewOptimizer(kc.ctx, kc.l, regions)
	newRegions, err := o.AttachEmissionsToRegions(kc.client.Metadata.Provider)
	if err != nil {
		kc.l.Debug(kc.ctx, "Failed to attach emissions to regions", "reason", err)
		return nil, err
	}
	o.AvailRegions = newRegions

	regionMap := make(map[string]provider.RegionOutput)
	for _, region := range newRegions {
		regionMap[region.Sku] = region
	}

	if kc.client.Metadata.ClusterType == consts.ClusterTypeSelfMang {
		_o := o.OptimizeSelfManagedInstanceTypesAcrossRegions(
			req.ControlPlane,
			req.WorkerPlane,
			req.DataStorePlane,
			req.LoadBalancer,
			kc.findInstanceCostAcrossRegions,
		)
		resC, err := o.InstanceTypeOptimizerAcrossRegions(
			regionMap,
			consts.ClusterTypeMang,
			nil, _o,
			currentRegion,
			req.CountOfControlPlaneNodes, req.CountOfWorkerNodes, req.CountOfEtcdNodes,
			"", req.ControlPlane.Sku, req.WorkerPlane.Sku, req.DataStorePlane.Sku, req.LoadBalancer.Sku,
		)
		if err != nil {
			return nil, err
		}
		return &resC, nil
	}

	_o := o.OptimizeManagedOfferingsAcrossRegions(
		req.ManagedOffering,
		req.ManagedPlane,
		kc.findInstanceCostAcrossRegions,
		kc.findManagedOfferingCostAcrossRegions,
	)

	resC, err := o.InstanceTypeOptimizerAcrossRegions(
		regionMap,
		consts.ClusterTypeMang,
		_o, nil,
		currentRegion,
		req.CountOfControlPlaneNodes, req.CountOfManagedNodes, 0,
		req.ManagedOffering.Sku, "", req.ManagedPlane.Sku, "", "",
	)
	if err != nil {
		return nil, err
	}
	return &resC, nil
}
