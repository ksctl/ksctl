package cloud

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

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
	Apps            []string
	Cni             string
}
