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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type DirectConnect struct {
	ctx context.Context
	l   logger.Logger

	conn  *tls.Config
	url   string
	token *string
	r     *rest.Config
	rC    *Client
}

func NewDirectConnect(
	ctx context.Context,
	l logger.Logger,
	clusterContextName string,
	kubeconfig string,
) (*DirectConnect, error) {
	dc := new(DirectConnect)
	dc.ctx = ctx
	dc.l = l
	dc.r = &rest.Config{}

	err := dc.establishConnection(clusterContextName, kubeconfig)
	if err != nil {
		return nil, err
	}

	dc.r, err = clientcmd.BuildConfigFromKubeconfigGetter(
		"",
		func() (*api.Config, error) {
			return clientcmd.Load([]byte(kubeconfig))
		})
	if err != nil {
		return nil, err
	}

	dc.rC, err = NewK8sClient(ctx, l, dc.r)
	if err != nil {
		return nil, err
	}

	return dc, nil
}

func (c *DirectConnect) establishConnection(contextName string, kubeconfig string) error {
	config, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed deserializes the contents into Config object", "Reason", err),
		)
	}

	clusterContext := ""
	authContext := ""
	isPresent := false
	c.l.Debug(c.ctx, "searching for current-context", "contextName", contextName)
	if config.CurrentContext != contextName {
		c.l.Warn(c.ctx, "failed context looking for is not the current one", "expected", contextName, "got", config.CurrentContext)
		c.l.Print(c.ctx, "using the context which is present in the state for configuration", "stateContext", contextName)
	}

	for ctxK8s, info := range config.Contexts {

		if ctxK8s == contextName {
			isPresent = true
			clusterContext = info.Cluster
			authContext = info.AuthInfo
			c.l.Debug(c.ctx, "Found cluster in kubeconfig",
				"current-context", config.CurrentContext,
				"contexts[...].context.cluster", clusterContext,
				"contexts[...].context.authinfo", authContext,
			)
		}
	}

	if !isPresent {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to find the context", "contextName", contextName),
		)
	}

	cluster := config.Clusters[clusterContext]
	kubeapiURL := cluster.Server

	if authContext == "aws" {
		token := config.AuthInfos[authContext].Token
		if len(token) == 0 {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedConnectingKubernetesCluster,
				c.l.NewError(c.ctx, "failed to get the token", "Reason", "token is empty"),
			)
		}
		tlsConf, _err := c.httpClient(true, cluster.CertificateAuthorityData, nil, nil)
		if _err != nil {
			return _err
		}
		c.conn = tlsConf
		c.url = kubeapiURL
		c.token = &token

		return nil

	} else {
		usr := config.AuthInfos[authContext]

		caCert := cluster.CertificateAuthorityData
		clientCert := usr.ClientCertificateData
		clientKey := usr.ClientKeyData
		if len(caCert) == 0 || len(clientCert) == 0 || len(clientKey) == 0 {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedConnectingKubernetesCluster,
				c.l.NewError(c.ctx, "failed to get the tls certs", "Reason", "one of the tls certs is empty"),
			)
		}

		tlsConf, _err := c.httpClient(false, caCert, clientCert, clientKey)
		if _err != nil {
			return _err
		}
		c.conn = tlsConf
		c.url = kubeapiURL
		c.token = nil

		return nil
	}
}

func (c *DirectConnect) httpClient(isTokenBased bool, caCert, clientCert, clientKey []byte) (*tls.Config, error) {

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	if isTokenBased {
		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}
		return tlsConfig, nil
	}

	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "Error loading client certificate and key", "Reason", err),
		)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig, nil
}

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

func (c *DirectConnect) MeasureLatency() (string, string, error) {
	url := fmt.Sprintf("%s/version", c.url)

	c.l.Debug(c.ctx, "full url for version", "url", url)

	tr := &http.Transport{
		TLSClientConfig: c.conn,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed, client could not create request", "Reason", err),
		)
	}

	if c.token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *c.token))
	}

	client := &http.Client{Transport: tr, Timeout: 1 * time.Minute}

	start := time.Now()

	resHttp, err := client.Do(req)
	if err != nil {
		return "", "", ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to connect",
				"Reason", err,
			),
		)
	}
	latency := time.Since(start).String()

	if resHttp.StatusCode != http.StatusOK {
		return "", "", ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to connect",
				"Reason", fmt.Sprintf("status code was %d", resHttp.StatusCode),
			),
		)
	}

	defer resHttp.Body.Close()
	type Version struct {
		GitVersion string `json:"gitVersion"`
	}
	var v Version

	if err := json.NewDecoder(resHttp.Body).Decode(&v); err != nil {
		return latency, "", ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to decode the body",
				"Reason", err,
			),
		)
	}

	return latency, v.GitVersion, nil
}

