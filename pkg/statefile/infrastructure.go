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

type BaseInfra struct {
	IsCompleted bool `json:"status" bson:"status"`

	SSHID      string `json:"ssh_id" bson:"ssh_id"`
	SSHUser    string `json:"ssh_usr" bson:"ssh_usr"`
	SSHKeyName string `json:"sshkey_name" bson:"sshkey_name"`

	KubernetesVer string `json:"k8s_version" bson:"k8s_version"`
}
