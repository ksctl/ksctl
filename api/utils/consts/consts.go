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

type KsctlCounterConts uint32

const (
	SSH_PAUSE_IN_SECONDS time.Duration = 20 * time.Second
)

const (
	MAX_RETRY_COUNT       KsctlCounterConts = 8
	MAX_WATCH_RETRY_COUNT KsctlCounterConts = 4
)

const (
	CREDENTIAL_PATH     KsctlUtilsConsts = 0
	CLUSTER_PATH        KsctlUtilsConsts = 1
	SSH_PATH            KsctlUtilsConsts = 2
	OTHER_PATH          KsctlUtilsConsts = 3
	EXEC_WITH_OUTPUT    KsctlUtilsConsts = 1
	EXEC_WITHOUT_OUTPUT KsctlUtilsConsts = 0
)

const (
	ROLE_CP KsctlRole = "controlplane"
	ROLE_WP KsctlRole = "workerplane"
	ROLE_LB KsctlRole = "loadbalancer"
	ROLE_DS KsctlRole = "datastore"
)
const (
	CLOUD_CIVO  KsctlCloud = "civo"
	CLOUD_AZURE KsctlCloud = "azure"
	CLOUD_LOCAL KsctlCloud = "local"
	CLOUD_AWS   KsctlCloud = "aws"
)
const (
	K8S_K3S     KsctlKubernetes = "k3s"
	K8S_KUBEADM KsctlKubernetes = "kubeadm"
)

const (
	STORE_LOCAL  KsctlStore = "local"
	STORE_REMOTE KsctlStore = "remote"
)
const (
	OPERATION_STATE_GET    KsctlOperation = "get"
	OPERATION_STATE_CREATE KsctlOperation = "create"
	OPERATION_STATE_DELETE KsctlOperation = "delete"
)

const (
	CLUSTER_TYPE_HA   KsctlClusterType = "ha"
	CLUSTER_TYPE_MANG KsctlClusterType = "managed"
)

const (
	// makes the fake client
	KSCTL_FAKE_FLAG KsctlSpecialFlags = "KSCTL_FAKE_FLAG_ENABLED"

	// KSCTL_TEST_DIR_ENABLED use this as environment variable to set a different home directory for ksctl during testing
	KSCTL_TEST_DIR_ENABLED KsctlSpecialFlags = "KSCTL_TEST_DIR_ENABLED"
)
