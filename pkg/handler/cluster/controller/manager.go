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

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/bootstrap"
	"github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions"
	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/errors"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/storage"
)

type Client struct {
	Cloud provider.Cloud

	PreBootstrap bootstrap.Bootstrap

	Bootstrap distributions.KubernetesDistribution

	Storage storage.Storage

	Metadata Metadata
}

type KsctlWorkerConfiguration struct {
	// WorkerCtx contains the creds and also userId
	WorkerCtx   context.Context
	PollerCache cache.Cache
	Storage     storage.Storage
}

type Metadata struct {
	ClusterName   string                  `json:"cluster_name"`
	Region        string                  `json:"region"`
	Provider      consts.KsctlCloud       `json:"cloud_provider"`
	K8sDistro     consts.KsctlKubernetes  `json:"kubernetes_distro"`
	StateLocation consts.KsctlStore       `json:"storage_type"`
	ClusterType   consts.KsctlClusterType `json:"cluster_type"`

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

	EtcdVersion string `json:"etcd_version"`

	// Addons Helps us with specifying cloud managed cluster addons (aks, eks, gke)
	// to k3s, kubeadm specific as well
	Addons addons.ClusterAddons `json:"addons,omitempty"`
}

type Controller struct {
	l                 logger.Logger
	ctx               context.Context
	KsctlStore        storage.Storage
	KsctlWorkloadConf KsctlWorkerConfiguration
}

func NewBaseController(
	ctx context.Context,
	l logger.Logger,
	ksctlConfig KsctlWorkerConfiguration,
) *Controller {
	b := new(Controller)
	b.l = l
	b.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "manager-base")
	b.KsctlWorkloadConf = ksctlConfig

	return b
}

func (cc *Controller) PanicHandler(log logger.Logger) error {
	if r := recover(); r != nil {
		return errors.WrapErrorf(ksctlErrors.ErrPanic, "Cause: {%v}\nTraceback\n%v", r, string(debug.Stack()))
	}
	return nil
}
