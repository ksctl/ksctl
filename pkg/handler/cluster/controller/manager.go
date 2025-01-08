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

package controller

import (
	"context"
	"runtime/debug"

	"github.com/ksctl/ksctl/pkg/addons"
	"github.com/ksctl/ksctl/pkg/bootstrap"
	"github.com/ksctl/ksctl/pkg/bootstrap/distributions"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/provider"
	"github.com/ksctl/ksctl/pkg/storage"
)

type Client struct {
	Cloud provider.Cloud

	PreBootstrap bootstrap.Bootstrap

	Bootstrap distributions.KubernetesDistribution

	Storage storage.Storage

	Metadata Metadata
}

type Metadata struct {
	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`

	Provider      consts.KsctlCloud      `json:"cloud_provider"`
	K8sDistro     consts.KsctlKubernetes `json:"kubernetes_distro"`
	StateLocation consts.KsctlStore      `json:"storage_type"`

	SelfManaged bool `json:"self_managed"`

	K8sVersion string `json:"kubernetes_version"`

	ManagedNodeType      string `json:"node_type_managed"`
	WorkerPlaneNodeType  string `json:"node_type_workerplane"`
	ControlPlaneNodeType string `json:"node_type_controlplane"`
	DataStoreNodeType    string `json:"node_type_datastore"`
	LoadBalancerNodeType string `json:"node_type_loadbalancer"`

	NoMP int `json:"desired_no_of_managed_nodes"`      // No of managed Nodes
	NoWP int `json:"desired_no_of_workerplane_nodes"`  // No of woerkplane VMs
	NoCP int `json:"desired_no_of_controlplane_nodes"` // No of Controlplane VMs
	NoDS int `json:"desired_no_of_datastore_nodes"`    // No of DataStore VMs

	Addons addons.ClusterAddons `json:"addons"`
}

type Controller struct {
	l   logger.Logger
	ctx context.Context
}

func NewBaseController(ctx context.Context, l logger.Logger) *Controller {
	b := new(Controller)
	b.l = l
	b.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "manager-base")

	return b
}

// PanicHandler This function is intended to be used by the Cli and used once thus getting required information for developers to debug
func (cc *Controller) PanicHandler(log logger.Logger) {
	if r := recover(); r != nil {
		log.Error("Failed to recover stack trace", "error", r)
		log.Print(cc.ctx, "Controller Information", "context", cc.ctx)
		debug.PrintStack()
	}
}
