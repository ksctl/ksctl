// Copyright 2024 ksctl
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

package statefile

type AWSStateVm struct {
	HostName               string `json:"name" bson:"name"`
	InstanceID             string `json:"instance_id" bson:"instance_id"`
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
	InstanceIds             []string `json:"instance_id" bson:"instance_id"`
	PrivateIPs              []string `json:"private_ip" bson:"private_ip"`
	PublicIPs               []string `json:"public_ip" bson:"public_ip"`
	NetworkInterfaceIDs     []string `json:"network_interface_id" bson:"network_interface_id"`
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

	ManagedClusterName     string `json:"managed_cluster_name" bson:"managed_cluster_name"`
	ManagedNodeGroupName   string `json:"managed_node_group_name" bson:"managed_node_group_name"`
	NoManagedNodes         int    `json:"no_managed_nodes" bson:"no_managed_nodes"`
	ManagedNodeSize        string `json:"managed_node_size" bson:"managed_node_size"`
	ManagedNodeGroupArn    string `json:"managed_node_group_arns" bson:"managed_node_group_arns"`
	ManagedClusterArn      string `json:"managed_cluster_arn" bson:"managed_cluster_arn"`
	ManagedNodeGroupVmSize string `json:"managed_node_group_vm_size" bson:"managed_node_group_vm_size"`

	SubnetNames  []string `json:"subnet_names" bson:"subnet_names"`
	SubnetIDs    []string `json:"subnet_id" bson:"subnet_ids"`
	NetworkAclID string   `json:"network_acl_id" bson:"network_acl_id"`

	GatewayID    string `json:"gateway_id" bson:"gateway_id"`
	RouteTableID string `json:"route_table_id" bson:"route_table_id"`

	InfoControlPlanes AWSStateVms `json:"info_control_planes" bson:"info_control_planes"`
	InfoWorkerPlanes  AWSStateVms `json:"info_worker_planes" bson:"info_worker_planes"`
	InfoDatabase      AWSStateVms `json:"info_database" bson:"info_database"`
	InfoLoadBalancer  AWSStateVm  `json:"info_load_balancer" bson:"info_load_balancer"`
}
