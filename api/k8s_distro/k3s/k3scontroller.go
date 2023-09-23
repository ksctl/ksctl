package k3s

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
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

	ClusterName string           `json:"cluster_name"`
	Region      string           `json:"region"`
	ClusterType KsctlClusterType `json:"cluster_type"`
	ClusterDir  string           `json:"cluster_dir"`
	Provider    KsctlCloud       `json:"provider"`
}

var (
	k8sState *StateConfiguration
)

func ReturnK3sStruct() *K3sDistro {
	return &K3sDistro{
		SSHInfo: &utils.SSHPayload{},
	}
}

type K3sDistro struct {
	K3sVer string
	// it will be used for SSH
	SSHInfo utils.SSHCollection
}

// GetStateFiles implements resources.DistroFactory.
func (*K3sDistro) GetStateFile(resources.StorageFactory) (string, error) {
	state, err := json.Marshal(k8sState)
	if err != nil {
		return "", err
	}
	return string(state), nil
}

const (
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("k8s-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

func scriptKUBECONFIG() string {
	return `#!/bin/bash
sudo cat /etc/rancher/k3s/k3s.yaml`
}

// InitState implements resources.DistroFactory.
// try to achieve deepCopy
func (k3s *K3sDistro) InitState(cloudState cloud.CloudResourceState, storage resources.StorageFactory, operation KsctlOperation) error {
	// add the nil check here as well
	path := utils.GetPath(CLUSTER_PATH, cloudState.Metadata.Provider, cloudState.Metadata.ClusterType, cloudState.Metadata.ClusterDir, STATE_FILE_NAME)

	switch operation {
	case OPERATION_STATE_CREATE:
		// add  a flag of completion check
		k8sState = &StateConfiguration{}
		k8sState.DataStoreEndPoint = ""
		k8sState.K3sToken = ""
	case OPERATION_STATE_GET:
		raw, err := storage.Path(path).Load()
		if err != nil {
			return err
		}
		err = json.Unmarshal(raw, &k8sState)
		if err != nil {
			return err
		}

	}
	k8sState.PublicIPs.ControlPlanes = cloudState.IPv4ControlPlanes
	k8sState.PrivateIPs.ControlPlanes = cloudState.PrivateIPv4ControlPlanes

	k8sState.PublicIPs.DataStores = cloudState.IPv4DataStores
	k8sState.PrivateIPs.DataStores = cloudState.PrivateIPv4DataStores

	k8sState.PublicIPs.WorkerPlanes = cloudState.IPv4WorkerPlanes

	k8sState.PublicIPs.Loadbalancer = cloudState.IPv4LoadBalancer
	k8sState.PrivateIPs.Loadbalancer = cloudState.PrivateIPv4LoadBalancer
	k8sState.SSHInfo = cloudState.SSHState

	k3s.SSHInfo.LocPrivateKey(k8sState.SSHInfo.PathPrivateKey)
	k3s.SSHInfo.Username(k8sState.SSHInfo.UserName)

	k8sState.ClusterName = cloudState.Metadata.ClusterName
	k8sState.Region = cloudState.Metadata.Region
	k8sState.Provider = cloudState.Metadata.Provider
	k8sState.ClusterDir = cloudState.Metadata.ClusterDir
	k8sState.ClusterType = cloudState.Metadata.ClusterType
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

func (k3s *K3sDistro) Version(ver string) resources.DistroFactory {
	if isValidK3sVersion(ver) {
		// valid
		k3s.K3sVer = fmt.Sprintf("v%s+k3s1", ver)
		return k3s
	}
	return nil
}
