package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestGetUsername(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, os.Getenv("USERPROFILE"), GetUserName(), "Unable to fetch correct username")
	} else {
		assert.Equal(t, os.Getenv("HOME"), GetUserName(), "Unable to fetch correct username")
	}
}

func TestGetCredentials(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, fmt.Sprintf("%s\\.ksctl\\cred\\XX-yy-Zz.json", GetUserName()), getCredentials("XX-yy-Zz"),
			"unable to fetch the cred path")
	} else {
		assert.Equal(t, fmt.Sprintf("%s/.ksctl/cred/XX-yy-Zz.json", GetUserName()), getCredentials("XX-yy-Zz"),
			"unable to fetch the cred path")
	}
}

func TestGetPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		dummy := fmt.Sprintf("%s\\.ksctl\\config\\xx\\Yy zz", GetUserName())
		assert.Equal(t, dummy, getPaths("xx", "Yy zz"))
		if strings.Compare(dummy, getPaths("xx", "Ydcsd")) == 0 {
			t.Fatalf("GetPath testing failed")
		}
	} else {
		dummy := fmt.Sprintf("%s/.ksctl/config/xx/Yy zz", GetUserName())
		assert.Equal(t, dummy, getPaths("xx", "Yy zz"))
		if strings.Compare(dummy, getPaths("xx", "Ydcsd")) == 0 {
			t.Fatalf("GetPath testing failed")
		}
	}
}

func TestGetClusterPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		testProviders := map[string]string{
			"civo":  fmt.Sprintf("%s\\.ksctl\\config\\civo\\Yy zz", GetUserName()),
			"local": fmt.Sprintf("%s\\.ksctl\\config\\local\\Yy zz", GetUserName()),
			"xx":    "",
			"Xyz":   "",
			"azure": fmt.Sprintf("%s\\.ksctl\\config\\azure\\Yy zz", GetUserName()),
		}
		for provider, expectedPath := range testProviders {
			assert.Equal(t, expectedPath, getKubeconfig(provider, "Yy zz")) // must return empty string as its invalid provider
		}
	} else {
		testProviders := map[string]string{
			"civo":  fmt.Sprintf("%s/.ksctl/config/civo/Yy zz", GetUserName()),
			"local": fmt.Sprintf("%s/.ksctl/config/local/Yy zz", GetUserName()),
			"xx":    "",
			"Xyz":   "",
			"azure": fmt.Sprintf("%s/.ksctl/config/azure/Yy zz", GetUserName()),
		}
		for provider, expectedPath := range testProviders {
			assert.Equal(t, expectedPath, getKubeconfig(provider, "Yy zz")) // must return empty string as its invalid provider
		}
	}
}

func TestGetOtherPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, fmt.Sprintf("%s\\.ksctl\\config\\abcd\\Yy zz", GetUserName()), getPaths("abcd", "Yy zz"))
	} else {
		assert.Equal(t, fmt.Sprintf("%s/.ksctl/config/abcd/Yy zz", GetUserName()), getPaths("abcd", "Yy zz"))
	}
}

func TestCreateSSHKeyPair(t *testing.T) {
	// driver

	t.Deadline()
	provider := "Provider"
	clusterName := "cluster"
	clusterRegion := "RegionXYz" // with the region as well

	t.Cleanup(func() {
		_ = os.RemoveAll(GetPath(OTHER_PATH, provider))
	})

	path := GetPath(OTHER_PATH, provider, "ha", clusterName+" "+clusterRegion)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Unable to create dummy folder")
	}
	if _, err := CreateSSHKeyPair(provider, clusterName, clusterRegion); err != nil {
		t.Fatalf("Unable to create SSH keypair")
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
	assert.Equal(T, true, IsValidName("demo"), "Returns false for valid cluster name")
	assert.Equal(T, false, IsValidName("Dem-o234"), "Returns True for invalid cluster name")
	assert.Equal(T, true, IsValidName("d-234"), "Returns false for valid cluster name")
	assert.Equal(T, false, IsValidName("234"), "Returns true for invalid cluster name")
	assert.Equal(T, false, IsValidName("-2342"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("demo-"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("dscdscsd-#$#$#"), "Returns True for invalid cluster name")
	assert.Equal(T, false, IsValidName("ds@#$#$#"), "Returns True for invalid cluster name")
}

func TestSSHExecute(t *testing.T) {
	// make a dummy ssh server with ssh keypair auth
	// payloadSSH := SSHPayload{UserName: "dipankar", PathPrivateKey: GetPath(SSH_PATH, "dcs"), PublicIP: "0.0.0.0", Output: ""}
	// payloadSSH.SSHExecute()
}

// TODO: Add testing for credentials
func TestSaveCred(t *testing.T) {}
