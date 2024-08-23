package components

import (
	"sort"
	"strings"
	"testing"

	"github.com/ksctl/ksctl/poller"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
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
