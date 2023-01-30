package utils

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"gotest.tools/assert"
)

func TestGetKubeConfigWithUserName(t *testing.T) {
	if runtime.GOOS == "windows" {

		assert.Equal(t, os.Getenv("USERPROFILE"), GetUserName(), "Unable to fetch correct username")

		assert.Equal(t, fmt.Sprintf("%s\\.ksctl\\config\\abcd\\123w", GetUserName()),
			GetKubeconfig("abcd", "123w"), "Kube config failed, as expected is not equal to actual")

	} else {

		assert.Equal(t, os.Getenv("HOME"), GetUserName(), "Unable to fetch correct username")

		assert.Equal(t, fmt.Sprintf("%s/.ksctl/config/abcd/123w", GetUserName()),
			GetKubeconfig("abcd", "123w"), "Kube config failed, as expected is not equal to actual")
	}
}

func TestValidRegionsInCIVO(t *testing.T) {
	testcase := map[string]bool{
		"LON1":  true,
		"FRA1":  true,
		"NYC1":  true,
		"nYv":   false,
		"LON1 ": false,
	}
	for actualRegion, expectedRes := range testcase {
		if expectedRes != IsValidRegionCIVO(actualRegion) {
			t.Fatalf("Region validation of CIVO failed!")
		}
	}
}

func TestIsValidClusterName(T *testing.T) {
	assert.Equal(T, true, IsValidName("demo"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("Dem-o234"), "Returns True for invalid cluster name")
	assert.Equal(T, true, IsValidName("d-234"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("234"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("-2342"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("dscdscsd-#$#$#"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("ds@#$#$#"), "Returns True for invalid cluster name")
}
