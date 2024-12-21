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

package storage

import "github.com/ksctl/ksctl/pkg/consts"

type Storage interface {
	// Kill to achieve graceful termination we can store a boolean flag in the
	// storagedriver that whether there was any write operation if yes and a reference
	//always present in the storagedriver we can make the driver write the struct once termination is triggered
	Kill() error

	Connect() error

	Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	Write(*storage.StorageDocument) error

	WriteCredentials(consts.KsctlCloud, *storage.CredentialsDocument) error

	Read() (*storage.StorageDocument, error)

	ReadCredentials(consts.KsctlCloud) (*storage.CredentialsDocument, error)

	DeleteCluster() error

	AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	GetOneOrMoreClusters(filters map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*storage.StorageDocument, error)

	// Export is not goroutine safe, but the child process it calls is!
	Export(filters map[consts.KsctlSearchFilter]string) (*StorageStateExportImport, error)

	// Import is not goroutine safe, but the child process it calls is!
	Import(*StorageStateExportImport) error
}

type StorageStateExportImport struct {
	Clusters    []*storage.StorageDocument
	Credentials []*storage.CredentialsDocument
}
