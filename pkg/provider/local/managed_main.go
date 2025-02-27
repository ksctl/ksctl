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
	"os"
	"path/filepath"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func GetManagedCNIAddons() (addons.ClusterAddons, string) {
	return addons.ClusterAddons{
		{
			Name:   string(consts.CNINone),
			Label:  string(consts.K8sKind),
			IsCNI:  true,
			Config: nil,
		},
		{
			Name:   "kindnet",
			Label:  string(consts.K8sKind),
			IsCNI:  true,
			Config: nil,
		},
	}, "kindnet"
}

func (p *Provider) ManagedAddons(s addons.ClusterAddons) (externalCNI bool) {
	p.l.Debug(p.ctx, "Printing", "cni", s)
	clusterAddons := s.GetAddons(string(consts.K8sKind))

	p.managedAddonCNI = "kindnet" // Default: value
	externalCNI = false

	for _, addon := range clusterAddons {
		if addon.IsCNI {
			switch addon.Name {
			case string(consts.CNINone):
				p.managedAddonCNI = "none"
				externalCNI = true
			case "kindnet":
				p.managedAddonCNI = addon.Name
				externalCNI = false
			}
		} else {
			continue
		}
	}

	return
}

func (p *Provider) DelManagedCluster() error {

	_path := filepath.Join(p.tempDirKubeconfig, "kubeconfig")
	p.client.NewProvider(p, nil)
	if len(p.tempDirKubeconfig) == 0 {
		var err error
		p.tempDirKubeconfig, err = os.MkdirTemp("", p.ClusterName+"*")
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.l.NewError(p.ctx, "mkdirTemp", "Reason", err),
			)
		}
		if err := os.WriteFile(_path,
			[]byte(p.state.ClusterKubeConfig), 0755); err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.l.NewError(p.ctx, "failed to write file", "Reason", err),
			)
		}
		defer func() {
			_ = os.RemoveAll(p.tempDirKubeconfig)
		}()
	}

	p.l.Print(p.ctx, "Deleting the managed cluster")

	if err := p.client.Delete(p.ClusterName, _path); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			p.l.NewError(p.ctx, "failed to delete cluster", "Reason", err),
		)
	}

	if err := p.store.DeleteCluster(); err != nil {
		return err
	}

	return nil
}

func (p *Provider) NewManagedCluster(noOfNodes int) error {

	vmType := p.vmType

	p.client.NewProvider(p, nil)

	cni := false
	if consts.KsctlValidCNIPlugin(p.managedAddonCNI) == consts.CNINone {
		cni = true
	} else {
		p.state.ProvisionerAddons.Cni = statefile.SlimProvisionerAddon{
			Name: p.managedAddonCNI,
			For:  consts.K8sKind,
		}
	}

	p.l.Print(p.ctx, "Creating a new managed cluster")

	withConfig, err := p.configOption(noOfNodes, cni)
	if err != nil {
		return err
	}

	p.state.Versions.Kind = utilities.Ptr(p.K8sVersion)
	p.state.CloudInfra.Local.Nodes = noOfNodes

	p.state.BootstrapProvider = consts.K8sKind
	p.state.CloudInfra.Local.ManagedNodeSize = vmType

	Wait := 50 * time.Second

	p.tempDirKubeconfig, err = os.MkdirTemp("", p.ClusterName+"*")
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			p.l.NewError(p.ctx, "mkdirTemp", "Reason", err),
		)
	}

	ConfigHandler := func() string {
		_path, err := p.createNecessaryConfigs(p.tempDirKubeconfig)
		if err != nil {
			p.l.Error("rollback Cannot continue 😢")
			err = p.DelManagedCluster()
			if err != nil {
				p.l.Error("failed to perform cleanup", "Reason", err)
				return "" // asumming it never comes here
			}
		}
		return _path
	}
	Image := "kindest/node:v" + p.K8sVersion

	if err := p.client.Create(p.ClusterName, withConfig, Image, Wait, ConfigHandler); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlClusterOperation,
			p.l.NewError(p.ctx, "failed to create cluster", "err", err),
		)
	}

	_path := filepath.Join(p.tempDirKubeconfig, "kubeconfig")

	data, err := os.ReadFile(_path)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrKubeconfigOperations,
			p.l.NewError(p.ctx, "failed to read kubeconfig", "Reason", err),
		)
	}

	p.state.ClusterKubeConfig = string(data)
	p.state.ClusterKubeConfigContext = "kind-" + p.ClusterName
	p.state.CloudInfra.Local.B.IsCompleted = true

	if err := p.store.Write(p.state); err != nil {
		return err
	}
	_ = os.RemoveAll(p.tempDirKubeconfig) // remove the temp directory

	p.l.Success(p.ctx, "Created a new managed cluster", "name", p.ClusterName)

	return nil
}
