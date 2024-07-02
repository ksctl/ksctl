package storage

type AWSStateVm struct {
	Vpc                    string `json:"vpc" bson:"vpc"`
	HostName               string `json:"name" bson:"name"`
	DiskSize               string `json:"disk_size" bson:"disk_size"`
	InstanceType           string `json:"instance_type" bson:"instance_type"`
	InstanceID             string `json:"instance_id" bson:"instance_id"`
	Subnet                 string `json:"subnet" bson:"subnet"`
	NetworkSecurityGroupID string `json:"network_security_group_id" bson:"network_security_group_id"`
	PublicIP               string `json:"public_ip" bson:"public_ip"`
	PrivateIP              string `json:"private_ip" bson:"private_ip"`
	NetworkInterfaceId     string `json:"network_interface_id" bson:"network_interface_id"`
	VMSize                 string `json:"vm_size" bson:"vm_size"`
}

type CredentialsAws struct {
	AccessKeyId     string `json:"access_key_id" bson:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" bson:"secret_access_key"`
}

type AWSStateVms struct {
	HostNames               []string `json:"names" bson:"name"`
	DiskNames               []string `json:"disk_name" bson:"disk_name"`
	InstanceIds             []string `json:"instance_id" bson:"instance_id"`
	PrivateIPs              []string `json:"private_ip" bson:"private_ip"`
	PublicIPs               []string `json:"public_ip" bson:"public_ip"`
	NetworkInterfaceIDs     []string `json:"network_interface_id" bson:"network_interface_id"`
	SubnetNames             []string `json:"subnet_name" bson:"subnet_name"`
	SubnetIDs               []string `json:"subnet_id" bson:"subnet_id"`
	NetworkSecurityGroupIDs string   `json:"network_security_group_ids" bson:"network_security_group_ids"`
	VMSizes                 []string `json:"vm_sizes" bson:"vm_sizes"` // keeping a dynamic sizes for autoscaler feature
}

type StateConfigurationAws struct {
	B BaseInfra `json:"b" bson:"b"`

	IsCompleted bool

	VpcName string `json:"vpc" bson:"vpc"`
	VpcId   string `json:"vpc_id" bson:"vpc_id"`
	VpcCidr string `json:"vpc_cidr" bson:"vpc_cidr"`

	IamRoleNameCN string `json:"iam_role_name" bson:"iam_role_name"`
	IamRoleArnCN  string `json:"iam_role_arn" bson:"iam_role_arn"`

	IamRoleNameWP string `json:"iam_role_name_wp" bson:"iam_role_name_wp"`
	IamRoleArnWP  string `json:"iam_role_arn_wp" bson:"iam_role_arn_wp"`

	ManagedClusterName   string `json:"managed_cluster_name" bson:"managed_cluster_name"`
	ManagedNodeGroupName string `json:"managed_node_group_name" bson:"managed_node_group_name"`
	NoManagedNodes       int    `json:"no_managed_nodes" bson:"no_managed_nodes"`
	ManagedNodeSize      string `json:"managed_node_size" bson:"managed_node_size"`
	ManagedNodeGroupArn  string `json:"managed_node_group_arns" bson:"managed_node_group_arns"`
	ManagedClusterArn    string `json:"managed_cluster_arn" bson:"managed_cluster_arn"`

	SubnetNames  []string `json:"subnet_names" bson:"subnet_names"`
	SubnetIDs    []string `json:"subnet_id" bson:"subnet_ids"`
	NetworkAclID string   `json:"network_acl_id" bson:"network_acl_id"`

	GatewayID    string `json:"gateway_id" bson:"gateway_id"`
	RouteTableID string `json:"route_table_id" bson:"route_table_id"`

	InfoControlPlanes AWSStateVms `json:"info_control_planes" bson:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes" bson:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database" bson:"info_database"`
	InfoLoadBalancer  AWSStateVm  `json:"info_load_balancer" bson:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro" bson:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version" bson:"k8s_version"`
}
