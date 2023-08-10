package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/utils"

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

	path := GetPath(OTHER_PATH, provider, utils.CLUSTER_TYPE_HA, clusterName+" "+clusterRegion)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Unable to create dummy folder")
	}
	ksctl := resources.KsctlClient{Storage: localstate.InitStorage(false)}
	if _, err := CreateSSHKeyPair(ksctl.Storage, provider, clusterName+" "+clusterRegion); err != nil {
		t.Fatalf("Unable to create SSH keypair")
	}
}

func TestIsValidClusterName(T *testing.T) {
	errorStr := fmt.Errorf("CLUSTER NAME INVALID")
	assert.Equal(T, nil, IsValidName("demo"), "Returns false for valid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("Dem-o234").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, nil, IsValidName("d-234"), "Returns false for valid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("234").Error(), "Returns true for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("-2342").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("demo-").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("dscdscsd-#$#$#").Error(), "Returns True for invalid cluster name")
	assert.Equal(T, errorStr.Error(), IsValidName("ds@#$#$#").Error(), "Returns True for invalid cluster name")
}

func TestSSHExecute(t *testing.T) {
	// make a dummy ssh server with ssh keypair auth
	// payloadSSH := SSHPayload{UserName: "dipankar", PathPrivateKey: GetPath(SSH_PATH, "dcs"), PublicIP: "0.0.0.0", Output: ""}
	// payloadSSH.SSHExecute()
}

// TODO: Add testing for credentials
func TestSaveCred(t *testing.T) {}
