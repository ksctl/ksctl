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

package local

import (
	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
	"time"
)

type metadata struct {
	resName           string
	version           string
	tempDirKubeconfig string
	cni               string
}

type LocalProvider struct {
	clusterName string
	region      string
	vmType      string
	metadata

	client LocalGo
}

type LocalGo interface {
	NewProvider(log types.LoggerFactory, storage types.StorageFactory, options ...cluster.ProviderOption)
	Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error
	Delete(name string, explicitKubeconfigPath string) error
}
