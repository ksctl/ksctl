package k3s

import (
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"
)

type Instances struct {
	ControlPlanes []string `json:"controlplanes"`
	WorkerPlanes  []string `json:"workerplanes"`
	DataStores    []string `json:"datastores"`
	Loadbalancer  string   `json:"loadbalancer"`
}

type StateConfiguration struct {
	K3sToken          string        `json:"k3s_token"`
	DataStoreEndPoint string        `json:"datastore_endpoint"`
	SSHInfo           cloud.SSHInfo `json:"cloud_ssh_info"` // contains data from cloud
	PublicIPs         Instances     `json:"cloud_public_ips"`
	PrivateIPs        Instances     `json:"cloud_private_ips"`

	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`
	ClusterType string `json:"cluster_type"`
	ClusterDir  string `json:"cluster_dir"`
	Provider    string `json:"provider"`
}

type K3sDistro struct {
	Version string // FIXME: Add k3s version support
	// it will be used for SSH
	SSHInfo utils.SSHCollection
}

const (
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("k8s-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

func scriptKUBECONFIG() string {
	return `#!/bin/bash
cat /etc/rancher/k3s/k3s.yaml`
}

// InitState implements resources.DistroFactory.
// try to achieve deepCopy
func (k3s *K3sDistro) InitState(cloudState cloud.CloudResourceState, storage resources.StorageFactory) error {
	// add the nil check here as well
	// TODO: first check if the cluster already exist and then add worerkplane that adding more woerkplane feature
	k8sState = &StateConfiguration{}

	k8sState.PublicIPs.ControlPlanes = cloudState.IPv4ControlPlanes
	k8sState.PrivateIPs.ControlPlanes = cloudState.PrivateIPv4ControlPlanes

	k8sState.PublicIPs.DataStores = cloudState.IPv4DataStores
	k8sState.PrivateIPs.DataStores = cloudState.PrivateIPv4DataStores

	k8sState.PublicIPs.WorkerPlanes = cloudState.IPv4WorkerPlanes

	k8sState.PublicIPs.Loadbalancer = cloudState.IPv4LoadBalancer
	k8sState.PrivateIPs.Loadbalancer = cloudState.PrivateIPv4LoadBalancer

	k8sState.DataStoreEndPoint = ""
	k8sState.SSHInfo = cloudState.SSHState
	k8sState.K3sToken = ""

	k8sState.ClusterName = cloudState.Metadata.ClusterName
	k8sState.Region = cloudState.Metadata.Region
	k8sState.Provider = cloudState.Metadata.Provider
	k8sState.ClusterDir = cloudState.Metadata.ClusterDir
	k8sState.ClusterType = cloudState.Metadata.ClusterType

	k3s.SSHInfo.LocPrivateKey(k8sState.SSHInfo.PathPrivateKey)
	k3s.SSHInfo.Username(k8sState.SSHInfo.UserName)

	path := utils.GetPath(utils.CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err := saveStateHelper(storage, path)
	if err != nil {
		return fmt.Errorf("[k3s] failed to Initialized state from Cloud reason: %v", err)
	}

	storage.Logger().Success("[k3s] Initialized state from Cloud")
	return nil
}

// InstallApplication implements resources.DistroFactory.
func (*K3sDistro) InstallApplication(state resources.StorageFactory) error {
	panic("unimplemented")
}

var (
	k8sState *StateConfiguration
)

func ReturnK3sStruct() *K3sDistro {
	return &K3sDistro{
		SSHInfo: &utils.SSHPayload{},
	}
}
