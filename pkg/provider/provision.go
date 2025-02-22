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

package provider

import (
	"context"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/consts"
)

type CloudResourceState struct {
	SSHUserName       string
	SSHPrivateKey     string
	IPv4ControlPlanes []string
	IPv4WorkerPlanes  []string
	IPv4DataStores    []string
	IPv4LoadBalancer  string

	PrivateIPv4ControlPlanes []string
	PrivateIPv4DataStores    []string
	PrivateIPv4LoadBalancer  string
	ClusterName              string
	Region                   string
	ClusterType              consts.KsctlClusterType
	Provider                 consts.KsctlCloud
}

type InfrastructureProvisioner interface {
	CreateInstance(ctx context.Context, stateIdx int, instanceName string, instanceRole consts.KsctlRole, instanceType string, public bool) error
	DeleteInstance(ctx context.Context, stateIdx int, instanceRole consts.KsctlRole) error

	CreateFirewall(ctx context.Context, firewallName string, firewallRole consts.KsctlRole) error
	DeleteFirewall(ctx context.Context, firewallRole consts.KsctlRole) error

	CreateVirtualNetwork(ctx context.Context, vNetName string) error
	DeleteVirtualNetwork(ctx context.Context) error

	UploadSSHKeyPair(ctx context.Context, keyPairName string) error
	DeleteSSHKeyPair(ctx context.Context) error

	CreateManagedKubernetesCluster(ctx context.Context, name string, workerNodes int, workerNodeInstanceType string, kubernetesVersion string) error
	DeleteManagedKubernetesCluster(ctx context.Context) error
}

type ProvisionerQuerier interface {
	List(ctx context.Context) ([]ClusterData, error)

	GetHostNameOfWorkerNodes(ctx context.Context) []string

	GetCountOfWorkerNodes(ctx context.Context) (int, error)
	GetCountOfControlPlaneNodes(ctx context.Context) (int, error)
	GetCountOfDataStoreNodes(ctx context.Context) (int, error)

	GetClusterDataForBootstrap(ctx context.Context) (CloudResourceState, error)

	GetKubeconfig(ctx context.Context) ([]byte, error)

	// TODO: why can't we completely drop these methods?
	SetCountOfWorkerNodes(ctx context.Context, count int) error
	SetCountOfControlPlaneNodes(ctx context.Context, count int) error
	SetCountOfDataStoreNodes(ctx context.Context, count int) error
}

type Cloud interface {
	ConfigureProvider(context.Context, consts.KsctlOperation) error

	InfrastructureProvisioner
	ProvisionerQuerier

	Addons(context.Context, addons.ClusterAddons) (willBeInstalled bool)

	CheckClusterPresence(context.Context) error
}
