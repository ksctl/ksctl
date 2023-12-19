package helpers

import (
	"testing"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
)

func FuzzValidateStorage(f *testing.F) {
	testcases := []string{
		string(consts.StoreLocal),
		string(consts.StoreExtMongo),
	}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, store string) {
		ok := ValidateStorage(consts.KsctlStore(store))
		t.Logf("storage: %s and ok: %v", store, ok)
		switch consts.KsctlStore(store) {
		case consts.StoreLocal, consts.StoreExtMongo:
			if !ok {
				t.Errorf("Correct storage is invalid")
			} else {
				return
			}
		default:
			if ok {
				t.Errorf("incorrect storage is valid")
			} else {
				return
			}
		}
	})
}
