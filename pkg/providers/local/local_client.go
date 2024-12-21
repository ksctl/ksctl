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

//go:build !testing_local

package local

import (
	"github.com/ksctl/ksctl/pkg/types"
	"sigs.k8s.io/kind/pkg/cluster"
	"time"
)

type LocalClient struct {
	provider *cluster.Provider
	log      types.LoggerFactory
}

func ProvideClient() LocalGo {
	return &LocalClient{}
}

func (l *LocalClient) NewProvider(log types.LoggerFactory, _ types.StorageFactory, options ...cluster.ProviderOption) {
	logger := &customLogger{Logger: log}
	options = append(options, cluster.ProviderWithLogger(logger))
	l.log = log
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
