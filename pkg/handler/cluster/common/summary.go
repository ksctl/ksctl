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

import "github.com/ksctl/ksctl/v2/pkg/k8s"

type SummaryOutput struct {
	ClusterName   string
	CloudProvider string
	ClusterType   string
	k8s.SummaryOutput
}

func (kc *Controller) ClusterSummary() (_ *SummaryOutput, errC error) {

	kubeconfig, err := kc.Switch()
	if err != nil {
		return nil, err
	}

	report := &SummaryOutput{
		ClusterName:   kc.s.ClusterName,
		CloudProvider: string(kc.s.InfraProvider),
		ClusterType:   string(kc.s.ClusterType),
	}

	c, err := k8s.NewK8sClient(
		kc.ctx,
		kc.l,
		k8s.WithKubeconfigContent(*kubeconfig),
	)
	if err != nil {
		return nil, err
	}

	latency, k8sVer, err := c.GetServerVersionAndLatency()
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to measure latency", "error", err)
	} else {
		report.RoundTripLatency = latency
		report.KubernetesVersion = k8sVer
	}

	healthCheck, err := c.GetHealthz(15)
	if err != nil {
		return nil, err
	}
	report.APIServerHealthCheck = healthCheck

	nodes, err := c.GetNodesSummary(30)
	if err != nil {
		return nil, err
	}
	report.Nodes = nodes

	components, err := c.GetControlPlaneVersions(30)
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get components information", "error", err)
	} else {
		report.ControlPlaneComponentVers = components
	}

	workloads, err := c.GetWorkloadSummary(60)
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get workload information", "error", err)
	} else {
		report.WorkloadSummary = *workloads
	}

	events, err := c.GetRecentWarningEvents(30)
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to get recent events", "error", err)
	} else {
		report.RecentWarningEvents = events
	}

	issues, err := c.DetectClusterIssues(30)
	if err != nil {
		kc.l.Warn(kc.ctx, "Unable to detect cluster issues", "error", err)
	} else {
		report.DetectedIssues = issues
	}

	return report, nil
}
