package utils

import (
	"testing"

	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func FuzzValidateDistro(f *testing.F) {
	testcases := []string{
		string(consts.K8sKubeadm),
		string(consts.K8sK3s),
		"",
	}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, distro string) {
		ok := ValidateDistro(consts.KsctlKubernetes(distro))
		t.Logf("distro: %s and ok: %v", distro, ok)
		switch consts.KsctlKubernetes(distro) {
		case consts.K8sK3s, consts.K8sKubeadm, "":
			if !ok {
				t.Errorf("Correct distro is invalid")
			} else {
				return
			}
		default:
			if ok {
				t.Errorf("incorrect distro is valid")
			} else {
				return
			}
		}
	})
}
