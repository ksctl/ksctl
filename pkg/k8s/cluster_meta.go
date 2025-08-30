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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (k *Client) GetServerVersionAndLatency() (string, string, error) {
	start := time.Now()

	serverVersions, err := k.RawK.ServerVersion()
	if err != nil {
		return "", "", fmt.Errorf("failed to get server version: %w", err)
	}

	latency := time.Since(start).String()

	return latency, serverVersions.GitVersion, nil
}

func (k *Client) GetControlPlaneVersions(timeoutSeconds int64) (map[string]string, error) {
	pods, err := k.clientset.CoreV1().Pods("kube-system").List(k.ctx, v1.ListOptions{
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get kube-system pods", "Reason", err),
		)
	}

	versions := make(map[string]string)

	componentsToFind := map[string]bool{
		"kube-apiserver":          false,
		"kube-controller-manager": false,
		"kube-scheduler":          false,
		"etcd":                    false,
		"coredns":                 false,
	}

	for _, pod := range pods.Items {
		for component, alreadyFound := range componentsToFind {
			if alreadyFound {
				continue
			}
			if strings.Contains(pod.Name, component) {
				for _, container := range pod.Spec.Containers {
					if container.Name == component {
						versions[component] = container.Image
						componentsToFind[component] = true
						break
					}
				}
			}
		}
	}

	return versions, nil
}

func (k *Client) GetHealthz(timeoutSeconds int64) (*APIServerHealthCheck, error) {
	res := k.RawKAPI.Get().
		AbsPath("healthz").
		Param("verbose", "").
		Timeout(time.Duration(timeoutSeconds) * time.Second).
		Do(k.ctx)

	if res.Error() != nil {
		return nil, res.Error()
	}

	data, err := res.Raw()
	if err != nil {
		return nil, err
	}

	var failedComponent []string
	for line := range strings.SplitSeq(string(data), "\n") {
		components := strings.Split(line, " ")

		if !strings.HasPrefix(components[0], "[+]") {
			continue
		}

		if components[1] != "ok" {
			failedComponent = append(failedComponent, strings.TrimPrefix("[+]", components[0]))
		}
	}

	return &APIServerHealthCheck{
		Healthy:          len(failedComponent) == 0,
		FailedComponents: failedComponent,
	}, nil
}

func (k *Client) GetNodesSummary(timeoutSeconds int64) ([]NodeSummary, error) {
	nodes, err := k.clientset.CoreV1().Nodes().List(k.ctx, v1.ListOptions{
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get nodes", "Reason", err),
		)
	}

	nodeMetrics, err := k.getClusterUtilization(timeoutSeconds)
	if err != nil {
		k.l.Warn(k.ctx, "Unable to get utilization information", "error", err)
	}

	res := make([]NodeSummary, 0, len(nodes.Items))

	for _, node := range nodes.Items {
		resNode := NodeSummary{
			Name:                    node.Name,
			KernelVersion:           node.Status.NodeInfo.KernelVersion,
			OSImage:                 node.Status.NodeInfo.OSImage,
			ContainerRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion,
			KubeletVersion:          node.Status.NodeInfo.KubeletVersion,
			OperatingSystem:         node.Status.NodeInfo.OperatingSystem,
			Architecture:            node.Status.NodeInfo.Architecture,
			Unschedulable:           node.Spec.Unschedulable,
		}
		for _, condition := range node.Status.Conditions {
			kubeletHealthy := false
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					kubeletHealthy = true
					resNode.KubeletHealthy = true
					resNode.Ready = true
				}
			}
			if !kubeletHealthy {
				if condition.Type == corev1.NodeMemoryPressure && condition.Status == corev1.ConditionTrue {
					resNode.MemoryPressure = true
				}
				if condition.Type == corev1.NodeDiskPressure && condition.Status == corev1.ConditionTrue {
					resNode.DiskPressure = true
				}
				if condition.Type == corev1.NodeNetworkUnavailable && condition.Status == corev1.ConditionTrue {
					resNode.NetworkUnavailable = true
				}
			}
		}
		res = append(res, resNode)
	}

	for _, node := range nodeMetrics {
		for i := range res {
			if res[i].Name == node.Name {
				res[i].CPUUtilization = node.CPUUtilization
				res[i].MemoryUtilization = node.MemoryUtilization
				res[i].CPUUnits = node.CPUUnits
				res[i].MemUnits = node.MemUnits
			}
		}
	}

	return res, nil
}

