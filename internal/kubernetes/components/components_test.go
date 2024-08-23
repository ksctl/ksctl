package components

import (
	"sort"
	"testing"

	"github.com/ksctl/ksctl/poller"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		if org == "spinkube" {
			if repo == "spin-operator" {
				vers = append(vers, "v0.2.0")
			} else if repo == "containerd-shim-spin" {
				vers = append(vers, "v0.15.1")
			}
		}
		if org == "cert-manager" && repo == "cert-manager" {
			vers = append(vers, "v1.15.3")
		}
		if org == "cilium" && repo == "cilium" {
			vers = append(vers, "v1.16.1")
		}
		if org == "flannel-io" && repo == "flannel" {
			vers = append(vers, "v0.25.5")
		}

		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		return vers, nil
	})
	m.Run()
}
