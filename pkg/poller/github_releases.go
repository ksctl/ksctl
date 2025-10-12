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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/github"
)

type Status struct {
	Releases []string `json:"releases"`
	Err      string   `json:"error"`
}

func DecodeStatus(v string) (ret Status, err error) {
	err = json.Unmarshal([]byte(v), &ret)
	return
}

func EncodeStatus(v Status) (ret string, err error) {
	_v, err := json.Marshal(v)
	if err != nil {
		return
	}
	ret = string(_v)
	return
}

type GithubReleasePoller struct {
	interval   time.Duration
	rwm        *sync.RWMutex
	httpCaller func(string, string) ([]string, error)
	c          cache.Cache
}

const (
	delimiter    = "/"
	prefix_cache = "poller:"
)

var DefaultHttpCaller = github.HttpGetAllStableGithubReleases
var DefaultPollerDuration = 30 * time.Minute

func NewGithubReleasePoller(c cache.Cache, t time.Duration, httpCaller func(string, string) ([]string, error)) *GithubReleasePoller {
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
		c:          c,
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

func (gh *GithubReleasePoller) getSubscribedRepos() []string {
	gh.rwm.RLock()
	defer gh.rwm.RUnlock()

	var repos []string
	keys, err := gh.c.KeysWithPrefix(context.TODO(), prefix_cache)
	if err != nil {
		return repos
	}

	for _, k := range keys {
		repos = append(repos, strings.TrimPrefix(k, prefix_cache))
	}

	return repos
}

func (gh *GithubReleasePoller) getReleases(org, repo string) (Status, bool) {
	gh.rwm.RLock()
	defer gh.rwm.RUnlock()

	v, ok := gh.c.Get(context.TODO(), prefix_cache+org+delimiter+repo)
	if !ok {
		return Status{}, false
	}

	c, err := DecodeStatus(v)
	if err != nil {
		return Status{}, false
	}
	return c, ok
}

func (gh *GithubReleasePoller) setReleases(org, repo string, v Status) {
	gh.rwm.Lock()
	defer gh.rwm.Unlock()

	c, err := EncodeStatus(v)
	if err != nil {
		return
	}

	gh.c.SetWithExpire(context.TODO(), prefix_cache+org+delimiter+repo, c, 24*time.Hour)
}

func (gh *GithubReleasePoller) subscribe(repo string) {
	_repo := strings.Split(repo, delimiter)
	releases, err := gh.httpCaller(_repo[0], _repo[1])

	var s Status
	if err != nil {
		s = Status{
			Releases: nil,
			Err:      fmt.Errorf("latest call failed, can contain stale data: %w", err).Error(),
		}
	} else {
		s = Status{
			Releases: releases,
			Err:      "",
		}
	}
	gh.setReleases(_repo[0], _repo[1], s)
}

func (gh *GithubReleasePoller) Get(org, repo string) ([]string, error) {
	v, ok := gh.getReleases(org, repo) // get from cache
	if ok {
		if v.Err == "" {
			return v.Releases, nil
		}
		return v.Releases, errors.New(v.Err)
	}

	gh.subscribe(org + delimiter + repo)
	v, ok = gh.getReleases(org, repo)
	if ok {
		if v.Err == "" {
			return v.Releases, nil
		}
		return v.Releases, errors.New(v.Err)
	} else {
		return nil, fmt.Errorf("failed to store the release in poller")
	}
}
