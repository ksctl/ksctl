package cloud

import "github.com/ksctl/ksctl/pkg/helpers/consts"

// CloudResourceState provides the state which cloud provider creates
// and which is consumed by the kubernetes to configure them
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
	//ClusterDir  string
	ClusterType consts.KsctlClusterType
	Provider    consts.KsctlCloud
}

type SSHInfo struct {
	UserName   string
	PrivateKey string
}

type AllClusterData struct {
	Name       string
	Provider   consts.KsctlCloud
	Type       consts.KsctlClusterType
	Region     string
	NoWP       int
	NoCP       int
	NoDS       int
	NoMgt      int
	K8sDistro  consts.KsctlKubernetes
	K8sVersion string
}
