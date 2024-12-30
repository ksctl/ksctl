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

//go:build testing_civo

package civo

import (
	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/consts"
	"time"
)

func ProvideClient() CloudSDK {
	return &CivoClient{}
}

type CivoClient struct {
	client *civogo.FakeClient
	region string
}

func (client *CivoClient) ListAvailableKubernetesVersions() ([]civogo.KubernetesVersion, error) {

	return []civogo.KubernetesVersion{
		{
			ClusterType: string(consts.K8sK3s),
			Label:       "1.27.4-k3s1",
		},
		{
			ClusterType: string(consts.K8sK3s),
			Label:       "1.27.1-k3s1",
		},
		{
			ClusterType: string(consts.K8sK3s),
			Label:       "1.26.4-k3s1",
		},
	}, nil
}

func (client *CivoClient) ListRegions() ([]civogo.Region, error) {

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

func (client *CivoClient) ListInstanceSizes() ([]civogo.InstanceSize, error) {

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

func (client *CivoClient) GetNetwork(id string) (*civogo.Network, error) {

	return &civogo.Network{
		ID:      id,
		Default: false,
		Status:  "Active",
	}, nil
}

func (client *CivoClient) GetKubernetesCluster(id string) (*civogo.KubernetesCluster, error) {
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

func (client *CivoClient) NewKubernetesClusters(kc *civogo.KubernetesClusterConfig) (*civogo.KubernetesCluster, error) {
	return &civogo.KubernetesCluster{
		ID:        "fake-k8s",
		Name:      kc.Name,
		NetworkID: kc.NetworkID,
		Version:   kc.KubernetesVersion,
		CreatedAt: time.Now(),
	}, nil
}

func (client *CivoClient) DeleteKubernetesCluster(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake kubernetes cluster deleted",
	}, nil
}

func (client *CivoClient) GetDiskImageByName(name string) (*civogo.DiskImage, error) {

	return &civogo.DiskImage{
		Name:  name,
		ID:    "disk-fake",
		State: "ACTIVE",
	}, nil
}

func (client *CivoClient) CreateNetwork(label string) (*civogo.NetworkResult, error) {

	return &civogo.NetworkResult{
		Label:  label,
		ID:     "fake-net",
		Result: "created fake network",
	}, nil
}

func (client *CivoClient) DeleteNetwork(id string) (*civogo.SimpleResponse, error) {

	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake network deleted",
	}, nil
}

func (client *CivoClient) NewFirewall(config *civogo.FirewallConfig) (*civogo.FirewallResult, error) {

	return &civogo.FirewallResult{
		ID:     "fake-fw",
		Name:   config.Name,
		Result: "fake firewall created",
	}, nil
}

func (client *CivoClient) DeleteFirewall(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake firewall deleted",
	}, nil
}

func (client *CivoClient) NewSSHKey(_, _ string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     "fake-ssh",
		Result: "created fake ssh key",
	}, nil
}

func (client *CivoClient) DeleteSSHKey(id string) (*civogo.SimpleResponse, error) {
	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake ssh key deleted",
	}, nil
}

func (client *CivoClient) CreateInstance(config *civogo.InstanceConfig) (*civogo.Instance, error) {

	return &civogo.Instance{
		ID:         "vm-fake",
		Region:     config.Region,
		PrivateIP:  "192.168.1.2",
		PublicIP:   "A.B.C.D",
		CreatedAt:  time.Now(),
		FirewallID: config.FirewallID,
		SSHKeyID:   config.SSHKeyID,
		Size:       config.Size,
		Hostname:   "fake-hostname",
		NetworkID:  config.NetworkID,
	}, nil
}

func (client *CivoClient) GetInstance(id string) (*civogo.Instance, error) {

	return &civogo.Instance{
		ID:        id,
		PrivateIP: "192.168.1.2",
		PublicIP:  "A.B.C.D",
		Hostname:  "fake-hostname",
		Status:    "ACTIVE",
	}, nil
}

func (client *CivoClient) DeleteInstance(id string) (*civogo.SimpleResponse, error) {

	return &civogo.SimpleResponse{
		ID:     id,
		Result: "fake vm deleted",
	}, nil
}

func (client *CivoClient) InitClient(_ *Provider, region string) (err error) {
	client.client, err = civogo.NewFakeClient()
	if err != nil {
		return
	}
	client.region = region
	return
}
