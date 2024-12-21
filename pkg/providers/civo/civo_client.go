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

//go:build !testing_civo

package civo

import (
	"errors"
	"github.com/civo/civogo"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"strings"
)

type CivoClient struct {
	client *civogo.Client
	region string
}

func ProvideClient() CivoGo {
	return &CivoClient{}
}

func (client *CivoClient) ListAvailableKubernetesVersions() ([]civogo.KubernetesVersion, error) {
	v, err := client.client.ListAvailableKubernetesVersions()
	if err != nil {
		return nil, ksctlErrors.ErrInvalidVersion.Wrap(
			log.NewError(civoCtx, "failed to get the valid managed kubernetes versions", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) ListRegions() ([]civogo.Region, error) {
	v, err := client.client.ListRegions()
	if err != nil {
		return nil, ksctlErrors.ErrInvalidCloudRegion.Wrap(
			log.NewError(civoCtx, "failed to get the valid regions", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) ListInstanceSizes() ([]civogo.InstanceSize, error) {
	v, err := client.client.ListInstanceSizes()
	if err != nil {
		return nil, ksctlErrors.ErrInvalidCloudVMSize.Wrap(
			log.NewError(civoCtx, "failed to get the valid virtual machine sizes", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) GetNetwork(id string) (*civogo.Network, error) {
	v, err := client.client.GetNetwork(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed to get network", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) GetKubernetesCluster(id string) (*civogo.KubernetesCluster, error) {
	v, err := client.client.GetKubernetesCluster(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed get management kubernetes cluster", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) NewKubernetesClusters(kc *civogo.KubernetesClusterConfig) (*civogo.KubernetesCluster, error) {
	v, err := client.client.NewKubernetesClusters(kc)
	if err != nil {
		if errors.Is(err, civogo.DatabaseKubernetesClusterDuplicateError) {
			return nil, ksctlErrors.ErrDuplicateRecords.Wrap(
				log.NewError(civoCtx, "failed to ", err.Error()),
			)
		}
		if errors.Is(err, civogo.AuthenticationFailedError) {
			return nil, ksctlErrors.ErrFailedCloudAccountAuth.Wrap(
				log.NewError(civoCtx, err.Error()),
			)
		}
		if errors.Is(err, civogo.UnknownError) {
			return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
				log.NewError(civoCtx, "failed to create management kubernetes cluster", "Reason", err),
			)
		}
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(civoCtx, "failed to create management kubernetes cluster", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) DeleteKubernetesCluster(id string) (*civogo.SimpleResponse, error) {
	v, err := client.client.DeleteKubernetesCluster(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed to delete management kubernetes cluster", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) GetDiskImageByName(name string) (*civogo.DiskImage, error) {
	v, err := client.client.GetDiskImageByName(name)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed get diskImage", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) CreateNetwork(label string) (*civogo.NetworkResult, error) {
	v, err := client.client.NewNetwork(label)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed create network", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) DeleteNetwork(id string) (*civogo.SimpleResponse, error) {
	v, err := client.client.DeleteNetwork(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed delete network", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) NewFirewall(config *civogo.FirewallConfig) (*civogo.FirewallResult, error) {
	v, err := client.client.NewFirewall(config)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed create firewall", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) DeleteFirewall(id string) (*civogo.SimpleResponse, error) {
	v, err := client.client.DeleteFirewall(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed delete firewall", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) NewSSHKey(name string, publicKey string) (*civogo.SimpleResponse, error) {
	v, err := client.client.NewSSHKey(strings.ToLower(name), publicKey)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed create sshkey", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) DeleteSSHKey(id string) (*civogo.SimpleResponse, error) {
	v, err := client.client.DeleteSSHKey(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed delete sshkey", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) CreateInstance(config *civogo.InstanceConfig) (*civogo.Instance, error) {
	v, err := client.client.CreateInstance(config)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed create vm", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) GetInstance(id string) (*civogo.Instance, error) {
	v, err := client.client.GetInstance(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed get vm", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) DeleteInstance(id string) (*civogo.SimpleResponse, error) {
	v, err := client.client.DeleteInstance(id)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKsctlClusterOperation.Wrap(
			log.NewError(civoCtx, "failed delete vm", "Reason", err),
		)
	}
	return v, nil
}

func (client *CivoClient) InitClient(factory types.StorageFactory, region string) (err error) {
	apiKey, err := fetchAPIKey(factory)
	if err != nil {
		return err
	}
	client.client, err = civogo.NewClient(apiKey, region)
	if err != nil {
		err = ksctlErrors.ErrInternal.Wrap(
			log.NewError(civoCtx, "Failed Init civo client", "Reason", err),
		)
		return
	}
	client.region = region
	return
}
