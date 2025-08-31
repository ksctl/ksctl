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
	Name              string
	CPUUtilization    float64 // in percentage
	MemoryUtilization float64

	CPUUnits string
	MemUnits string
}

type ContainerSummary struct {
	Name           string
	RestartCount   int32
	Ready          bool
	WaitingProblem corev1.ContainerStateWaiting
}

type PodOwnerRef struct {
	Kind      string
	Name      string
	Namespace string
}

type PodSummary struct {
	Name      string
	Namespace string
	OwnerRef  []PodOwnerRef
	IsFailed  bool
	IsPending bool

	FailedContainers []ContainerSummary
}

type WorkloadSummary struct {
	Deployments  int
	StatefulSets int
	DaemonSets   int
	CronJobs     int
	Namespaces   int
	PVC          int
	PV           int
	SC           int

	LoadbalancerSVC int
	ClusterIPSVC    int

	RunningPods   int
	UnHealthyPods []PodSummary
}

type ClusterIssue struct {
	Severity       string // "Warning", "Error", "Critical"
	Component      string
	Message        string
	Recommendation string
}

type EventSummary struct {
	Time      time.Time
	Kind      string
	Name      string
	Namespace string

	Reason     string
	Message    string
	ReportedBy string

	Count int32
}

type APIServerHealthCheck struct {
	Healthy          bool
	FailedComponents []string
}

type NodeSummary struct {
	Name               string
	KubeletHealthy     bool
	Ready              bool
	MemoryPressure     bool
	DiskPressure       bool
	NetworkUnavailable bool

	// Node Details
	KernelVersion           string
	OSImage                 string
	ContainerRuntimeVersion string
	KubeletVersion          string
	OperatingSystem         string
	Architecture            string

	CPUUtilization    float64 // in percentage
	MemoryUtilization float64

	CPUUnits string
	MemUnits string

	// Spec
	Unschedulable bool
}

type SummaryOutput struct {
	RoundTripLatency  string
	KubernetesVersion string

	APIServerHealthCheck      *APIServerHealthCheck
	ControlPlaneComponentVers map[string]string

	Nodes []NodeSummary

	WorkloadSummary WorkloadSummary

	DetectedIssues []ClusterIssue

	RecentWarningEvents []EventSummary
}
