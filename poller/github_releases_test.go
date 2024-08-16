//go:build !poller_real

package poller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func dummyHttpHandler(org, repo string) ([]string, error) {
	return []string{"v1.0.0", "v1.0.1"}, nil
}

func TestInitFunc(t *testing.T) {
	t.Run("Test without httpCaller", func(t *testing.T) {
		obj := NewGithubReleasePoller(0, dummyHttpHandler)

		assert.NotNil(t, obj.httpCaller)
		assert.Equal(t, DefaultPollerDuration, obj.interval)
		assert.NotNil(t, obj.rwm)
		assert.NotNil(t, obj.cache)
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
	obj := NewGithubReleasePoller(0, dummyHttpHandler)

	t.Run("getReposWithReleases", func(t *testing.T) {
		repos := obj.getSubscribedRepos()
		if len(repos) != 0 {
			t.Errorf("Expected 0, got %d", len(repos))
		}
	})

	t.Run("setRepoWithReleases and try to get it", func(t *testing.T) {
		obj.setReleases("org", "repo", status{
			releases: []string{"v1.0.0", "v1.0.1"},
		})

		repos, ok := obj.getReleases("org", "repo")

		assert.Equal(t, true, ok)
		assert.Equal(t, repos.releases, []string{"v1.0.0", "v1.0.1"})
		assert.Nil(t, repos.err)
	})

	t.Run("setRepoWithReleases and try to get it with errors", func(t *testing.T) {
		obj.setReleases("org", "repo", status{
			err: assert.AnError,
		})

		repos, ok := obj.getReleases("org", "repo")

		assert.Equal(t, true, ok)
		assert.Equal(t, len(repos.releases), 0)
		assert.NotNil(t, repos.err)
	})

	t.Run("Not found repo", func(t *testing.T) {
		_, ok := obj.getReleases("org", "repo111")
		assert.Equal(t, false, ok)
	})
}

func TestGetSubscribedRepos(t *testing.T) {
	obj := NewGithubReleasePoller(0, dummyHttpHandler)

	repos := obj.getSubscribedRepos()
	if len(repos) != 0 {
		t.Errorf("Expected 0, got %d", len(repos))
	}

	obj.cache["org/repo"] = status{}
	repos = obj.getSubscribedRepos()
	if len(repos) != 1 {
		t.Errorf("Expected 1, got %d", len(repos))
	} else {
		if repos[0] != "org/repo" {
			t.Errorf("Expected org/repo, got %s", repos[0])
		}
	}
}

func TestSubscribe(t *testing.T) {
	obj := NewGithubReleasePoller(0, dummyHttpHandler)

	obj.subscribe("org/repo")

	repos, ok := obj.getReleases("org", "repo")
	assert.True(t, ok)
	assert.Equal(t, repos.releases, []string{"v1.0.0", "v1.0.1"})
	assert.Nil(t, repos.err)
}

func TestGetData(t *testing.T) {
	obj := NewGithubReleasePoller(0, dummyHttpHandler)

	repos, err := obj.Get("org", "repo")
	assert.Nil(t, err)
	assert.Equal(t, repos, []string{"v1.0.0", "v1.0.1"})

	<-time.NewTicker(1 * time.Second).C
	repos, err = obj.Get("org", "repo")
	assert.Nil(t, err)
	assert.Equal(t, repos, []string{"v1.0.0", "v1.0.1"})
}

func TestInitSharedGithubReleasePollerInitializesInstance(t *testing.T) {
	InitSharedGithubReleasePoller()
	assert.NotNil(t, instance)
}

func TestInitSharedGithubReleasePollerIsSingleton(t *testing.T) {
	InitSharedGithubReleasePoller()
	firstInstance := instance
	InitSharedGithubReleasePoller()
	assert.Equal(t, firstInstance, instance)
}

func TestInitSharedGithubReleaseFakePollerInitializesInstance(t *testing.T) {
	InitSharedGithubReleaseFakePoller(nil)
	assert.NotNil(t, instance)
}

func TestInitSharedGithubReleaseFakePollerIsSingleton(t *testing.T) {
	InitSharedGithubReleaseFakePoller(nil)
	firstInstance := instance
	InitSharedGithubReleaseFakePoller(nil)
	assert.Equal(t, firstInstance, instance)
}

func TestGetSharedPollerReturnsInstance(t *testing.T) {
	InitSharedGithubReleasePoller()
	poller := GetSharedPoller()
	assert.NotNil(t, poller)
	assert.Equal(t, instance, poller)
}

func TestGetSharedPollerReturnsNilIfNotInitialized(t *testing.T) {
	instance = nil
	poller := GetSharedPoller()
	assert.Nil(t, poller)
}
