package helpers

import (
	"testing"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func FuzzValidateRole(f *testing.F) {
	testcases := []string{
		string(consts.RoleCp),
		string(consts.RoleDs),
		string(consts.RoleLb),
		string(consts.RoleWp),
	}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, role string) {
		ok := ValidateRole(consts.KsctlRole(role))
		t.Logf("storage: %s and ok: %v", role, ok)
		switch consts.KsctlRole(role) {
		case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
			if !ok {
				t.Errorf("Correct role is invalid")
			} else {
				return
			}
		default:
			if ok {
				t.Errorf("incorrect role is valid")
			} else {
				return
			}
		}
	})
}
