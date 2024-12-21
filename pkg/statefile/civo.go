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

package statefile

type CivoStateVMs struct {
	VMIDs      []string `json:"vm_ids" bson:"vm_ids"`
	PrivateIPs []string `json:"private_ips" bson:"private_ips"`
	PublicIPs  []string `json:"public_ips" bson:"public_ips"`
	Hostnames  []string `json:"hostnames" bson:"hostnames"`
	VMSizes    []string `json:"vm_sizes" bson:"vm_sizes"` // keeping a dynamic sizes for autoscaler feature
}

type CivoStateVM struct {
	VMID      string `json:"vm_id" bson:"vm_id"`
	PrivateIP string `json:"private_ip" bson:"private_ip"`
	PublicIP  string `json:"public_ip" bson:"public_ip"`
	HostName  string `json:"hostname" bson:"hostname"`
	VMSize    string `json:"vm_size" bson:"vm_size"`
}

type CredentialsCivo struct {
	Token string `json:"token" bson:"token"`
}

type StateConfigurationCivo struct {
	B BaseInfra `json:"b" bson:"b"`

	ManagedClusterID string `json:"managed_cluster_id" bson:"managed_cluster_id"`
	NoManagedNodes   int    `json:"no_managed_cluster_nodes" bson:"no_managed_cluster_nodes"`
	ManagedNodeSize  string `json:"managed_node_size" bson:"managed_node_size"`

	FirewallIDControlPlanes string `json:"fwidcontrolplanenode" bson:"fwidcontrolplanenode"`
	FirewallIDWorkerNodes   string `json:"fwidworkernode" bson:"fwidworkernode"`
	FirewallIDLoadBalancer  string `json:"fwidloadbalancenode" bson:"fwidloadbalancenode"`
	FirewallIDDatabaseNodes string `json:"fwiddatabasenode" bson:"fwiddatabasenode"`
	NetworkID               string `json:"clusternetworkid" bson:"clusternetworkid"`
	NetworkCIDR             string `json:"clusternetworkcidr" bson:"clusternetworkcidr"`

	InfoControlPlanes CivoStateVMs `json:"info_control_planes" bson:"info_control_planes"`
	InfoWorkerPlanes  CivoStateVMs `json:"info_worker_planes" bson:"info_worker_planes"`
	InfoDatabase      CivoStateVMs `json:"info_database" bson:"info_database"`
	InfoLoadBalancer  CivoStateVM  `json:"info_load_balancer" bson:"info_load_balancer"`
}
