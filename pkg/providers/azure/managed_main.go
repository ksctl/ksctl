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
	"os"

	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *AzureProvider) DelManagedCluster(storage types.StorageFactory) error {
	if len(mainStateDocument.CloudInfra.Azure.ManagedClusterName) == 0 {
		log.Print(azureCtx, "skipped already deleted AKS cluster")
		return nil
	}

	pollerResp, err := obj.client.BeginDeleteAKS(mainStateDocument.CloudInfra.Azure.ManagedClusterName, nil)
	if err != nil {
		return err
	}
	log.Print(azureCtx, "Deleting AKS cluster...", "name", mainStateDocument.CloudInfra.Azure.ManagedClusterName)

	_, err = obj.client.PollUntilDoneDelAKS(azureCtx, pollerResp, nil)
	if err != nil {
		return err
	}

	log.Success(azureCtx, "Deleted the AKS cluster", "name", mainStateDocument.CloudInfra.Azure.ManagedClusterName)

	mainStateDocument.CloudInfra.Azure.ManagedClusterName = ""
	mainStateDocument.CloudInfra.Azure.ManagedNodeSize = ""
	return storage.Write(mainStateDocument)
}

func (obj *AzureProvider) NewManagedCluster(storage types.StorageFactory, noOfNodes int) error {
	name := <-obj.chResName
	vmtype := <-obj.chVMType

	log.Debug(azureCtx, "Printing", "name", name, "vmtype", vmtype)

	if len(mainStateDocument.CloudInfra.Azure.ManagedClusterName) != 0 {
		log.Print(azureCtx, "skipped already created AKS cluster", "name", mainStateDocument.CloudInfra.Azure.ManagedClusterName)
		return nil
	}

	mainStateDocument.CloudInfra.Azure.NoManagedNodes = noOfNodes
	mainStateDocument.CloudInfra.Azure.B.KubernetesVer = obj.metadata.k8sVersion
	mainStateDocument.BootstrapProvider = "managed"

	parameter := armcontainerservice.ManagedCluster{
		Location: utilities.Ptr(mainStateDocument.Region),
		Properties: &armcontainerservice.ManagedClusterProperties{
			DNSPrefix:         utilities.Ptr("aksgosdk"),
			KubernetesVersion: utilities.Ptr(mainStateDocument.CloudInfra.Azure.B.KubernetesVer),
			NetworkProfile: &armcontainerservice.NetworkProfile{
				NetworkPlugin: utilities.Ptr[armcontainerservice.NetworkPlugin](armcontainerservice.NetworkPlugin(obj.metadata.cni)),
			},
			AutoUpgradeProfile: &armcontainerservice.ManagedClusterAutoUpgradeProfile{
				NodeOSUpgradeChannel: utilities.Ptr[armcontainerservice.NodeOSUpgradeChannel](armcontainerservice.NodeOSUpgradeChannelNodeImage),
				UpgradeChannel:       utilities.Ptr[armcontainerservice.UpgradeChannel](armcontainerservice.UpgradeChannelPatch),
			},
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					Name:              utilities.Ptr("askagent"),
					Count:             utilities.Ptr[int32](int32(noOfNodes)),
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
		},
	}

	log.Debug(azureCtx, "Printing", "AKSConfig", parameter)

	pollerResp, err := obj.client.BeginCreateAKS(name, parameter, nil)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Azure.ManagedClusterName = name
	mainStateDocument.CloudInfra.Azure.ManagedNodeSize = vmtype

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Print(azureCtx, "Creating AKS cluster...")

	resp, err := obj.client.PollUntilDoneCreateAKS(azureCtx, pollerResp, nil)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Azure.B.IsCompleted = true

	kubeconfig, err := obj.client.ListClusterAdminCredentials(name, nil)
	if err != nil {
		return err
	}
	kubeconfigStr := string(kubeconfig.Kubeconfigs[0].Value)

	mainStateDocument.ClusterKubeConfig = kubeconfigStr
	mainStateDocument.ClusterKubeConfigContext = *resp.Name

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	log.Success(azureCtx, "created AKS", "name", *resp.Name)
	return nil
}
