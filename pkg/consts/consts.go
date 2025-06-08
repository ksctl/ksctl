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

package consts

import (
	"time"
)

type KsctlRole string

type KsctlCloud string

type KsctlKubernetes string

type KsctlStore string

type KsctlOperation string

type KsctlClusterType string

type KsctlSpecialFlags string

type KsctlUtilsConsts uint8

type KsctlCounterConsts uint32

type KsctlValidCNIPlugin string

type KsctlSupportedScriptRunners string

type KsctlSearchFilter string

type FirewallRuleProtocol int

type FirewallRuleAction int

type FirewallRuleDirection int

const (
	FirewallActionAllow FirewallRuleAction = iota
	FirewallActionDeny  FirewallRuleAction = iota

	FirewallActionIngress FirewallRuleDirection = iota
	FirewallActionEgress  FirewallRuleDirection = iota

	FirewallActionTCP FirewallRuleProtocol = iota
	FirewallActionUDP FirewallRuleProtocol = iota
)

const (
	Cloud       KsctlSearchFilter = "cloud"
	ClusterType KsctlSearchFilter = "clusterType"
	Name        KsctlSearchFilter = "clusterName"
	Region      KsctlSearchFilter = "region"
)

const (
	LinuxSh   KsctlSupportedScriptRunners = "/bin/sh"
	LinuxBash KsctlSupportedScriptRunners = "/bin/bash"
)

const (
	DurationSSHPause time.Duration = 20 * time.Second
)

const (
	CounterMaxRetryCount          KsctlCounterConsts = 5
	CounterMaxNetworkSessionRetry KsctlCounterConsts = 5
	CounterMaxWatchRetryCount     KsctlCounterConsts = 3
)

const (
	UtilCredentialPath    KsctlUtilsConsts = 0
	UtilClusterPath       KsctlUtilsConsts = 1
	UtilSSHPath           KsctlUtilsConsts = 2
	UtilOtherPath         KsctlUtilsConsts = 3
	UtilExecWithOutput    KsctlUtilsConsts = 1
	UtilExecWithoutOutput KsctlUtilsConsts = 0
)

const (
	RoleCp KsctlRole = "controlplane"
	RoleWp KsctlRole = "workerplane"
	RoleLb KsctlRole = "loadbalancer"
	RoleDs KsctlRole = "datastore"
)
const (
	CloudAzure KsctlCloud = "azure"
	CloudLocal KsctlCloud = "local"
	CloudAws   KsctlCloud = "aws"
	CloudGcp   KsctlCloud = "gcp"
	CloudAll   KsctlCloud = "all"
)
const (
	K8sK3s     KsctlKubernetes = "k3s"
	K8sKubeadm KsctlKubernetes = "kubeadm"
	K8sAks     KsctlKubernetes = "aks"
	K8sEks     KsctlKubernetes = "eks"
	K8sKind    KsctlKubernetes = "kind"
	K8sKsctl   KsctlKubernetes = "ksctl"
)

const (
	StoreLocal    KsctlStore = "store-local"
	StoreK8s      KsctlStore = "store-kubernetes"
	StoreExtMongo KsctlStore = "external-store-mongodb"
)

const (
	OperationGet       KsctlOperation = "get"
	OperationCreate    KsctlOperation = "create"
	OperationDelete    KsctlOperation = "delete"
	OperationScale     KsctlOperation = "scale"
	OperationConfigure KsctlOperation = "configure"
)

const (
	ClusterTypeSelfMang KsctlClusterType = "selfmanaged"
	ClusterTypeMang     KsctlClusterType = "managed"
)

type KsctlContextKeyType int

const (
	KsctlTestFlagKey        KsctlContextKeyType = iota
	KsctlModuleNameKey      KsctlContextKeyType = iota
	KsctlCustomDirLoc       KsctlContextKeyType = iota
	KsctlContextUser        KsctlContextKeyType = iota
	KsctlContextTeam        KsctlContextKeyType = iota
	KsctlComponentOverrides KsctlContextKeyType = iota

	KsctlAwsCredentials     KsctlContextKeyType = iota // the value to be the AzureCredentials struct
	KsctlAzureCredentials   KsctlContextKeyType = iota // the value to be the AwsCredentials struct
	KsctlMongodbCredentials KsctlContextKeyType = iota // the value to be the MongodbCredentials struct
	KsctlRedisCredentials   KsctlContextKeyType = iota // the value to be the RedisCredentials struct
)

const (
	CNIFlannel KsctlValidCNIPlugin = "flannel"
	CNICilium  KsctlValidCNIPlugin = "cilium"
	CNIAzure   KsctlValidCNIPlugin = "azure"
	CNIKubenet KsctlValidCNIPlugin = "kubenet"
	CNIKind    KsctlValidCNIPlugin = "kind"
	CNINone    KsctlValidCNIPlugin = "none"
)
