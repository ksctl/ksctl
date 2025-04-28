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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	corev1 "k8s.io/api/core/v1"
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
	c.l.Print(c.ctx, "searching for current-context", "contextName", contextName)
	if config.CurrentContext != contextName {
		c.l.Warn(c.ctx, "failed context looking for is not the current one", "expected", contextName, "got", config.CurrentContext)
		c.l.Print(c.ctx, "using the context which is present in the state for configuration", "stateContext", contextName)
	}

	for ctxK8s, info := range config.Contexts {

		if ctxK8s == contextName {
			isPresent = true
			clusterContext = info.Cluster
			authContext = info.AuthInfo
			c.l.Print(c.ctx, "Found cluster in kubeconfig",
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

type APIServerHealthCheck struct {
	Healthy          bool
	FailedComponents []string
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

	// Spec
	Unschedulable bool
}

func (c *DirectConnect) GetNodesSummary() ([]NodeSummary, error) {
	nodes, err := c.rC.clientset.CoreV1().Nodes().List(c.ctx, v1.ListOptions{})
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedConnectingKubernetesCluster,
			c.l.NewError(c.ctx, "failed to get nodes", "Reason", err),
		)
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

	return res, nil
}
