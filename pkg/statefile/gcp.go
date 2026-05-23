// Copyright 2026 Ksctl Authors
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

type GcpStateVMs struct {
	Names      []string `json:"names" bson:"names"`
	PrivateIPs []string `json:"private_ips" bson:"private_ips"`
	PublicIPs  []string `json:"public_ips" bson:"public_ips"`
	Hostnames  []string `json:"hostnames" bson:"hostnames"`
	VMSizes    []string `json:"vm_sizes" bson:"vm_sizes"` // keeping a dynamic sizes for autoscaler feature
}

type GcpStateVM struct {
	Name      string `json:"name" bson:"name"`
	PrivateIP string `json:"private_ip" bson:"private_ip"`
	PublicIP  string `json:"public_ip" bson:"public_ip"`
	HostName  string `json:"hostname" bson:"hostname"`
	VMSize    string `json:"vm_size" bson:"vm_size"`
}

type CredentialsGcp struct {
	GcpServiceAccountJSON string `json:"gcp_service_account_json"`
}

type StateConfigurationGcp struct {
	B BaseInfra `json:"b" bson:"b"`

	ProjectID          string `json:"project_id" bson:"project_id"`
	Zone               string `json:"zone" bson:"zone"`
	ManagedClusterName string `json:"managed_cluster_name" bson:"managed_cluster_name"`
}
