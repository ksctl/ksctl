package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/storage/localstate"

	. "github.com/kubesimplify/ksctl/api/utils/consts"
	"gotest.tools/assert"
)

var (
	dir = fmt.Sprintf("%s/ksctl-k3s-test", os.TempDir())
)

func TestConsts(t *testing.T) {
	assert.Equal(t, string(CLOUD_CIVO), "civo", "civo constant not correct assigned")
	assert.Equal(t, string(CLOUD_AZURE), "azure", "azure constant not correct assgined")
	assert.Equal(t, string(CLOUD_LOCAL), "local", "local constant not correct assgined")
	assert.Equal(t, string(CLOUD_AWS), "aws", "aws constant not correct assgined")
	assert.Equal(t, string(K8S_K3S), "k3s", "k3s constant not correct assgined")
	assert.Equal(t, string(K8S_KUBEADM), "kubeadm", "kubeadm constant not correct assgined")
	assert.Equal(t, string(STORE_LOCAL), "local", "local constant not correct assgined")
	assert.Equal(t, string(STORE_REMOTE), "remote", "remote constant not correct assgined")
	assert.Equal(t, string(ROLE_CP), "controlplane", "controlplane constant not correct assgined")
	assert.Equal(t, string(ROLE_LB), "loadbalancer", "loadbalancer constant not correct assgined")
	assert.Equal(t, string(ROLE_DS), "datastore", "datastore constant not correct assgined")
	assert.Equal(t, string(ROLE_WP), "workerplane", "workerplane constant not correct assgined")
	assert.Equal(t, string(CLUSTER_TYPE_HA), "ha", "HA constant not correct assgined")
	assert.Equal(t, string(CLUSTER_TYPE_MANG), "managed", "Managed constant not correct assgined")
	assert.Equal(t, string(OPERATION_STATE_CREATE), "create", "operation create constant not correct assgined")
	assert.Equal(t, string(OPERATION_STATE_GET), "get", "operation get constant not correct assgined")
	assert.Equal(t, string(OPERATION_STATE_DELETE), "delete", "operation delete constant not correct assgined")
	assert.Equal(t, uint8(CLUSTER_PATH), uint8(1), "cluster_path constant not correct assgined")
	assert.Equal(t, uint8(OTHER_PATH), uint8(3), "other_path constant not correct assgined")
	assert.Equal(t, uint8(SSH_PATH), uint8(2), "ssh_path constant not correct assgined")
	assert.Equal(t, uint8(CREDENTIAL_PATH), uint8(0), "credential_path constant not correct assgined")
	assert.Equal(t, uint8(EXEC_WITHOUT_OUTPUT), uint8(0), "exec_without_output constant not correct assgined")
	assert.Equal(t, uint8(EXEC_WITH_OUTPUT), uint8(1), "exec_without_output constant not correct assgined")
}

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
			assert.Equal(t, expectedPath, getKubeconfig(KsctlCloud(provider), "Yy zz")) // must return empty string as its invalid provider
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
			assert.Equal(t, expectedPath, getKubeconfig(KsctlCloud(provider), "Yy zz")) // must return empty string as its invalid provider
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
		_ = os.RemoveAll(GetPath(OTHER_PATH, KsctlCloud(provider), CLUSTER_TYPE_HA))
	})

	path := GetPath(OTHER_PATH, KsctlCloud(provider), CLUSTER_TYPE_HA, clusterName+" "+clusterRegion)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Unable to create dummy folder")
	}
	ksctl := resources.KsctlClient{Storage: localstate.InitStorage(false)}
	if _, err := CreateSSHKeyPair(ksctl.Storage, KsctlCloud(provider), clusterName+" "+clusterRegion); err != nil {
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
	var storage resources.StorageFactory = localstate.InitStorage(false)
	assert.Equal(t, os.MkdirAll(GetPath(CLUSTER_PATH, CLOUD_AZURE, CLUSTER_TYPE_HA, "abcd"), 0755), nil, "create folders")
	_ = os.Setenv(string(KSCTL_CUSTOM_DIR_ENABLED), dir)
	azHA := GetPath(CLUSTER_PATH, CLOUD_AZURE, CLUSTER_TYPE_HA, "abcd")

	if err := os.MkdirAll(azHA, 0755); err != nil {
		t.Fatalf("Reason: %v", err)
	}
	fmt.Println("Created tmp directories")

	_, err := CreateSSHKeyPair(storage, CLOUD_AZURE, "abcd")
	if err != nil {
		t.Fatalf("Reason: %v", err)
	}
	var sshTest SSHCollection = &SSHPayload{}
	sshTest.LocPrivateKey(GetPath(SSH_PATH, CLOUD_AZURE, "ha", "abcd"))
	sshTest.Username("fake")
	assert.Assert(t, sshTest.Flag(EXEC_WITHOUT_OUTPUT).Script("").
		IPv4("A.A.A.A").
		FastMode(true).SSHExecute(storage) != nil, "ssh should fail")

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

}

// TODO: Add testing for credentials
func TestSaveCred(t *testing.T) {}
