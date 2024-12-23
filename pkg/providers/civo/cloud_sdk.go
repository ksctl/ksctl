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

package civo

import (
	"github.com/civo/civogo"
)

type CloudSDK interface {
	CreateNetwork(label string) (*civogo.NetworkResult, error)
	DeleteNetwork(id string) (*civogo.SimpleResponse, error)
	GetNetwork(id string) (*civogo.Network, error)

	NewFirewall(config *civogo.FirewallConfig) (*civogo.FirewallResult, error)
	DeleteFirewall(id string) (*civogo.SimpleResponse, error)

	NewSSHKey(name string, publicKey string) (*civogo.SimpleResponse, error)
	DeleteSSHKey(id string) (*civogo.SimpleResponse, error)

	CreateInstance(config *civogo.InstanceConfig) (*civogo.Instance, error)
	DeleteInstance(id string) (*civogo.SimpleResponse, error)
	GetInstance(id string) (*civogo.Instance, error)

	GetDiskImageByName(name string) (*civogo.DiskImage, error)

	InitClient(p *Provider, region string) error

	GetKubernetesCluster(id string) (*civogo.KubernetesCluster, error)
	NewKubernetesClusters(kc *civogo.KubernetesClusterConfig) (*civogo.KubernetesCluster, error)
	DeleteKubernetesCluster(id string) (*civogo.SimpleResponse, error)

	ListAvailableKubernetesVersions() ([]civogo.KubernetesVersion, error)

	ListRegions() ([]civogo.Region, error)
	ListInstanceSizes() ([]civogo.InstanceSize, error)
}
