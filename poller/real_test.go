//go:build poller_real

package poller_test

import (
	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/poller"
	"testing"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleasePoller()
	m.Run()
}

func TestGithub(t *testing.T) {
	obj := poller.GetSharedPoller()

	r, err := obj.Get("ksctl", "cli")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	dump.Println(r)
}
