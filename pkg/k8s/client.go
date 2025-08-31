// Copyright 2024 Ksctl Authors
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

	"github.com/ksctl/ksctl/v2/pkg/logger"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	ctx                 context.Context
	l                   logger.Logger
	clientset           kubernetes.Interface
	apiextensionsClient clientset.Interface
	r                   *rest.Config

	RawK    *kubernetes.Clientset
	RawKAPI rest.Interface
}

type App struct {
	CreateNamespace bool
	Namespace       string

	Version     string   // how to get the version propogated?
	Urls        []string // make sure you traverse backwards to uninstall resources
	PostInstall string
	Metadata    string
}

type restConfig func() (*rest.Config, error)

func WithKubeconfigContent(kubeconfigContent string) restConfig {
	return func() (*rest.Config, error) {
		return clientcmd.BuildConfigFromKubeconfigGetter(
			"",
			func() (*api.Config, error) {
				return clientcmd.Load([]byte(kubeconfigContent))
			},
		)
	}
}

func WithInClusterConfig() restConfig {
	return func() (*rest.Config, error) {
		return rest.InClusterConfig()
	}
}

func NewK8sClient(ctx context.Context, l logger.Logger, restConfig restConfig) (k *Client, err error) {
	k = new(Client)
	k.ctx = context.WithValue(ctx, "module", "kubernetes-client")
	k.l = l

	restC, err := restConfig()
	if err != nil {
		return nil, err
	}
	k.r = restC

	k.apiextensionsClient, err = clientset.NewForConfig(restC)
	if err != nil {
		return nil, err
	}

	k.RawK, err = kubernetes.NewForConfig(restC)
	if err != nil {
		return nil, err
	}

	k.clientset = k.RawK

	k.RawKAPI = k.RawK.Discovery().RESTClient()

	return k, nil
}
