// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/ksctl/ksctl/pkg/addons"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/logger"
	"gotest.tools/v3/assert"
)

func FuzzValidateCloud(f *testing.F) {
	testcases := []string{
		string(consts.CloudAll),
		string(consts.CloudLocal),
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
		case consts.CloudAll, consts.CloudAws, consts.CloudAzure, consts.CloudLocal:
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

func FuzzName(f *testing.F) {
	testcases := []string{"avcd", "nice-23rde-fe423"}

	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, name string) {
		outErr := IsValidName(context.TODO(), logger.NewStructuredLogger(-1, os.Stdout), name)
		t.Logf("name: %s and err: %v", name, outErr)
		matched, err := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, name)

		if outErr == nil && len(name) > 50 {
			t.Errorf("incorrect error for long length string")
		}
		if outErr != nil && (!matched || err != nil) {
			return
		}
		if outErr == nil && (matched && err == nil) {
			return
		}
	})
}

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

var (
	dummyCtx = context.WithValue(context.TODO(), consts.KsctlTestFlagKey, "true")
	log      = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestIsValidClusterName(t *testing.T) {
	assert.Check(t, nil == IsValidName(dummyCtx, log, "demo"), "Returns false for valid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "Dem-o234")
			return err != nil && ksctlErrors.IsInvalidResourceName(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(t, nil == IsValidName(dummyCtx, log, "d-234"), "Returns false for valid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "234")
			return err != nil && ksctlErrors.IsInvalidResourceName(err)
		}(),
		"Returns true for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "-2342")
			return err != nil && ksctlErrors.IsInvalidResourceName(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "demo-")
			return err != nil && ksctlErrors.IsInvalidResourceName(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "dscdscsd-#$#$#")
			return err != nil && ksctlErrors.IsInvalidResourceName(err)
		}(),
		"Returns True for invalid cluster name")
	assert.Check(
		t,
		func() bool {
			err := IsValidName(dummyCtx, log, "dds@#$#$#ds@#$#$#ds@#$#$#ds@#$#$#ds@#$#$#s@#$#$wefe#")
			return err != nil && ksctlErrors.IsInvalidResourceName(err)
		}(),
		"Returns True for invalid cluster name")
}

func TestIsValidVersion(t *testing.T) {
	testCases := map[string]bool{
		"1.1.1":            true,
		"latest":           true,
		"v1":               true,
		"v1.1":             true,
		"v1.1.1":           true,
		"1.1":              true,
		"1":                true,
		"v":                false,
		"stable":           true,
		"enhancement-2342": true,
		"enhancement":      true,
		"feature-2342":     true,
		"feature":          true,
		"feat":             true,
		"feat234":          true,
		"fix234":           true,
		"f14cd9094b2160c40ef8734e90141df81c22999e": true,
	}

	for ver, expected := range testCases {
		err := IsValidKsctlComponentVersion(dummyCtx, log, ver)
		got := err == nil
		assert.Equal(t, got, expected, fmt.Sprintf("Ver: %s, got: %v, expected: %v", ver, got, expected))
	}
}

func TestIsValidKsctlClusterAddons(t *testing.T) {

	testCases := []struct {
		ca            addons.ClusterAddons
		expectedError bool
	}{
		{
			ca:            addons.ClusterAddons{},
			expectedError: false,
		},
		{
			ca: addons.ClusterAddons{
				{Label: "dcdcd", Name: "cilium", IsCNI: true},
			},
			expectedError: true,
		},
		{
			ca: addons.ClusterAddons{
				{Label: "ksctl", Name: "cilium", IsCNI: true},
			},
			expectedError: true,
		},
		{
			ca: addons.ClusterAddons{
				{Label: "ksctl", Name: "cilium", IsCNI: true},
				{Label: "aks", Name: "none", IsCNI: true},
			},
			expectedError: false,
		},
		{
			ca: addons.ClusterAddons{
				{Label: "aks", Name: "none", IsCNI: true},
				{Label: "aks", Name: "none", IsCNI: true},
			},
			expectedError: true,
		},
		{ // when CNI is specified as none it doesn't matter if the ksctl handles it with its own cni
			ca: addons.ClusterAddons{
				{Label: "eks", Name: "none", IsCNI: true},
			},
			expectedError: false,
		},
		{ // No Duplicates in the addons
			ca: addons.ClusterAddons{
				{Label: "kind", Name: "random123"},
				{Label: "kind", Name: "random123"},
			},
			expectedError: true,
		},
		{
			ca: addons.ClusterAddons{
				{Label: "aks", Name: "random123"},
				{Label: "eks", Name: "random123"},
				{Label: "gke", Name: "random123"},
				{Label: "kind", Name: "random123"},
				{Label: "k3s", Name: "random123"},
				{Label: "kubeadm", Name: "random123"},
				{Label: "ksctl", Name: "random123"},
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		err := IsValidKsctlClusterAddons(dummyCtx, log, tc.ca)
		assert.Equal(t, err == nil, !tc.expectedError)
	}
}
