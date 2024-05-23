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
	CounterMaxRetryCount          KsctlCounterConsts = 8
	CounterMaxNetworkSessionRetry KsctlCounterConsts = 9
	CounterMaxWatchRetryCount     KsctlCounterConsts = 4
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
	CloudCivo  KsctlCloud = "civo"
	CloudAzure KsctlCloud = "azure"
	CloudLocal KsctlCloud = "local"
	CloudAws   KsctlCloud = "aws"
	CloudAll   KsctlCloud = "all"
)
const (
	K8sK3s     KsctlKubernetes = "k3s"
	K8sKubeadm KsctlKubernetes = "kubeadm"
)

const (
	StoreLocal    KsctlStore = "local"
	StoreK8s      KsctlStore = "kubernetes"
	StoreExtMongo KsctlStore = "external-mongo"
)

const (
	OperationGet    KsctlOperation = "get"
	OperationCreate KsctlOperation = "create"
	OperationDelete KsctlOperation = "delete"
)

const (
	ClusterTypeHa   KsctlClusterType = "ha"
	ClusterTypeMang KsctlClusterType = "managed"
)

type KsctlContextKeyType int

const (
	KsctlTestFlag = iota
	ContextModuleNameKey
)

const (
	// TODO: replace this
	// KsctlCustomDirEnabled use this as environment variable to set a different home directory for ksctl during testing
	// make sure the value is space seperated <directory> <directory> ....
	KsctlCustomDirEnabled KsctlSpecialFlags = "KSCTL_CUSTOM_DIR_ENABLED"
)

const (
	CNIFlannel KsctlValidCNIPlugin = "flannel"
	CNICilium  KsctlValidCNIPlugin = "cilium"
	CNIAzure   KsctlValidCNIPlugin = "azure"
	CNIKubenet KsctlValidCNIPlugin = "kubenet"
	CNIKind    KsctlValidCNIPlugin = "kind"
	CNINone    KsctlValidCNIPlugin = "none"
)
