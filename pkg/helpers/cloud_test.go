package helpers

import (
	"testing"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
)

func FuzzValidateCloud(f *testing.F) {
	testcases := []string{
		string(consts.CloudAll),
		string(consts.CloudLocal),
		string(consts.CloudCivo),
		string(consts.CloudAws),
		string(consts.CloudAzure),
	}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, cloud string) {
		ok := ValidateCloud(consts.KsctlCloud(cloud))
		t.Logf("cloud: %s and ok: %v", cloud, ok)
		switch consts.KsctlCloud(cloud) {
		case consts.CloudAll, consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
			if !ok {
				t.Errorf("Correct cloud provider is invalid")
			} else {
				return
			}
		default:
			if ok {
				t.Errorf("incorrect cloud provider is valid")
			} else {
				return
			}
		}
	})
}
