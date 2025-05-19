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

package controller

import (
	"sort"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/config"

	"github.com/ksctl/ksctl/v2/pkg/poller"

	"github.com/ksctl/ksctl/v2/pkg/consts"
)

func (cc *Controller) StartPoller(cacheClient cache.Cache) error {
	if _, ok := config.IsContextPresent(cc.ctx, consts.KsctlTestFlagKey); !ok {
		poller.InitSharedGithubReleasePoller(cacheClient)
	} else {
		poller.InitSharedGithubReleaseFakePoller(cacheClient, func(org, repo string) ([]string, error) {
			vers := []string{"v0.0.1"}

			if org == "etcd-io" && repo == "etcd" {
				vers = append(vers, "v3.5.15")
			}

			if org == "k3s-io" && repo == "k3s" {
				vers = append(vers, "v1.30.3+k3s1")
			}

			if org == "kubernetes" && repo == "kubernetes" {
				vers = append(vers, "v1.31.0")
			}

			sort.Slice(vers, func(i, j int) bool {
				return vers[i] > vers[j]
			})

			return vers, nil
		})
	}

	return nil
}
