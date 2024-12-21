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

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CredentialsDocument object which stores the credentials for each provider
type CredentialsDocument struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Aws           *CredentialsAws    `json:"aws,omitempty" bson:"aws,omitempty"`
	Azure         *CredentialsAzure  `json:"azure,omitempty" bson:"azure,omitempty"`
	Civo          *CredentialsCivo   `json:"civo,omitempty" bson:"civo,omitempty"`
	InfraProvider consts.KsctlCloud  `json:"cloud_provider" bson:"cloud_provider"`
}

// StorageDocument object which stores the state of infra and bootstrap in a doc
type StorageDocument struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	ClusterType string `json:"cluster_type" bson:"cluster_type" `
	Region      string `json:"region" bson:"region"`
	ClusterName string `json:"cluster_name" bson:"cluster_name"`

	InfraProvider     consts.KsctlCloud      `json:"cloud_provider" bson:"cloud_provider"`
	BootstrapProvider consts.KsctlKubernetes `json:"bootstrap_provider" bson:"bootstrap_provider"`

	CloudInfra               *InfrastructureState      `json:"cloud_infrastructure_state" bson:"cloud_infrastructure_state,omitempty"`
	K8sBootstrap             *KubernetesBootstrapState `json:"kubernetes_bootstrap_state" bson:"kubernetes_bootstrap_state,omitempty"`
	ClusterKubeConfig        string                    `json:"cluster_kubeconfig" bson:"cluster_kubeconfig"`
	ClusterKubeConfigContext string                    `json:"cluster_kubeconfig_context" bson:"cluster_kubeconfig_context"`

	SSHKeyPair SSHKeyPairState `json:"ssh_key_pair" bson:"ssh_key_pair"`

	Addons KubernetesAddons `json:"addons" bson:"addons"`
}

type InfrastructureState struct {
	Aws   *StateConfigurationAws   `json:"aws,omitempty" bson:"aws,omitempty"`
	Azure *StateConfigurationAzure `json:"azure,omitempty" bson:"azure,omitempty"`
	Civo  *StateConfigurationCivo  `json:"civo,omitempty" bson:"civo,omitempty"`
	Local *StateConfigurationLocal `json:"local,omitempty" bson:"local,omitempty"`
}

type KubernetesBootstrapState struct {
	B       BaseK8sBootstrap           `json:"b" bson:"b"`
	K3s     *StateConfigurationK3s     `json:"k3s,omitempty" bson:"k3s,omitempty"`
	Kubeadm *StateConfigurationKubeadm `json:"kubeadm,omitempty" bson:"kubeadm,omitempty"`
}

type SSHKeyPairState struct {
	PublicKey  string `json:"public_key" bson:"public_key"`
	PrivateKey string `json:"private_key" bson:"private_key"`
}

// Instances specific to each infrastructure like civo or azure
type Instances struct {
	ControlPlanes []string `json:"controlplanes" bson:"controlplanes"`
	WorkerPlanes  []string `json:"workerplanes" bson:"workerplanes"`
	DataStores    []string `json:"datastores" bson:"datastores"`
	LoadBalancer  string   `json:"loadbalancer" bson:"loadbalancer"`
}
