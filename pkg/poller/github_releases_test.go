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

//go:build !poller_real

package poller

import (
	"testing"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func dummyHttpHandler(org, repo string) ([]string, error) {
	return []string{"v1.0.0", "v1.0.1"}, nil
}

func TestInitFunc(t *testing.T) {
	t.Run("Test without httpCaller", func(t *testing.T) {
		cc := cache.NewInMemCache(t.Context())
		defer cc.Close()

		obj := NewGithubReleasePoller(cc, 0, dummyHttpHandler)

		assert.NotNil(t, obj.httpCaller)
		assert.Equal(t, DefaultPollerDuration, obj.interval)
		assert.NotNil(t, obj.rwm)
	})
}

func TestGlobalvars(t *testing.T) {
	t.Run("Test global vars", func(t *testing.T) {
		assert.NotNil(t, DefaultHttpCaller)
		assert.Equal(t, 30*time.Minute, DefaultPollerDuration)
		assert.Equal(t, "/", delimiter)
	})
}

func TestRepoSetterAndGetter(t *testing.T) {
	cc := cache.NewInMemCache(t.Context())
	defer cc.Close()
	obj := NewGithubReleasePoller(cc, 0, dummyHttpHandler)

	t.Run("getReposWithReleases", func(t *testing.T) {
		repos := obj.getSubscribedRepos()
		if len(repos) != 0 {
			t.Errorf("Expected 0, got %d", len(repos))
		}
	})

	t.Run("setRepoWithReleases and try to get it", func(t *testing.T) {
		obj.setReleases("org", "repo", Status{
			Releases: []string{"v1.0.0", "v1.0.1"},
		})

		repos, ok := obj.getReleases("org", "repo")

		assert.Equal(t, true, ok)
		assert.Equal(t, repos.Releases, []string{"v1.0.0", "v1.0.1"})
		assert.Equal(t, repos.Err, "")
	})

	t.Run("setRepoWithReleases and try to get it with errors", func(t *testing.T) {
		obj.setReleases("org", "repo", Status{
			Err: assert.AnError.Error(),
		})

		repos, ok := obj.getReleases("org", "repo")

		assert.Equal(t, true, ok)
		assert.Equal(t, len(repos.Releases), 0)
		assert.NotEqual(t, repos.Err, "")
	})

	t.Run("Not found repo", func(t *testing.T) {
		_, ok := obj.getReleases("org", "repo111")
		assert.Equal(t, false, ok)
	})
}

func TestGetSubscribedRepos(t *testing.T) {
	cc := cache.NewInMemCache(t.Context())
	defer cc.Close()
	obj := NewGithubReleasePoller(cc, 0, dummyHttpHandler)

	repos := obj.getSubscribedRepos()
	if len(repos) != 0 {
		t.Errorf("Expected 0, got %d", len(repos))
	}

	obj.c.Set(prefix_cache+"org/repo", `{"releases": [], "err": ""}`)
	repos = obj.getSubscribedRepos()
	if len(repos) != 1 {
		t.Errorf("Expected 1, got %d", len(repos))
	} else {
		if repos[0] != prefix_cache+"org/repo" {
			t.Errorf("Expected org/repo, got %s", repos[0])
		}
	}
}

func TestSubscribe(t *testing.T) {
	cc := cache.NewInMemCache(t.Context())
	defer cc.Close()

	obj := NewGithubReleasePoller(cc, 0, dummyHttpHandler)

	obj.subscribe("org/repo")

	repos, ok := obj.getReleases("org", "repo")
	assert.True(t, ok)
	assert.Equal(t, repos.Releases, []string{"v1.0.0", "v1.0.1"})
	assert.Equal(t, repos.Err, "")
}

func TestGetData(t *testing.T) {
	cc := cache.NewInMemCache(t.Context())
	defer cc.Close()
	obj := NewGithubReleasePoller(cc, 0, dummyHttpHandler)

	repos, err := obj.Get("org", "repo")
	assert.Nil(t, err)
	assert.Equal(t, repos, []string{"v1.0.0", "v1.0.1"})

	<-time.NewTicker(1 * time.Second).C
	repos, err = obj.Get("org", "repo")
	assert.Nil(t, err)
	assert.Equal(t, repos, []string{"v1.0.0", "v1.0.1"})
}
