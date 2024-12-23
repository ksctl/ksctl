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

package civo

import (
	"fmt"
	"github.com/ksctl/ksctl/pkg/waiter"
	"time"

	"github.com/civo/civogo"

	"github.com/ksctl/ksctl/pkg/consts"
)

func (p *Provider) watchManagedCluster(id string, name string) error {

	expoBackoff := waiter.NewWaiter(
		10*time.Second,
		2,
		2*int(consts.CounterMaxWatchRetryCount),
	)

	var clusterDS *civogo.KubernetesCluster
	_err := expoBackoff.Run(
		p.ctx,
		p.l,
		func() (err error) {
			clusterDS, err = p.client.GetKubernetesCluster(id)
			return err
		},
		func() bool {
			return clusterDS.Ready
		},
		nil,
		func() error {
			p.l.Print(p.ctx, "cluster ready", "name", name)
			p.state.CloudInfra.Civo.B.IsCompleted = true
			p.state.ClusterKubeConfig = clusterDS.KubeConfig
			p.state.ClusterKubeConfigContext = name
			return p.store.Write(p.state)
		},
		fmt.Sprintf("Waiting for managed cluster %s to be ready", id),
	)
	if _err != nil {
		return _err
	}

	return nil
}

func (p *Provider) NewManagedCluster(noOfNodes int) error {

	name := <-p.chResName
	vmtype := <-p.chVMType

	p.l.Debug(p.ctx, "Printing", "name", name, "vmtype", vmtype)

	if len(p.state.CloudInfra.Civo.ManagedClusterID) != 0 {
		p.l.Print(p.ctx, "skipped managed cluster creation found", "id", p.state.CloudInfra.Civo.ManagedClusterID)

		if err := p.watchManagedCluster(p.state.CloudInfra.Civo.ManagedClusterID, name); err != nil {
			return err
		}

		return nil
	}

	configK8s := &civogo.KubernetesClusterConfig{
		KubernetesVersion: p.K8sVersion,
		Name:              name,
		Region:            p.Region,
		NumTargetNodes:    noOfNodes,
		TargetNodesSize:   vmtype,
		NetworkID:         p.state.CloudInfra.Civo.NetworkID,
		Applications:      p.apps, // make the use of application and cni via some method
		CNIPlugin:         p.cni,  // make it use install application in the civo
	}
	p.l.Debug(p.ctx, "Printing", "configManagedK8s", configK8s)

	resp, err := p.client.NewKubernetesClusters(configK8s)
	if err != nil {
		return err
	}

	p.state.CloudInfra.Civo.NoManagedNodes = noOfNodes
	p.state.BootstrapProvider = "managed"
	p.state.CloudInfra.Civo.ManagedNodeSize = vmtype
	p.state.CloudInfra.Civo.B.KubernetesVer = p.K8sVersion
	p.state.CloudInfra.Civo.ManagedClusterID = resp.ID

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	if err := p.watchManagedCluster(resp.ID, name); err != nil {
		return err
	}
	p.l.Success(p.ctx, "Created Managed cluster", "clusterID", p.state.CloudInfra.Civo.ManagedClusterID)
	return nil
}

func (p *Provider) DelManagedCluster() error {
	if len(p.state.CloudInfra.Civo.ManagedClusterID) == 0 {
		p.l.Print(p.ctx, "skipped network deletion found", "id", p.state.CloudInfra.Civo.ManagedClusterID)
		return nil
	}
	_, err := p.client.DeleteKubernetesCluster(p.state.CloudInfra.Civo.ManagedClusterID)
	if err != nil {
		return err
	}
	p.l.Success(p.ctx, "Deleted Managed cluster", "clusterID", p.state.CloudInfra.Civo.ManagedClusterID)
	p.state.CloudInfra.Civo.ManagedClusterID = ""
	p.state.CloudInfra.Civo.ManagedNodeSize = ""

	return p.store.Write(p.state)
}
