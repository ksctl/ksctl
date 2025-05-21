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

package poller

import (
	"sync"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/cache"
)

var (
	instance *GithubReleasePoller
	once     sync.Once
)

func InitSharedGithubReleasePoller(c cache.Cache) {
	instance = NewGithubReleasePoller(c, DefaultPollerDuration, DefaultHttpCaller)
}

// InitSharedGithubReleaseFakePoller initializes a shared poller with a fake implementation
//
// fakeValidVersions: a function that returns a list of valid versions for a given org and repo, it should be sorted like latest version first to oldest
func InitSharedGithubReleaseFakePoller(c cache.Cache, fakeValidVersions func(org, repo string) ([]string, error)) {
	if fakeValidVersions == nil {
		fakeValidVersions = func(org, repo string) ([]string, error) {
			return []string{"v1.0.0", "v1.0.1"}, nil
		}
	}
	instance = NewGithubReleasePoller(c, 5*time.Second, fakeValidVersions)
}

func GetSharedPoller() Poller {
	return instance
}
