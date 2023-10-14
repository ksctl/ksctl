package consts

import "time"

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

const (
	DurationSSHPause time.Duration = 20 * time.Second
)

const (
	CounterMaxRetryCount      KsctlCounterConsts = 8
	CounterMaxWatchRetryCount KsctlCounterConsts = 4
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
	StoreLocal  KsctlStore = "local"
	StoreRemote KsctlStore = "remote"
)
const (
	OperationStateGet    KsctlOperation = "get"
	OperationStateCreate KsctlOperation = "create"
	OperationStateDelete KsctlOperation = "delete"
)

const (
	ClusterTypeHa   KsctlClusterType = "ha"
	ClusterTypeMang KsctlClusterType = "managed"
)

const (
	// makes the fake client
	KsctlFakeFlag KsctlSpecialFlags = "KSCTL_FAKE_FLAG_ENABLED"

	// KsctlCustomDirEnabled use this as environment variable to set a different home directory for ksctl during testing
	KsctlCustomDirEnabled KsctlSpecialFlags = "KSCTL_CUSTOM_DIR_ENABLED"

	// KsctlFeatureFlagHaAutoscale to be set if feature for AUTOSCALE is needed
	KsctlFeatureFlagHaAutoscale KsctlSpecialFlags = "KSCTL_FEATURE_FLAG_HA_AUTOSCALE"

	// KsctlFeatureFlagApplications to be set if feature for install Application is needed
	KsctlFeatureFlagApplications KsctlSpecialFlags = "KSCTL_FEATURE_FLAG_APPLICATIONS"
)

const (
	CNIFlannel KsctlValidCNIPlugin = "flannel"
	CNICilium  KsctlValidCNIPlugin = "cilium"
	CNIAzure   KsctlValidCNIPlugin = "azure"
	CNIKubenet KsctlValidCNIPlugin = "kubenet"
	CNIKind    KsctlValidCNIPlugin = "kind"
	CNINone    KsctlValidCNIPlugin = "none"
)
