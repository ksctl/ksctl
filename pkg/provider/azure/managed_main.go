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

package azure

import (
	"encoding/json"
	"os"

	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func GetManagedCNIAddons() (addons.ClusterAddons, string) {
	return addons.ClusterAddons{
		{
			Name:   string(consts.CNINone),
			Label:  string(consts.K8sAks),
			IsCNI:  true,
			Config: nil,
		},
		{
			Name:   "azure",
			Label:  string(consts.K8sAks),
			IsCNI:  true,
			Config: nil,
		},
	}, "azure"
}

func (p *Provider) ManagedAddons(s addons.ClusterAddons) (externalCNI bool) {

	p.l.Debug(p.ctx, "Printing", "cni", s)
	clusterAddons := s.GetAddons(string(consts.K8sAks))

	p.managedAddonCNI = "azure" // Default: value
	externalCNI = false

	for _, addon := range clusterAddons {
		if addon.IsCNI {
			switch addon.Name {
			case string(consts.CNINone):
				p.managedAddonCNI = "none"
				externalCNI = true
			case "azure":
				p.managedAddonCNI = addon.Name
				externalCNI = false
			}
		} else {
			var v map[string]*string

			if addon.Config != nil {

				v = make(map[string]*string)

				if err := json.Unmarshal([]byte(*addon.Config), &v); err != nil {
					p.l.Warn(p.ctx, "failed to unmarshal addon config", "addonName", addon.Name, "config", *addon.Config, "error", err)
				}
			} else {
				p.l.Warn(p.ctx, "empty addon config", "addonName", addon.Name)
			}

			p.managedAddonApp = make(map[string]map[string]*string)
			p.managedAddonApp[addon.Name] = v
		}
	}

	return
}

func (p *Provider) DelManagedCluster() error {
	if len(p.state.CloudInfra.Azure.ManagedClusterName) == 0 {
		p.l.Print(p.ctx, "skipped already deleted AKS cluster")
		return nil
	}

	pollerResp, err := p.client.BeginDeleteAKS(p.state.CloudInfra.Azure.ManagedClusterName, nil)
	if err != nil {
		return err
	}
	p.l.Print(p.ctx, "Deleting AKS cluster...", "name", p.state.CloudInfra.Azure.ManagedClusterName)

	_, err = p.client.PollUntilDoneDelAKS(p.ctx, pollerResp, nil)
	if err != nil {
		return err
	}

	p.l.Success(p.ctx, "Deleted the AKS cluster", "name", p.state.CloudInfra.Azure.ManagedClusterName)

	p.state.CloudInfra.Azure.ManagedClusterName = ""
	p.state.CloudInfra.Azure.ManagedNodeSize = ""
	return p.store.Write(p.state)
}

func (p *Provider) NewManagedCluster(noOfNodes int) error {
	name := <-p.chResName
	vmtype := <-p.chVMType

	p.l.Debug(p.ctx, "Printing", "name", name, "vmtype", vmtype)

	if len(p.state.CloudInfra.Azure.ManagedClusterName) != 0 {
		p.l.Print(p.ctx, "skipped already created AKS cluster", "name", p.state.CloudInfra.Azure.ManagedClusterName)
		return nil
	}

	p.state.CloudInfra.Azure.NoManagedNodes = noOfNodes
	p.state.Versions.Aks = utilities.Ptr(p.K8sVersion)
	p.state.BootstrapProvider = consts.K8sAks

	isKeda := false

	computedAddons := func() map[string]*armcontainerservice.ManagedClusterAddonProfile {
		addonProfiles := make(map[string]*armcontainerservice.ManagedClusterAddonProfile)

		vv := make([]statefile.SlimProvisionerAddon, 0, len(p.managedAddonApp))

		for k, v := range p.managedAddonApp {
			vv = append(vv, statefile.SlimProvisionerAddon{
				Name: k,
				For:  consts.K8sAks,
			})

			if k != "keda" {
				addonProfiles[k] = &armcontainerservice.ManagedClusterAddonProfile{
					Enabled: utilities.Ptr(true),
					Config:  v,
				}
			} else {
				isKeda = true
			}
		}

		p.state.ProvisionerAddons.Apps = vv
		return addonProfiles
	}

	if p.managedAddonCNI != string(consts.CNINone) {
		p.state.ProvisionerAddons.Cni = statefile.SlimProvisionerAddon{
			Name: p.managedAddonCNI,
			For:  consts.K8sAks,
		}
	}

	parameter := armcontainerservice.ManagedCluster{
		Location: utilities.Ptr(p.state.Region),
		SKU: &armcontainerservice.ManagedClusterSKU{
			Name: utilities.Ptr(armcontainerservice.ManagedClusterSKUNameBase),
			Tier: utilities.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
		},
		Properties: &armcontainerservice.ManagedClusterProperties{
			AddonProfiles: computedAddons(),

			DNSPrefix:         utilities.Ptr("aksgosdk"),
			KubernetesVersion: utilities.Ptr(p.K8sVersion),
			NetworkProfile: &armcontainerservice.NetworkProfile{
				NetworkPlugin: utilities.Ptr(armcontainerservice.NetworkPlugin(p.managedAddonCNI)),
			},
			AutoUpgradeProfile: &armcontainerservice.ManagedClusterAutoUpgradeProfile{
				NodeOSUpgradeChannel: utilities.Ptr(armcontainerservice.NodeOSUpgradeChannelNodeImage),
				UpgradeChannel:       utilities.Ptr(armcontainerservice.UpgradeChannelPatch),
			},
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					Name:              utilities.Ptr("askagent"),
					Count:             utilities.Ptr(int32(noOfNodes)),
					VMSize:            utilities.Ptr(vmtype),
					MaxPods:           utilities.Ptr[int32](110),
					MinCount:          utilities.Ptr[int32](1),
					MaxCount:          utilities.Ptr[int32](100),
					OSType:            utilities.Ptr(armcontainerservice.OSTypeLinux),
					Type:              utilities.Ptr(armcontainerservice.AgentPoolTypeVirtualMachineScaleSets),
					EnableAutoScaling: utilities.Ptr(true),
					Mode:              utilities.Ptr(armcontainerservice.AgentPoolModeSystem),
				},
			},
			ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
				ClientID: utilities.Ptr(os.Getenv("AZURE_CLIENT_ID")),
				Secret:   utilities.Ptr(os.Getenv("AZURE_CLIENT_SECRET")),
			},
			WorkloadAutoScalerProfile: &armcontainerservice.ManagedClusterWorkloadAutoScalerProfile{
				Keda: &armcontainerservice.ManagedClusterWorkloadAutoScalerProfileKeda{
					Enabled: utilities.Ptr(isKeda),
				},
			},
		},
	}

	pollerResp, err := p.client.BeginCreateAKS(name, parameter, nil)
	if err != nil {
		return err
	}
	p.state.CloudInfra.Azure.ManagedClusterName = name
	p.state.CloudInfra.Azure.ManagedNodeSize = vmtype

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Print(p.ctx, "Creating AKS cluster...")

	resp, err := p.client.PollUntilDoneCreateAKS(p.ctx, pollerResp, nil)
	if err != nil {
		return err
	}

	p.state.CloudInfra.Azure.B.IsCompleted = true

	kubeconfig, err := p.client.ListClusterAdminCredentials(name, nil)
	if err != nil {
		return err
	}
	kubeconfigStr := string(kubeconfig.Kubeconfigs[0].Value)

	p.state.ClusterKubeConfig = kubeconfigStr
	p.state.ClusterKubeConfigContext = *resp.Name

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "created AKS", "name", *resp.Name)
	return nil
}