func (c *DirectConnect) GetControlPlaneVersions() (map[string]string, error) {
	pods, err := c.rC.clientset.CoreV1().Pods("kube-system").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get kube-system pods", "Reason", err),
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

func (c *DirectConnect) GetHealthz() (*APIServerHealthCheck, error) {
	url := fmt.Sprintf("%s/healthz?verbose", c.url)

	c.l.Debug(c.ctx, "full url for state transfer", "url", url)

	tr := &http.Transport{
		TLSClientConfig: c.conn,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed, client could not create request", "Reason", err),
		)
	}

	if c.token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *c.token))
	}

	client := &http.Client{Transport: tr, Timeout: 1 * time.Minute}

	resHttp, err := client.Do(req)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to connect",
				"Reason", err,
			),
		)
	}

	fail_1 := resHttp.StatusCode == http.StatusOK

	defer resHttp.Body.Close()
	body, _err := io.ReadAll(resHttp.Body)
	if _err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "status code was 200, but failed to read response",
				"Reason", _err,
			),
		)
	}
	var failedComponent []string
	for _, line := range strings.Split(string(body), "\n") {
		components := strings.Split(line, " ")

		if !strings.HasPrefix(components[0], "[+]") {
			continue
		}

		if components[1] != "ok" {
			failedComponent = append(failedComponent, strings.TrimPrefix("[+]", components[0]))
		}
	}

	return &APIServerHealthCheck{fail_1, failedComponent}, nil
}

func (c *DirectConnect) GetNodesSummary() ([]NodeSummary, error) {
	nodes, err := c.rC.clientset.CoreV1().Nodes().List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get nodes", "Reason", err),
		)
	}

	nodeMetrics, err := c.getClusterUtilization()
	if err != nil {
		c.l.Warn(c.ctx, "Unable to get utilization information", "error", err)
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

func (c *DirectConnect) GetWorkloadSummary() (*WorkloadSummary, error) {
	ns, err := c.rC.clientset.CoreV1().Namespaces().List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get namespaces", "Reason", err),
		)
	}

	deployment, err := c.rC.clientset.AppsV1().Deployments("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get deployment", "Reason", err),
		)
	}

	statefulSet, err := c.rC.clientset.AppsV1().StatefulSets("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get statefulSet", "Reason", err),
		)
	}

	daemonSet, err := c.rC.clientset.AppsV1().DaemonSets("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get daemonSet", "Reason", err),
		)
	}

	cronJob, err := c.rC.clientset.BatchV1().CronJobs("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get cronJob", "Reason", err),
		)
	}

	persistentVolumeClaim, err := c.rC.clientset.CoreV1().PersistentVolumeClaims("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get persistentVolumeClaim", "Reason", err),
		)
	}

	persistentVolume, err := c.rC.clientset.CoreV1().PersistentVolumes().List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get persistentVolume", "Reason", err),
		)
	}

	storageClass, err := c.rC.clientset.StorageV1().StorageClasses().List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get storageClass", "Reason", err),
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

	pods, err := c.rC.clientset.CoreV1().Pods("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get pods", "Reason", err),
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

			if pod.Status.Phase == corev1.PodPending {
				v.IsPending = true
			} else if pod.Status.Phase == corev1.PodFailed {
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

func (c *DirectConnect) GetRecentWarningEvents() ([]EventSummary, error) {
	events, err := c.rC.clientset.EventsV1().Events("").List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get events", "Reason", err),
		)
	}

	res := make([]EventSummary, 0, len(events.Items))
	for _, event := range events.Items {
		if event.Type == corev1.EventTypeWarning {
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
	}

	return res, nil
}

func (c *DirectConnect) DetectClusterIssues() ([]ClusterIssue, error) {

	pods, err := c.rC.clientset.CoreV1().Pods("").List(c.ctx, v1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get nodes", "Reason", err),
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

	namespaces, err := c.rC.clientset.CoreV1().Namespaces().List(c.ctx, v1.ListOptions{
		FieldSelector: "status.phase=Active",
		LabelSelector: "kubernetes.io/metadata.name!=kube-system",
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get namespaces", "Reason", err),
		)
	}

	for _, ns := range namespaces.Items {
		if strings.HasPrefix(ns.Name, "kube-") {
			continue
		}

		resourceQuota, err := c.rC.clientset.CoreV1().ResourceQuotas(ns.Name).List(c.ctx, v1.ListOptions{})
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrFailedConnectingKubernetesCluster,
				c.l.NewError(c.ctx, "failed to get resource quotas", "Reason", err),
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

	svcs, err := c.rC.clientset.CoreV1().Services("").List(c.ctx, v1.ListOptions{
		LabelSelector: "kubernetes.io/metadata.name!=kube-system",
	})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get services", "Reason", err),
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

func (c *DirectConnect) getClusterUtilization() ([]nodeUtilization, error) {

	url := fmt.Sprintf("%s/apis/metrics.k8s.io/v1beta1/nodes", c.url)

	c.l.Debug(c.ctx, "full url for node usage", "url", url)

	tr := &http.Transport{
		TLSClientConfig: c.conn,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed, client could not create request", "Reason", err),
		)
	}

	if c.token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *c.token))
	}

	client := &http.Client{Transport: tr, Timeout: 1 * time.Minute}

	resHttp, err := client.Do(req)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to connect",
				"Reason", err,
			),
		)
	}

	if resHttp.StatusCode == http.StatusNotFound {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get utilization",
				"Reason", "metrics.k8s.io not found",
			),
		)
	}

	defer resHttp.Body.Close()
	body, _err := io.ReadAll(resHttp.Body)
	if _err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to read the body",
				"Reason", _err,
			),
		)
	}

	if resHttp.StatusCode != http.StatusOK {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to connect",
				"Reason", fmt.Sprintf("status code was %d\n%s", resHttp.StatusCode, string(body)),
			),
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

	nL, err := c.rC.clientset.CoreV1().Nodes().List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get nodes", "Reason", err),
		)
	}

	res := make([]nodeUtilization, 0, len(nL.Items))

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

				res = append(res, nodeUtilz)
			}
		}
	}

	return res, nil
}
