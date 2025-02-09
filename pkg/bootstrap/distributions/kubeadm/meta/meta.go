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
	"fmt"
	"strings"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

type KubeadmMeta struct{}

func NewKubeadmMeta() *KubeadmMeta {
	return &KubeadmMeta{}
}

func (m *KubeadmMeta) GetBootstrapedDistributionVersions() ([]string, error) {
	validVersion, err := poller.GetSharedPoller().Get("kubernetes", "kubernetes")
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

	for i := range validVersion {
		_v := strings.Split(validVersion[i], ".")
		if len(_v) == 3 {
			validVersion[i] = fmt.Sprintf("%s.%s", _v[0], _v[1])
		}
	}

	return utilities.DeduplicateStringsAlreadySorted(validVersion), nil
}
