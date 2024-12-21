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

import "github.com/ksctl/ksctl/pkg/types/controllers/cloud"

type BaseK8sBootstrap struct {
	SSHInfo    cloud.SSHInfo `json:"cloud_ssh_info" bson:"cloud_ssh_info"`
	PublicIPs  Instances     `json:"cloud_public_ips" bson:"cloud_public_ips"`
	PrivateIPs Instances     `json:"cloud_private_ips" bson:"cloud_private_ips"`

	HAProxyVersion string `json:"haproxy_version" bson:"haproxy_version"`
	EtcdVersion    string `json:"etcd_version" bson:"etcd_version"`

	CACert   string `json:"ca_cert" bson:"ca_cert"`
	EtcdCert string `json:"etcd_cert" bson:"etcd_cert"`
	EtcdKey  string `json:"etcd_key" bson:"etcd_key"`
}
