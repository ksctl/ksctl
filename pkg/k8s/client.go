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

	"github.com/ksctl/ksctl/pkg/logger"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	ctx                 context.Context
	l                   logger.Logger
	clientset           kubernetes.Interface
	apiextensionsClient clientset.Interface
	r                   *rest.Config
}

type App struct {
	CreateNamespace bool
	Namespace       string

	Version     string   // how to get the version propogated?
	Urls        []string // make sure you traverse backwards to uninstall resources
	PostInstall string
	Metadata    string
}

func NewK8sClient(ctx context.Context, l logger.Logger, c *rest.Config) (k *Client, err error) {
	k = new(Client)
	k.ctx = context.WithValue(ctx, "module", "kubernetes-client")
	k.r = c
	k.l = l
	k.apiextensionsClient, err = clientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	k.clientset, err = kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return k, nil
}
