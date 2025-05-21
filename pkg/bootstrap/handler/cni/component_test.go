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

package cni

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func TestMain(m *testing.M) {
	cc := cache.NewInMemCache(context.TODO())
	defer cc.Close()

	poller.InitSharedGithubReleaseFakePoller(cc, func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		switch org + " " + repo {
		case "spinkube spin-operator":
			vers = append(vers, "v0.2.0")
		case "spinkube containerd-shim-spin":
			vers = append(vers, "v0.15.1")
		case "cert-manager cert-manager":
			vers = append(vers, "v1.15.3")
		case "cilium cilium":
			vers = append(vers, "v1.16.1")
		case "flannel-io flannel":
			vers = append(vers, "v0.25.5")
		case "argoproj argo-rollouts":
			vers = append(vers, "v1.7.2")
		case "istio istio":
			vers = append(vers, "v1.22.4")
		}

		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		if org == "istio" {
			for i := range vers {
				if strings.HasPrefix(vers[i], "v") {
					vers[i] = strings.TrimPrefix(vers[i], "v")
				}
			}
		}

		return vers, nil
	})
	m.Run()
}

func TestGetVersionIfItsNotNilAndLatest(t *testing.T) {
	tests := []struct {
		name        string
		version     *string
		defaultVer  string
		expectedVer string
	}{
		{
			name:        "version is nil",
			version:     nil,
			defaultVer:  "v0.0.1",
			expectedVer: "v0.0.1",
		},
		{
			name:        "version is latest",
			version:     utilities.Ptr("latest"),
			defaultVer:  "v0.0.1",
			expectedVer: "v0.0.1",
		},
		{
			name:        "version is not latest",
			version:     utilities.Ptr("v1.0.0"),
			defaultVer:  "v0.0.1",
			expectedVer: "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualVer := getVersionIfItsNotNilAndLatest(tt.version, tt.defaultVer)
			if actualVer != tt.expectedVer {
				t.Errorf("expected version %s, got %s", tt.expectedVer, actualVer)
			}
		})
	}
}
