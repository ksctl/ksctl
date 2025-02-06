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
	"github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions"
	k3sMeta "github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions/k3s/meta"
	kubeadmMeta "github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions/kubeadm/meta"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/poller"
)

type BootstrapMetadata struct {
	// l   logger.Logger
	// ctx context.Context

	D distributions.DistributionMetadata
}

func NewBootstrapMetadata(clusterType consts.KsctlKubernetes) *BootstrapMetadata {
	b := &BootstrapMetadata{}
	switch clusterType {
	case consts.K8sK3s:
		b.D = k3sMeta.NewK3sMeta()
	case consts.K8sKubeadm:
		b.D = kubeadmMeta.NewKubeadmMeta()
	}

	return b
}

func (b *BootstrapMetadata) GetAvailableEtcdVersions() ([]string, error) {
	vers, err := poller.GetSharedPoller().Get("etcd-io", "etcd")
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			err,
		)
	}

	if len(vers) == 0 {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Unable to get any releases",
		)
	}

	return vers, nil
}