// GetWorkloadSummary returns list of all the workloads.
//
//	TODO: we need to improve the performace of this function like ensuring only specific namespace to be be retrived aka a namespace which is managed by ksctl
func (k *Client) GetWorkloadSummary(timeoutSeconds int64) (*WorkloadSummary, error) {

	listOpts := v1.ListOptions{
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	}

	ns, err := k.clientset.CoreV1().Namespaces().List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get namespaces", "Reason", err),
		)
	}

	deployment, err := k.clientset.AppsV1().Deployments("").List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get deployment", "Reason", err),
		)
	}

	statefulSet, err := k.clientset.AppsV1().StatefulSets("").List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get statefulSet", "Reason", err),
		)
	}

	daemonSet, err := k.clientset.AppsV1().DaemonSets("").List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get daemonSet", "Reason", err),
		)
	}

	cronJob, err := k.clientset.BatchV1().CronJobs("").List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get cronJob", "Reason", err),
		)
	}

	persistentVolumeClaim, err := k.clientset.CoreV1().PersistentVolumeClaims("").List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get persistentVolumeClaim", "Reason", err),
		)
	}

	persistentVolume, err := k.clientset.CoreV1().PersistentVolumes().List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get persistentVolume", "Reason", err),
		)
	}

	storageClass, err := k.clientset.StorageV1().StorageClasses().List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get storageClass", "Reason", err),
		)
	}

	res := &WorkloadSummary{
		Deployments:  len(deployment.Items),
		StatefulSets: len(statefulSet.Items),
		DaemonSets:   len(daemonSet.Items),
		CronJobs:     len(cronJob.Items),
		Namespaces:   len(ns.Items),
		PVC:          len(persistentVolumeClaim.Items),
		PV:           len(persistentVolume.Items),
		SC:           len(storageClass.Items),
	}

	pods, err := k.clientset.CoreV1().Pods("").List(k.ctx, listOpts)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get pods", "Reason", err),
		)
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			res.RunningPods++
		} else {
			v := PodSummary{
				Name:      pod.Name,
				OwnerRef:  make([]PodOwnerRef, 0, len(pod.OwnerReferences)),
				Namespace: pod.Namespace,
			}
			for _, ownerRef := range pod.OwnerReferences {
				v.OwnerRef = append(v.OwnerRef, PodOwnerRef{
					Kind:      ownerRef.Kind,
					Name:      ownerRef.Name,
					Namespace: pod.Namespace,
				})
			}

			switch pod.Status.Phase {
			case corev1.PodPending:
				v.IsPending = true
			case corev1.PodFailed:
				v.IsFailed = true
			}
			for _, containerStatus := range pod.Status.ContainerStatuses {
				s := ContainerSummary{
					Name:         containerStatus.Name,
					RestartCount: containerStatus.RestartCount,
					Ready:        containerStatus.Ready,
				}
				if containerStatus.State.Waiting != nil {
					s.WaitingProblem = *containerStatus.State.Waiting
				}

				v.FailedContainers = append(v.FailedContainers, s)
			}

			res.UnHealthyPods = append(res.UnHealthyPods, v)
		}
	}

	return res, nil
}

func (k *Client) GetRecentWarningEvents(timeoutSeconds int64) ([]EventSummary, error) {
	events, err := k.clientset.EventsV1().Events("").List(k.ctx, v1.ListOptions{
		FieldSelector:  "type=" + corev1.EventTypeWarning,
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get events", "Reason", err),
		)
	}

	res := make([]EventSummary, 0, len(events.Items))
	for _, event := range events.Items {
		if time.Since(event.CreationTimestamp.Time) > 24*time.Hour {
			continue
		}
		res = append(res, EventSummary{
			Kind:       event.Regarding.Kind,
			Name:       event.Regarding.Name,
			Namespace:  event.Regarding.Namespace,
			Time:       event.CreationTimestamp.Time,
			Count:      event.DeprecatedCount,
			Reason:     event.Reason,
			ReportedBy: event.ReportingInstance,
			Message:    event.Note,
		})
	}

	return res, nil
}

