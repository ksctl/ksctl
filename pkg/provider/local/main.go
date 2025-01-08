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

package local

import (
	"context"
	"encoding/json"

	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/provider"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
)

type Provider struct {
	l     logger.Logger
	ctx   context.Context
	state *statefile.StorageDocument
	store storage.Storage
	controller.Metadata

	cni               string
	tempDirKubeconfig string

	vmType  string
	resName string

	provider.Cloud

	client KindSDK
}

func NewClient(
	ctx context.Context,
	l logger.Logger,
	meta controller.Metadata,
	state *statefile.StorageDocument,
	storage storage.Storage,
	ClientOption func() KindSDK,
) (*Provider, error) {
	p := new(Provider)
	p.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, string(consts.CloudLocal))
	p.state = state
	p.Metadata = meta
	p.l = l
	p.client = ClientOption()
	p.store = storage

	p.l.Debug(p.ctx, "Printing", "LocalProvider", p)
	return p, nil
}

func (p *Provider) GetStateFile() (string, error) {
	cloudstate, err := json.Marshal(p.state)
	if err != nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			p.l.NewError(p.ctx, "failed to serialize the state", "Reason", err),
		)
	}
	return string(cloudstate), nil
}

func (p *Provider) InitState(operation consts.KsctlOperation) error {
	switch operation {
	case consts.OperationCreate:
		if err := p.isPresent(); err == nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				p.l.NewError(p.ctx, "already present", "name", p.ClusterName),
			)
		}
		p.l.Debug(p.ctx, "Fresh state!!")

		p.state.ClusterName = p.ClusterName
		p.state.Region = p.Region
		p.state.CloudInfra = &statefile.InfrastructureState{Local: &statefile.StateConfigurationLocal{}}
		p.state.InfraProvider = consts.CloudLocal
		p.state.ClusterType = string(consts.ClusterTypeMang)

		p.state.CloudInfra.Local.B.KubernetesVer = p.K8sVersion
	case consts.OperationDelete, consts.OperationGet:
		err := p.loadStateHelper()
		if err != nil {
			return err
		}
	}
	p.l.Debug(p.ctx, "initialized the state")
	return nil
}

func (p *Provider) Name(resName string) provider.Cloud {
	p.resName = resName
	return p
}

func (p *Provider) ManagedK8sVersion(ver string) provider.Cloud {
	p.l.Debug(p.ctx, "Printing", "k8sVersion", ver)
	p.K8sVersion = ver
	return p
}

func (p *Provider) GetRAWClusterInfos() ([]logger.ClusterDataForLogging, error) {

	var data []logger.ClusterDataForLogging
	clusters, err := p.store.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudLocal),
		consts.ClusterType: string(consts.ClusterTypeMang),
	})
	if err != nil {
		return nil, err
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, logger.ClusterDataForLogging{
				CloudProvider: consts.CloudLocal,
				Name:          v.ClusterName,
				Region:        v.Region,
				ClusterType:   K,

				NoMgt: v.CloudInfra.Local.Nodes,
				Mgt: logger.VMData{
					VMSize: v.CloudInfra.Local.ManagedNodeSize,
				},

				K8sDistro:  v.BootstrapProvider,
				K8sVersion: v.CloudInfra.Local.B.KubernetesVer,
			})
			p.l.Debug(p.ctx, "Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func (p *Provider) IsPresent() error {
	return p.isPresent()
}

func (p *Provider) VMType(_ string) provider.Cloud {
	p.vmType = "local_machine"
	return p
}

func (p *Provider) GetKubeconfig() (*string, error) {
	_read, err := p.store.Read()
	if err != nil {
		p.l.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig := _read.ClusterKubeConfig
	return &kubeconfig, nil
}
