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

package providers

import "github.com/ksctl/ksctl/pkg/consts"

type Cloud interface {
	NewVM(int) error

	DelVM(int) error

	NewFirewall() error

	DelFirewall() error

	NewNetwork() error

	DelNetwork() error

	Credential() error

	InitState(consts.KsctlOperation) error

	CreateUploadSSHKeyPair() error

	DelSSHKeyPair() error

	GetStateForHACluster() (cloud.CloudResourceState, error)

	NewManagedCluster(int) error

	DelManagedCluster() error

	GetRAWClusterInfos() ([]cloud.AllClusterData, error)

	Name(string) Cloud

	Role(consts.KsctlRole) Cloud

	VMType(string) Cloud

	Visibility(bool) Cloud

	Application([]string) bool

	CNI(string) (willBeInstalled bool)

	ManagedK8sVersion(string) Cloud

	NoOfWorkerPlane(int, bool) (int, error)

	NoOfControlPlane(int, bool) (int, error)

	NoOfDataStore(int, bool) (int, error)

	GetHostNameAllWorkerNode() []string

	IsPresent() error

	GetKubeconfig() (*string, error)
}
