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

//go:build !testing_local

package local

import (
	"time"

	"github.com/ksctl/ksctl/pkg/logger"
	"sigs.k8s.io/kind/pkg/cluster"
)

type LocalClient struct {
	provider  *cluster.Provider
	customLog logger.Logger
	b         *Provider
}

func ProvideClient() KindSDK {
	return &LocalClient{}
}

func (l *LocalClient) NewProvider(b *Provider, options ...cluster.ProviderOption) {
	options = append(options, cluster.ProviderWithLogger(&customLogger{Logger: b.l}))
	l.b = b
	l.provider = cluster.NewProvider(options...)
}

func (l *LocalClient) Create(name string, config cluster.CreateOption, image string, wait time.Duration, explictPath func() string) error {
	return l.provider.Create(
		name,
		config,
		cluster.CreateWithNodeImage(image),
		cluster.CreateWithWaitForReady(wait),
		cluster.CreateWithKubeconfigPath(explictPath()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
}

func (l *LocalClient) Delete(name string, explicitKubeconfigPath string) error {
	return l.provider.Delete(name, explicitKubeconfigPath)
}
