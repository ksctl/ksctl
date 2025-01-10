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

package handler

func (kc *Controller) CreateManagedCluster() (bool, error) {

	if !kc.b.IsLocalProvider(kc.p) {
		if err := kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(); err != nil {
			return false, err
		}
	}

	managedClient := kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-ksctl-managed")

	managedClient = managedClient.VMType(kc.p.Metadata.ManagedNodeType)

	externalCNI := managedClient.ManagedAddons(kc.p.Metadata.Addons)

	managedClient = managedClient.ManagedK8sVersion(kc.p.Metadata.K8sVersion)

	if managedClient == nil {
		return externalCNI, kc.l.NewError(kc.ctx, "invalid k8s version")
	}

	if err := managedClient.NewManagedCluster(kc.p.Metadata.NoMP); err != nil {
		return externalCNI, err
	}
	return externalCNI, nil
}

func (kc *Controller) DeleteManagedCluster() error {

	if err := kc.p.Cloud.DelManagedCluster(); err != nil {
		return err
	}

	if !kc.b.IsLocalProvider(kc.p) {
		if err := kc.p.Cloud.DelNetwork(); err != nil {
			return err
		}
	}
	return nil
}
