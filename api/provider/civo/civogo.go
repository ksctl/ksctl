package civo

import (
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"strings"
	"time"
)

type CivoGo interface {
	CreateNetwork(label string) (*civogo.NetworkResult, error)
	DeleteNetwork(id string) (*civogo.SimpleResponse, error)
	GetNetwork(id string) (*civogo.Network, error)

	NewFirewall(config *civogo.FirewallConfig) (*civogo.FirewallResult, error)
	DeleteFirewall(id string) (*civogo.SimpleResponse, error)

	NewSSHKey(name string, publicKey string) (*civogo.SimpleResponse, error)
	DeleteSSHKey(id string) (*civogo.SimpleResponse, error)

	CreateInstance(config *civogo.InstanceConfig) (*civogo.Instance, error)
	DeleteInstance(id string) (*civogo.SimpleResponse, error)
	GetInstance(id string) (*civogo.Instance, error)

	GetDiskImageByName(name string) (*civogo.DiskImage, error)

	InitClient(factory resources.StorageFactory, region string) error

	GetKubernetesCluster(id string) (*civogo.KubernetesCluster, error)
	NewKubernetesClusters(kc *civogo.KubernetesClusterConfig) (*civogo.KubernetesCluster, error)
	DeleteKubernetesCluster(id string) (*civogo.SimpleResponse, error)

	ListAvailableKubernetesVersions() ([]civogo.KubernetesVersion, error)

	ListRegions() ([]civogo.Region, error)
	ListInstanceSizes() ([]civogo.InstanceSize, error)
}

type CivoGoClient struct {
	client *civogo.Client
	region string
}

func ProvideMockCivoClient() CivoGo {
	return &CivoGoMockClient{}
}

func ProvideClient() CivoGo {
	return &CivoGoClient{}
}

func (client *CivoGoClient) ListAvailableKubernetesVersions() ([]civogo.KubernetesVersion, error) {
	return client.client.ListAvailableKubernetesVersions()
}

func (client *CivoGoClient) ListRegions() ([]civogo.Region, error) {
	return client.client.ListRegions()
}

func (client *CivoGoClient) ListInstanceSizes() ([]civogo.InstanceSize, error) {
	return client.client.ListInstanceSizes()
}

func (client *CivoGoClient) GetNetwork(id string) (*civogo.Network, error) {
	return client.client.GetNetwork(id)
}

func (client *CivoGoClient) GetKubernetesCluster(id string) (*civogo.KubernetesCluster, error) {
	return client.client.GetKubernetesCluster(id)
}

func (client *CivoGoClient) NewKubernetesClusters(kc *civogo.KubernetesClusterConfig) (*civogo.KubernetesCluster, error) {
	return client.client.NewKubernetesClusters(kc)
}

func (client *CivoGoClient) DeleteKubernetesCluster(id string) (*civogo.SimpleResponse, error) {
	return client.client.DeleteKubernetesCluster(id)
}

func (client *CivoGoClient) GetDiskImageByName(name string) (*civogo.DiskImage, error) {
	return client.client.GetDiskImageByName(name)
}

func (client *CivoGoClient) CreateNetwork(label string) (*civogo.NetworkResult, error) {
	return client.client.NewNetwork(label)
}

func (client *CivoGoClient) DeleteNetwork(id string) (*civogo.SimpleResponse, error) {
	return client.client.DeleteNetwork(id)
}

func (client *CivoGoClient) NewFirewall(config *civogo.FirewallConfig) (*civogo.FirewallResult, error) {
	return client.client.NewFirewall(config)
}

func (client *CivoGoClient) DeleteFirewall(id string) (*civogo.SimpleResponse, error) {
	return client.client.DeleteFirewall(id)
}

func (client *CivoGoClient) NewSSHKey(name string, publicKey string) (*civogo.SimpleResponse, error) {
	return client.client.NewSSHKey(strings.ToLower(name), publicKey)
}

func (client *CivoGoClient) DeleteSSHKey(id string) (*civogo.SimpleResponse, error) {
	return client.client.DeleteSSHKey(id)
}

func (client *CivoGoClient) CreateInstance(config *civogo.InstanceConfig) (*civogo.Instance, error) {
	return client.client.CreateInstance(config)
}

func (client *CivoGoClient) GetInstance(id string) (*civogo.Instance, error) {
	return client.client.GetInstance(id)
}

func (client *CivoGoClient) DeleteInstance(id string) (*civogo.SimpleResponse, error) {
	return client.client.DeleteInstance(id)
}

func (client *CivoGoClient) InitClient(factory resources.StorageFactory, region string) (err error) {
	client.client, err = civogo.NewClient(fetchAPIKey(factory), region)
	if err != nil {
		return
	}
	client.region = region
	return
}

// CivoGoMockClient ///////// Mock Client
type CivoGoMockClient struct {
	client *civogo.FakeClient
	region string
}

