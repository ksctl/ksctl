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

import "time"

type StateConfigurationKubeadm struct {
	KubeadmVersion string `json:"kubeadm_version" bson:"kubeadm_version"`

	CertificateKey string `json:"certificate_key" bson:"certificate_key"`

	BootstrapToken string `json:"bootstrap_token" bson:"bootstrap_token"`

	BootstrapTokenExpireTimeUtc time.Time `json:"bootstrap_token_expire_time_utc" bson:"bootstrap_token_expire_time_utc"`

	DiscoveryTokenCACertHash string `json:"discorvery_token_ca_cert_hash" bson:"discorvery_token_ca_cert_hash"`
}
