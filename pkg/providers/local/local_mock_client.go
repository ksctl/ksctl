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
	"sigs.k8s.io/kind/pkg/cluster"
	"time"
)

type LocalClient struct {
	b *Provider
}

func ProvideClient() KindSDK {
	return &LocalClient{}
}

func (l *LocalClient) NewProvider(b *Provider, options ...cluster.ProviderOption) {
	l.b = b
	l.b.l.Debug(l.b.ctx, "NewProvider initialized", "options", options)
}

func (l *LocalClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	path, _ := l.b.createNecessaryConfigs(name)
	l.b.l.Debug(l.b.ctx, "Printing", "path", path)
	l.b.l.Success(l.b.ctx, "Created the cluster",
		"name", name, "config", config,
		"image", image, "wait", wait.String(),
		"configPath", explictPath())
	return nil
}

func (l *LocalClient) Delete(name string, explicitKubeconfigPath string) error {
	l.b.l.Success(l.b.ctx, "Deleted the cluster", "name", name, "kubeconfigPath", explicitKubeconfigPath)
	return nil
}
