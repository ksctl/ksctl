// Copyright 2025 Ksctl Authors
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

package meta

import (
	"context"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/dockerhub"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
)

type LocalMeta struct {
	ctx context.Context
	l   logger.Logger

	provider.ProvisionMetadata
}

func NewLocalMeta(ctx context.Context, l logger.Logger) (*LocalMeta, error) {
	return &LocalMeta{
		ctx: ctx,
		l:   l,
	}, nil
}

func (l *LocalMeta) GetAvailableManagedK8sVersions(_ string) ([]string, error) {
	ver, err := dockerhub.HttpGetAllTags("kindest", "node")
	if err != nil {
		return nil, err
	}

	vers := make([]string, 0, len(ver))
	for _, v := range ver {
		vers = append(vers, strings.TrimPrefix(v, "v"))
	}

	l.l.Debug(l.ctx, "Managed K8s versions", "kindVersions", vers)

	return vers, nil
}
