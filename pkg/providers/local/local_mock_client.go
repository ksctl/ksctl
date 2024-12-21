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

//go:build testing_local

package local

import (
	"time"

	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
)

type LocalClient struct {
	log   types.LoggerFactory
	store types.StorageFactory
}

func ProvideClient() LocalGo {
	return &LocalClient{}
}

func (l *LocalClient) NewProvider(log types.LoggerFactory, storage types.StorageFactory, options ...cluster.ProviderOption) {
	log.Debug(localCtx, "NewProvider initialized", "options", options)
	l.store = storage
	l.log = log
}

func (l *LocalClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	path, _ := createNecessaryConfigs(name)
	l.log.Debug(localCtx, "Printing", "path", path)
	l.log.Success(localCtx, "Created the cluster",
		"name", name, "config", config,
		"image", image, "wait", wait.String(),
		"configPath", explictPath())
	return nil
}

func (l *LocalClient) Delete(name string, explicitKubeconfigPath string) error {
	l.log.Success(localCtx, "Deleted the cluster", "name", name, "kubeconfigPath", explicitKubeconfigPath)
	return nil
}
