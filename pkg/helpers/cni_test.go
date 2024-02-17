package helpers

import (
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func FuzzValidateCNI(f *testing.F) {
	testcases := []string{
		string(consts.CNIAzure),
		string(consts.CNIKind),
		string(consts.CNICilium),
		string(consts.CNIFlannel),
		string(consts.CNIKubenet),
	}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, cni string) {
		ok := ValidCNIPlugin(consts.KsctlValidCNIPlugin(cni))
		t.Logf("cni: %s and ok: %v", cni, ok)
		switch consts.KsctlValidCNIPlugin(cni) {
		case consts.CNIAzure, consts.CNICilium, consts.CNIFlannel, consts.CNIKubenet, consts.CNIKind, "":
			if !ok {
				t.Errorf("Correct cni is invalid")
			} else {
				return
			}
		default:
			if ok {
				t.Errorf("Incorrect cni is valid")
			} else {
				return
			}
		}
	})
}
