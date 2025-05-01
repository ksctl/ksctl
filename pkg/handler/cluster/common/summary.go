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

package common

import (
	"github.com/ksctl/ksctl/v2/pkg/k8s"
)

type SummaryOutput struct {
	// Cluster information
	ClusterName   string
	CloudProvider string
	ClusterType   string

	// Health information
	APIServerHealthCheck   *k8s.APIServerHealthCheck
	ControlPlaneComponents map[string]string // Component -> status

	// Resource information
	Nodes []k8s.NodeSummary

	ResourceUtilization []k8s.NodeUtilization

	// Workload information
	WorkloadSummary k8s.WorkloadSummary

	// Issues (for quick overview)
	DetectedIssues []k8s.ClusterIssue

	// Events
	RecentWarningEvents []k8s.EventSummary

	// Add-on information
	AddonsStatus map[string]string // Add-on name -> status
}

func (kc *Controller) ClusterSummary() (_ *SummaryOutput, errC error) {

	kubeconfig, err := kc.Switch()
	if err != nil {
		return nil, err
	}

	c, err := k8s.NewDirectConnect(kc.ctx, kc.l, kc.s.ClusterKubeConfigContext, *kubeconfig)
	if err != nil {
		return nil, err
	}

	res := &SummaryOutput{
		ClusterName:   kc.s.ClusterName,
		CloudProvider: string(kc.s.InfraProvider),
		ClusterType:   string(kc.s.ClusterType),
	}

	healthCheck, err := c.GetHealthz()
	if err != nil {
		return nil, err
	}
	res.APIServerHealthCheck = healthCheck

	nodes, err := c.GetNodesSummary()
	if err != nil {
		return nil, err
	}
	res.Nodes = nodes

	utilization, err := c.GetClusterUtilization()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get utilization information", "error", err)
	} else {
		res.ResourceUtilization = utilization
	}

	workloads, err := c.GetWorkloadSummary()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get workload information", "error", err)
	} else {
		res.WorkloadSummary = *workloads
	}

	events, err := c.GetRecentWarningEvents()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get recent events", "error", err)
	} else {
		res.RecentWarningEvents = events
	}

	issues, err := c.DetectClusterIssues()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to detect cluster issues", "error", err)
	} else {
		res.DetectedIssues = issues
	}

	return res, nil
}
