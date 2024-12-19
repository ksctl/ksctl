package poller

import (
	"sync"
	"time"
)

var (
	instance *GithubReleasePoller
	once     sync.Once
)

func InitSharedGithubReleasePoller() {
	once.Do(func() {
		instance = NewGithubReleasePoller(DefaultPollerDuration, DefaultHttpCaller)
	})
}

// InitSharedGithubReleaseFakePoller initializes a shared poller with a fake implementation
//
// fakeValidVersions: a function that returns a list of valid versions for a given org and repo, it should be sorted like latest version first to oldest
func InitSharedGithubReleaseFakePoller(fakeValidVersions func(org, repo string) ([]string, error)) {
	once.Do(func() {
		if fakeValidVersions == nil {
			fakeValidVersions = func(org, repo string) ([]string, error) {
				return []string{"v1.0.0", "v1.0.1"}, nil
			}
		}
		instance = NewGithubReleasePoller(5*time.Second, fakeValidVersions)
	})
}

func GetSharedPoller() Poller {
	return instance
}
