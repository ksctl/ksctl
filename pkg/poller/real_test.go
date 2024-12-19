//go:build poller_real

package poller_test

import (
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/poller"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleasePoller()
	m.Run()
}

func TestGithub(t *testing.T) {
	obj := poller.GetSharedPoller()

	r, err := obj.Get("flannel-io", "flannel")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	dump.Println(r)
}