func (k *Client) DetectClusterIssues(timeoutSeconds int64) ([]ClusterIssue, error) {

	pods, err := k.clientset.CoreV1().Pods("").List(k.ctx, v1.ListOptions{
		FieldSelector:  "status.phase=Running",
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get nodes", "Reason", err),
		)
	}

	res := make([]ClusterIssue, 0, len(pods.Items))
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Namespace, "kube-") {
			continue
		}
		for _, c := range pod.Spec.Containers {
			if c.ReadinessProbe == nil {
				res = append(res, ClusterIssue{
					Severity:       "Warning",
					Component:      "pod/" + pod.Name + ":::" + c.Name + "@" + pod.Namespace,
					Message:        "Container does not have a readiness probe",
					Recommendation: "Add a readiness probe to the container",
				})
			}
			if c.LivenessProbe == nil {
				res = append(res, ClusterIssue{
					Severity:       "Warning",
					Component:      "pod/" + pod.Name + ":::" + c.Name + "@" + pod.Namespace,
					Message:        "Container does not have a liveness probe",
					Recommendation: "Add a liveness probe to the container",
				})
			}
			if c.Resources.Limits == nil {
				res = append(res, ClusterIssue{
					Severity:       "Warning",
					Component:      "pod/" + pod.Name + ":::" + c.Name + "@" + pod.Namespace,
					Message:        "Container does not have resource limits set",
					Recommendation: "Set resource limits for the container",
				})
			}
			if c.Resources.Requests == nil {
				res = append(res, ClusterIssue{
					Severity:       "Critical",
					Component:      "pod/" + pod.Name + ":::" + c.Name + "@" + pod.Namespace,
					Message:        "Container does not have resource requests set",
					Recommendation: "Set resource requests for the container",
				})
			}
			if c.SecurityContext == nil {
				res = append(res, ClusterIssue{
					Severity:       "Critical",
					Component:      "pod/" + pod.Name + ":::" + c.Name + "@" + pod.Namespace,
					Message:        "Container does not have security context set",
					Recommendation: "Set security context for the container",
				})
			}
		}
	}

	namespaces, err := k.clientset.CoreV1().Namespaces().List(k.ctx, v1.ListOptions{
		FieldSelector:  "status.phase=Active",
		LabelSelector:  "kubernetes.io/metadata.name!=kube-system",
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get namespaces", "Reason", err),
		)
	}

	for _, ns := range namespaces.Items {
		if strings.HasPrefix(ns.Name, "kube-") {
			continue
		}

		resourceQuota, err := k.clientset.CoreV1().ResourceQuotas(ns.Name).List(k.ctx, v1.ListOptions{
			TimeoutSeconds: utilities.Ptr(timeoutSeconds),
		})
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrFailedConnectingKubernetesCluster,
				k.l.NewError(k.ctx, "failed to get resource quotas", "Reason", err),
			)
		}
		if len(resourceQuota.Items) == 0 {
			res = append(res, ClusterIssue{
				Severity:       "Warning",
				Component:      "namespace/" + ns.Name,
				Message:        "Namespace does not have resource quotas set",
				Recommendation: "Set resource quotas for the namespace",
			})
		}
	}

	svcs, err := k.clientset.CoreV1().Services("").List(k.ctx, v1.ListOptions{
		LabelSelector:  "kubernetes.io/metadata.name!=kube-system",
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			k.l.NewError(k.ctx, "failed to get services", "Reason", err),
		)
	}
	for _, svc := range svcs.Items {
		if strings.HasPrefix(svc.Name, "kube-") {
			continue
		}
		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			res = append(res, ClusterIssue{
				Severity:       "Critical",
				Component:      "service/" + svc.Name + "@" + svc.Namespace,
				Message:        "Service is of type NodePort",
				Recommendation: "Use LoadBalancer or ClusterIP service type",
			})
		}
	}

	return res, nil
}

func (c *Client) getClusterUtilization(timeoutSeconds int64) ([]nodeUtilization, error) {
	x := c.RawKAPI.
		Get().
		AbsPath("apis", "metrics.k8s.io", "v1beta1", "nodes")

	res := x.Do(c.ctx)
	if res.Error() != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get node metrics", "Reason", res.Error()),
		)
	}

	body, err := res.Raw()
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to read the body", "Reason", err),
		)
	}

	type NodeMetrics struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"items"`
	}

	var nm NodeMetrics
	err = json.Unmarshal(body, &nm)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to decode the body",
				"Reason", err,
			),
		)
	}

	nL, err := c.clientset.CoreV1().Nodes().List(c.ctx, v1.ListOptions{
		TimeoutSeconds: utilities.Ptr(timeoutSeconds),
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get nodes", "Reason", err),
		)
	}

	nodes := make([]nodeUtilization, 0, len(nL.Items))

	for _, n := range nL.Items {
		for _, nM := range nm.Items {
			if n.Name == nM.Metadata.Name {
				nodeUtilz := nodeUtilization{
					Name: n.Name,
				}

				cpuTotal := n.Status.Capacity[corev1.ResourceCPU]
				memTotal := n.Status.Capacity[corev1.ResourceMemory]
				cpuTotalInt := cpuTotal.MilliValue()
				memTotalInt := memTotal.MilliValue()

				cpuUtilized, _ := resource.ParseQuantity(nM.Usage.CPU)
				memUtilized, _ := resource.ParseQuantity(nM.Usage.Memory)
				cpuUtilizedInt := cpuUtilized.MilliValue()
				memUtilizedInt := memUtilized.MilliValue()

				nodeUtilz.CPUUtilization = float64(cpuUtilizedInt) / float64(cpuTotalInt) * 100
				nodeUtilz.MemoryUtilization = float64(memUtilizedInt) / float64(memTotalInt) * 100

				nodeUtilz.CPUUnits = cpuTotal.String()
				nodeUtilz.MemUnits = memTotal.String()

				nodes = append(nodes, nodeUtilz)
			}
		}
	}

	return nodes, nil
}
