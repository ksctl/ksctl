package types

import "github.com/ksctl/ksctl/pkg/helpers/consts"

type AWSStateVm struct {
	Vpc                  string `json:"vpc"`
	HostName             string `json:"name"`
	DiskSize             string `json:"disk_size"`
	InstanceType         string `json:"instance_type"`
	InstanceID           string `json:"instance_id"`
	Subnet               string `json:"subnet"`
	NetworkSecurityGroup string `json:"network_security_group"`
	PublicIP             string `json:"public_ip"`
	PrivateIP            string `json:"private_ip"`
	NetworkInterfaceId   string `json:"network_interface_id"`
}
type CredentialsAws struct {
	AcessKeyID     string
	AcessKeySecret string
}
type metadata struct {
	resName string
	role    consts.KsctlRole
	vmType  string
	public  bool

	apps    string
	cni     string
	version string

	noCP int
	noWP int
	noDS int

	k8sName    consts.KsctlKubernetes
	k8sVersion string
}
type AWSStateVms struct {
	HostNames            []string `json:"names"`
	DiskNames            []string `json:"disk_name"`
	InstanceIds          []string `json:"instance_id"`
	PrivateIPs           []string `json:"private_ip"`
	PublicIPs            []string `json:"public_ip"`
	NetworkInterfaceIDs  []string `json:"network_interface_id"`
	SubnetNames          []string `json:"subnet_name"`
	SubnetIDs            []string `json:"subnet_id"`
	NetworkSecurityGroup string   `json:"network_security_group"`
}

type StateConfigurationAws struct {
	B BaseInfra `json:"b" bson:"b"`

	IsCompleted bool
	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`
	VpcName     string `json:"vpc"`
	VpcId       string `json:"vpc_id"`

	ManagedClusterName string `json:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_nodes"`
	SubnetName         string `json:"subnet_name"`
	SubnetID           string `json:"subnet_id"`
	NetworkAclID       string `json:"network_acl_id"`

	GatewayID    string `json:"gateway_id"`
	RouteTableID string `json:"route_table_id"`

	InfoControlPlanes AWSStateVms `json:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database"`
	InfoLoadBalancer  AWSStateVm  `json:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
}
