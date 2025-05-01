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

	RoundTripLatency  string
	KubernetesVersion string

	APIServerHealthCheck      *k8s.APIServerHealthCheck
	ControlPlaneComponentVers map[string]string

	Nodes []k8s.NodeSummary

	WorkloadSummary k8s.WorkloadSummary

	DetectedIssues []k8s.ClusterIssue

	RecentWarningEvents []k8s.EventSummary
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

	latency, k8sVer, err := c.MeasureLatency()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to measure latency", "error", err)
	} else {
		res.RoundTripLatency = latency
		res.KubernetesVersion = k8sVer
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

	components, err := c.GetControlPlaneVersions()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get components information", "error", err)
	} else {
		res.ControlPlaneComponentVers = components
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
