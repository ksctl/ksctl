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

package storage

import (
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

type Storage interface {
	Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	Write(*statefile.StorageDocument) error

	Read() (*statefile.StorageDocument, error)

	DeleteCluster() error

	AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error

	GetOneOrMoreClusters(filters map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*statefile.StorageDocument, error)
}
