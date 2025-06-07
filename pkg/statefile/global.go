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
	"fmt"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ControllerTaskCode string

const (
	TaskCodeCreate      ControllerTaskCode = "create"
	TaskCodeDelete      ControllerTaskCode = "delete"
	TaskCodeGet         ControllerTaskCode = "get"
	TaskCodeScale       ControllerTaskCode = "scale"
	TaskCodeConfigAddon ControllerTaskCode = "configure_addon"
)

type ClusterState string

const (
	Fresh          ClusterState = "fresh"
	Creating       ClusterState = "creating"
	CreationFailed ClusterState = "creation_failed"

	Running ClusterState = "running"

	Configuring       ClusterState = "configuring"
	ConfiguringFailed ClusterState = "configuring_failed"

	Deleting       ClusterState = "deleting"
	DeletionFailed ClusterState = "deletion_failed"
)

// IsControllerOpValid checks if the operation is valid for the current cluster state.
func (s ClusterState) IsControllerOperationAllowed(operation ControllerTaskCode) error {
	err := func(_op ControllerTaskCode, _s ClusterState) error {
		return fmt.Errorf("operation %s is not allowed in state %s", _op, _s)
	}

	switch s {
	case Fresh:
		if operation != TaskCodeCreate {
			return err(operation, s)
		}

	case Creating, Deleting: // we cannot perform any operation while creating or deleting
		if operation != TaskCodeGet {
			return err(operation, s)
		}

	case CreationFailed: // we can retry creation or delete
		if operation != TaskCodeCreate && operation != TaskCodeDelete {
			return err(operation, s)
		}

	case DeletionFailed: // we can retry deletion
		if operation != TaskCodeDelete {
			return err(operation, s)
		}

	case Running:
		if operation != TaskCodeGet &&
			operation != TaskCodeScale &&
			operation != TaskCodeConfigAddon &&
			operation != TaskCodeDelete {
			return err(operation, s)
		}

	case Configuring:
		if operation != TaskCodeGet {
			return err(operation, s)
		}

	case ConfiguringFailed:
		if operation != TaskCodeGet &&
			operation != TaskCodeConfigAddon &&
			operation != TaskCodeScale &&
			operation != TaskCodeDelete {
			return err(operation, s)
		}

	default:
		return nil
	}
	return nil
}

type PlatformSpec struct {
	Team  string       `json:"team" bson:"team"`
	Owner string       `json:"owner" bson:"owner"`
	State ClusterState `json:"state" bson:"state"`
}

// StorageDocument object which stores the state of infra and bootstrap in a doc
type StorageDocument struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	PlatformSpec PlatformSpec `json:"platform" bson:"platform"`

	ClusterType string `json:"cluster_type" bson:"cluster_type" `
	Region      string `json:"region" bson:"region"`
	ClusterName string `json:"cluster_name" bson:"cluster_name"`

	InfraProvider     consts.KsctlCloud      `json:"cloud_provider" bson:"cloud_provider"`
	BootstrapProvider consts.KsctlKubernetes `json:"bootstrap_provider" bson:"bootstrap_provider"`
	Versions          ComponentVersions      `json:"versions" bson:"versions"`

	CloudInfra               *InfrastructureState      `json:"cloud_infrastructure_state" bson:"cloud_infrastructure_state,omitempty"`
	K8sBootstrap             *KubernetesBootstrapState `json:"kubernetes_bootstrap_state" bson:"kubernetes_bootstrap_state,omitempty"`
	ClusterKubeConfig        string                    `json:"cluster_kubeconfig" bson:"cluster_kubeconfig"`
	ClusterKubeConfigContext string                    `json:"cluster_kubeconfig_context" bson:"cluster_kubeconfig_context"`

	SSHKeyPair SSHKeyPairState `json:"ssh_key_pair" bson:"ssh_key_pair"`

	ProvisionerAddons SlimProvisionerAddons `json:"provisioner_addons,omitempty" bson:"provisioner_addons,omitempty"`
}

func NewStorageDocument(
	clusterName string,
	region string,
	cloud consts.KsctlCloud,
	clusterClass consts.KsctlClusterType,
	team string,
	owner string,
) *StorageDocument {
	return &StorageDocument{
		PlatformSpec: PlatformSpec{
			State: Fresh,
			Team:  team,
			Owner: owner,
		},
		ClusterName:   clusterName,
		Region:        region,
		InfraProvider: cloud,
		ClusterType:   string(clusterClass),
	}
}

type SlimProvisionerAddons struct {
	Apps []SlimProvisionerAddon `json:"apps" bson:"apps"`
	Cni  SlimProvisionerAddon   `json:"cni" bson:"cni"`
}

type SlimProvisionerAddon struct {
	// +required
	Name string `json:"name" bson:"name"`

	// +required
	For consts.KsctlKubernetes `json:"for" bson:"for"`

	// +optional
	Version *string `json:"version,omitempty" bson:"version,omitempty"`

	// +required for ksctl specific apps
	KsctlSpecificComponents map[string]KsctlSpecificComponent `json:"ksctl_specific_components,omitempty" bson:"ksctl_specific_components,omitempty"`
}

func (s SlimProvisionerAddon) String() string {
	ver := "<nil>"
	if s.Version != nil {
		ver = *s.Version
	}
	return fmt.Sprintf("Name: %s, For: %s, Version: %v, KsctlSpecificComponents: %+v", s.Name, s.For, ver, s.KsctlSpecificComponents)
}

type KsctlSpecificComponent struct {
	Version string `json:"version" bson:"version"`
}

type InfrastructureState struct {
	Aws   *StateConfigurationAws   `json:"aws,omitempty" bson:"aws,omitempty"`
	Azure *StateConfigurationAzure `json:"azure,omitempty" bson:"azure,omitempty"`
	Local *StateConfigurationLocal `json:"local,omitempty" bson:"local,omitempty"`
}

type KubernetesBootstrapState struct {
	B       BaseK8sBootstrap           `json:"b" bson:"b"`
	K3s     *StateConfigurationK3s     `json:"k3s,omitempty" bson:"k3s,omitempty"`
	Kubeadm *StateConfigurationKubeadm `json:"kubeadm,omitempty" bson:"kubeadm,omitempty"`
}

type ComponentVersions struct {
	K3s     *string `json:"k3s,omitempty" bson:"k3s,omitempty"`
	Kubeadm *string `json:"kubeadm,omitempty" bson:"kubeadm,omitempty"`
	Aks     *string `json:"aks,omitempty" bson:"aks,omitempty"`
	Eks     *string `json:"eks,omitempty" bson:"eks,omitempty"`
	Kind    *string `json:"kind,omitempty" bson:"kind,omitempty"`
	Etcd    *string `json:"etcd,omitempty" bson:"etcd,omitempty"`
	HAProxy *string `json:"haproxy,omitempty" bson:"haproxy,omitempty"`
}

type SSHKeyPairState struct {
	PublicKey  string `json:"public_key" bson:"public_key"`
	PrivateKey string `json:"private_key" bson:"private_key"`
}

type Instances struct {
	ControlPlanes []string `json:"controlplanes" bson:"controlplanes"`
	WorkerPlanes  []string `json:"workerplanes" bson:"workerplanes"`
	DataStores    []string `json:"datastores" bson:"datastores"`
	LoadBalancer  string   `json:"loadbalancer" bson:"loadbalancer"`
}
