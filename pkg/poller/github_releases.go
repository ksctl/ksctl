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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

type status struct {
	releases []string
	err      error
}

type GithubReleasePoller struct {
	interval   time.Duration
	rwm        *sync.RWMutex
	httpCaller func(string, string) ([]string, error)
	cache      map[string]status
}

const (
	delimiter = "/"
)

var DefaultHttpCaller = utilities.HttpGetAllStableGithubReleases
var DefaultPollerDuration = 30 * time.Minute

func NewGithubReleasePoller(t time.Duration, httpCaller func(string, string) ([]string, error)) *GithubReleasePoller {
	if httpCaller == nil {
		httpCaller = DefaultHttpCaller
	}
	if t == 0 {
		t = DefaultPollerDuration
	}

	gh := &GithubReleasePoller{
		interval:   t,
		rwm:        new(sync.RWMutex),
		httpCaller: httpCaller,
		cache:      make(map[string]status),
	}

	go gh.handleRefresh()

	return gh
}

func (gh *GithubReleasePoller) handleRefresh() {
	t := time.NewTicker(gh.interval)
	for range t.C {
		repos := gh.getSubscribedRepos()

		for _, repo := range repos {
			gh.subscribe(repo)
		}
	}
}

func (p *GithubReleasePoller) getSubscribedRepos() []string {
	p.rwm.RLock()
	defer p.rwm.RUnlock()

	var repos []string
	for k, _ := range p.cache {
		repos = append(repos, k)
	}

	return repos
}

func (p *GithubReleasePoller) getReleases(org, repo string) (status, bool) {
	p.rwm.RLock()
	defer p.rwm.RUnlock()

	v, ok := p.cache[org+delimiter+repo]
	return v, ok
}

func (gh *GithubReleasePoller) setReleases(org, repo string, v status) {
	gh.rwm.Lock()
	defer gh.rwm.Unlock()

	gh.cache[org+delimiter+repo] = v
}

func (gh *GithubReleasePoller) subscribe(repo string) {
	_repo := strings.Split(repo, delimiter)
	releases, err := gh.httpCaller(_repo[0], _repo[1])

	var s status
	if err != nil {
		s = status{
			releases: nil,
			err:      fmt.Errorf("Latest call failed, can contain stale data: %w", err),
		}
	} else {
		s = status{
			releases: releases,
			err:      nil,
		}
	}
	gh.setReleases(_repo[0], _repo[1], s)
}

func (p *GithubReleasePoller) Get(org, repo string) ([]string, error) {
	v, ok := p.getReleases(org, repo) // get from cache
	if ok {
		return v.releases, v.err
	}

	p.subscribe(org + delimiter + repo)
	v, ok = p.getReleases(org, repo)
	if ok {
		return v.releases, v.err
	} else {
		return nil, fmt.Errorf("Failed to store the release in poller")
	}
}
