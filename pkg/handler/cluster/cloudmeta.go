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

package cluster

import (
	"github.com/ksctl/ksctl/pkg/consts"
)

// TODO: not sure where to put this

type CloudResourceState struct {
	SSHState          SSHInfo
	IPv4ControlPlanes []string
	IPv4WorkerPlanes  []string
	IPv4DataStores    []string
	IPv4LoadBalancer  string

	PrivateIPv4ControlPlanes []string
	PrivateIPv4DataStores    []string
	PrivateIPv4LoadBalancer  string
	Metadata                 Metadata
}

type Metadata struct {
	ClusterName string
	Region      string
	ClusterType consts.KsctlClusterType
	Provider    consts.KsctlCloud
}

type SSHInfo struct {
	UserName   string
	PrivateKey string
}

type VMData struct {
	VMID         string
	VMName       string
	VMSize       string
	FirewallID   string
	FirewallName string
	SubnetID     string
	SubnetName   string
	PublicIP     string
	PrivateIP    string
}

type AllClusterData struct {
	Name            string
	CloudProvider   consts.KsctlCloud
	ClusterType     consts.KsctlClusterType
	K8sDistro       consts.KsctlKubernetes
	SSHKeyName      string
	SSHKeyID        string
	Region          string
	ResourceGrpName string
	NetworkName     string
	NetworkID       string
	ManagedK8sID    string
	ManagedK8sName  string
	WP              []VMData
	CP              []VMData
	DS              []VMData
	LB              VMData
	Mgt             VMData
	NoWP            int
	NoCP            int
	NoDS            int
	NoMgt           int
	K8sVersion      string
	EtcdVersion     string
	HAProxyVersion  string
	Apps            []string
	Cni             string
}
