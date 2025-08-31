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

package k8s

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type nodeUtilization struct {
	Name              string  `json:"node_name"`
	CPUUtilization    float64 `json:"cpu_utilization"` // in percentage
	MemoryUtilization float64 `json:"memory_utilization"`

	CPUUnits string `json:"cpu_units"`
	MemUnits string `json:"mem_units"`
}

type ContainerSummary struct {
	Name           string                       `json:"name"`
	RestartCount   int32                        `json:"restart_count"`
	Ready          bool                         `json:"ready"`
	WaitingProblem corev1.ContainerStateWaiting `json:"waiting_problem"`
}

type PodOwnerRef struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PodSummary struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	OwnerRef  []PodOwnerRef `json:"owner_ref"`
	IsFailed  bool          `json:"is_failed"`
	IsPending bool          `json:"is_pending"`

	FailedContainers []ContainerSummary `json:"failed_containers"`
}

type WorkloadSummary struct {
	Deployments  int `json:"deployments"`
	StatefulSets int `json:"statefulsets"`
	DaemonSets   int `json:"daemonsets"`
	CronJobs     int `json:"cronjobs"`
	Namespaces   int `json:"namespaces"`
	PVC          int `json:"pvc"`
	PV           int `json:"pv"`
	SC           int `json:"sc"`

	LoadbalancerSVC int `json:"loadbalancer_svc"`
	ClusterIPSVC    int `json:"clusterip_svc"`

	RunningPods   int          `json:"running_pods"`
	UnHealthyPods []PodSummary `json:"unhealthy_pods"`
}

type ClusterIssue struct {
	Severity       string `json:"severity"` // "Warning", "Error", "Critical"
	Component      string `json:"component"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

type EventSummary struct {
	Time      time.Time `json:"time"`
	Kind      string    `json:"kind"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`

	Reason     string `json:"reason"`
	Message    string `json:"message"`
	ReportedBy string `json:"reported_by"`

	Count int32 `json:"count"`
}

type APIServerHealthCheck struct {
	Healthy          bool     `json:"healthy"`
	FailedComponents []string `json:"failed_components"`
}

type NodeSummary struct {
	Name               string `json:"name"`
	KubeletHealthy     bool   `json:"kubelet_healthy"`
	Ready              bool   `json:"ready"`
	MemoryPressure     bool   `json:"memory_pressure"`
	DiskPressure       bool   `json:"disk_pressure"`
	NetworkUnavailable bool   `json:"network_unavailable"`

	// Node Details
	KernelVersion           string `json:"kernel_version"`
	OSImage                 string `json:"os_image"`
	ContainerRuntimeVersion string `json:"container_runtime_version"`
	KubeletVersion          string `json:"kubelet_version"`
	OperatingSystem         string `json:"operating_system"`
	Architecture            string `json:"architecture"`

	CPUUtilization    float64 `json:"cpu_utilization"`    // in percentage
	MemoryUtilization float64 `json:"memory_utilization"` // in percentage

	CPUUnits string `json:"cpu_units"`
	MemUnits string `json:"mem_units"`

	// Spec
	Unschedulable bool `json:"unschedulable"`
}

type SummaryOutput struct {
	RoundTripLatency          string                `json:"round_trip_latency"`
	KubernetesVersion         string                `json:"kubernetes_version"`
	APIServerHealthCheck      *APIServerHealthCheck `json:"api_server_health_check,omitempty"`
	ControlPlaneComponentVers map[string]string     `json:"control_plane_component_versions,omitempty"`
	Nodes                     []NodeSummary         `json:"nodes"`
	WorkloadSummary           WorkloadSummary       `json:"workload_summary"`
	DetectedIssues            []ClusterIssue        `json:"detected_issues"`
	RecentWarningEvents       []EventSummary        `json:"recent_warning_events,omitempty"`
}