func (client *CivoGoMockClient) ListAvailableKubernetesVersions() ([]civogo.KubernetesVersion, error) {
	return []civogo.KubernetesVersion{
		{
			ClusterType: utils.K8S_K3S,
			Label:       "1.27.4-k3s1",
		},
		{
			ClusterType: utils.K8S_K3S,
			Label:       "1.27.1-k3s1",
		},
		{
			ClusterType: utils.K8S_K3S,
			Label:       "1.26.4-k3s1",
		},
	}, nil
}

func (client *CivoGoMockClient) ListRegions() ([]civogo.Region, error) {
	return []civogo.Region{
		{
			Name: "FAKE",
			Code: "LON1",
		},
		{
			Name: "FAKE",
			Code: "FRA1",
		},
		{
			Name: "FAKE",
			Code: "NYC1",
		},
	}, nil
}

func (client *CivoGoMockClient) ListInstanceSizes() ([]civogo.InstanceSize, error) {
	return []civogo.InstanceSize{
		{
			Name: "g3.small",
		},
		{
			Name: "fake.small",
		},
		{
			Name: "g4s.kube.small",
		},
	}, nil
}

func (client *CivoGoMockClient) GetNetwork(id string) (*civogo.Network, error) {
	return &civogo.Network{
		ID:      id,
		Default: false,
		Status:  "ACTIVE",
	}, nil
}

func (client *CivoGoMockClient) GetKubernetesCluster(id string) (*civogo.KubernetesCluster, error) {
	return &civogo.KubernetesCluster{
		ID:        id,
		NetworkID: "fake",
		Name:      "fake",
		KubeConfig: `
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /workdir/.minikube/ca.crt
    server: https://127.0.0.1:6443
  name: fake
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate-data: /workdir/.minikube/client.crt
    client-key-data: /workdir/.minikube/client.key`,
		Ready: true,
	}, nil
}

func (client *CivoGoMockClient) NewKubernetesClusters(kc *civogo.KubernetesClusterConfig) (*civogo.KubernetesCluster, error) {
	return &civogo.KubernetesCluster{
		ID:        "managed-k8s-213",
		Name:      kc.Name,
		NetworkID: kc.NetworkID,
		Version:   kc.KubernetesVersion,
		CreatedAt: time.Now(),
	}, nil
}

func (client *CivoGoMockClient) DeleteKubernetesCluster(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake kubernetes cluster deleted",
	}, nil
}

func (client *CivoGoMockClient) GetDiskImageByName(name string) (*civogo.DiskImage, error) {
	return &civogo.DiskImage{
		Name:  name,
		ID:    "disk-123",
		State: "ACTIVE",
	}, nil
}

func (client *CivoGoMockClient) CreateNetwork(label string) (*civogo.NetworkResult, error) {
	return &civogo.NetworkResult{
		Label:  label,
		ID:     "net-213",
		Result: "created fake network",
	}, nil
}

func (client *CivoGoMockClient) DeleteNetwork(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake network deleted",
	}, nil
}

func (client *CivoGoMockClient) NewFirewall(config *civogo.FirewallConfig) (*civogo.FirewallResult, error) {
	return &civogo.FirewallResult{
		ID:     "firewall-123",
		Name:   config.Name,
		Result: "fake firewall created",
	}, nil
}

func (client *CivoGoMockClient) DeleteFirewall(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake firewall deleted",
	}, nil
}

func (client *CivoGoMockClient) NewSSHKey(_, _ string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     "ssh-123",
		Result: "created fake ssh key",
	}, nil
}

func (client *CivoGoMockClient) DeleteSSHKey(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake ssh key deleted",
	}, nil
}

func (client *CivoGoMockClient) CreateInstance(config *civogo.InstanceConfig) (*civogo.Instance, error) {
	return &civogo.Instance{
		ID:         "vm-123",
		Region:     config.Region,
		PrivateIP:  "192.169.X.X",
		PublicIP:   "A.B.C.D",
		CreatedAt:  time.Now(),
		FirewallID: config.FirewallID,
		SSHKeyID:   config.SSHKeyID,
		Size:       config.Size,
		Hostname:   "fake-hostname",
		NetworkID:  config.NetworkID,
	}, nil
}

func (client *CivoGoMockClient) GetInstance(id string) (*civogo.Instance, error) {
	return &civogo.Instance{
		ID:        id,
		PrivateIP: "192.169.X.X",
		PublicIP:  "A.B.C.D",
		Hostname:  "fake-hostname",
		Status:    "ACTIVE",
	}, nil
}

func (client *CivoGoMockClient) DeleteInstance(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake vm deleted",
	}, nil
}

func (client *CivoGoMockClient) InitClient(factory resources.StorageFactory, region string) (err error) {
	client.client, err = civogo.NewFakeClient()
	if err != nil {
		return
	}
	client.region = region
	return
}
