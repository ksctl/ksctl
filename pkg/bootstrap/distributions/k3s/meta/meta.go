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
	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions/k3s"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/poller"
)

type K3sMeta struct{}

func NewK3sMeta() *K3sMeta {
	return &K3sMeta{}
}

func (m *K3sMeta) GetBootstrapedDistributionVersions() ([]string, error) {
	validVersion, err := poller.GetSharedPoller().Get("k3s-io", "k3s")
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			err,
		)
	}

	if len(validVersion) == 0 {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Unable to get any releases",
		)
	}

	return validVersion, nil
}

func (m *K3sMeta) GetAvailableCNIPlugins() (addons.ClusterAddons, string, error) {
	v, d := k3s.GetCNIs()

	return v, d, nil
}
