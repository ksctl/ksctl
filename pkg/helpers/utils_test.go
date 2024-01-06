package helpers

import (
	"fmt"
	"github.com/gookit/goutil/dump"
	"github.com/kubesimplify/ksctl/internal/storage/types"
	"os"
	"runtime"
	"testing"

	. "github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"

	"gotest.tools/assert"
)

var (
	dir                         = fmt.Sprintf("%s/ksctl-k3s-test", os.TempDir())
	log resources.LoggerFactory = func() resources.LoggerFactory {
		var l resources.LoggerFactory = logger.NewDefaultLogger(-1, os.Stdout)
		l.SetPackageName("utils")
		return l
	}()
	mainStateDoc = &types.StorageDocument{}
)

func TestConsts(t *testing.T) {
	assert.Equal(t, string(CloudCivo), "civo", "civo constant not correct assigned")
	assert.Equal(t, string(CloudAzure), "azure", "azure constant not correct assgined")
	assert.Equal(t, string(CloudLocal), "local", "local constant not correct assgined")
	assert.Equal(t, string(CloudAws), "aws", "aws constant not correct assgined")
	assert.Equal(t, string(K8sK3s), "k3s", "k3s constant not correct assgined")
	assert.Equal(t, string(K8sKubeadm), "kubeadm", "kubeadm constant not correct assgined")
	assert.Equal(t, string(StoreLocal), "local", "local constant not correct assgined")
	assert.Equal(t, string(StoreExtMongo), "external-mongo", "remote constant not correct assgined")
	assert.Equal(t, string(RoleCp), "controlplane", "controlplane constant not correct assgined")
	assert.Equal(t, string(RoleLb), "loadbalancer", "loadbalancer constant not correct assgined")
	assert.Equal(t, string(RoleDs), "datastore", "datastore constant not correct assgined")
	assert.Equal(t, string(RoleWp), "workerplane", "workerplane constant not correct assgined")
	assert.Equal(t, string(ClusterTypeHa), "ha", "HA constant not correct assgined")
	assert.Equal(t, string(ClusterTypeMang), "managed", "Managed constant not correct assgined")
	assert.Equal(t, string(OperationStateCreate), "create", "operation create constant not correct assgined")
	assert.Equal(t, string(OperationStateGet), "get", "operation get constant not correct assgined")
	assert.Equal(t, string(OperationStateDelete), "delete", "operation delete constant not correct assgined")
	assert.Equal(t, uint8(UtilClusterPath), uint8(1), "cluster_path constant not correct assgined")
	assert.Equal(t, uint8(UtilOtherPath), uint8(3), "other_path constant not correct assgined")
	assert.Equal(t, uint8(UtilSSHPath), uint8(2), "ssh_path constant not correct assgined")
	assert.Equal(t, uint8(UtilCredentialPath), uint8(0), "credential_path constant not correct assgined")
	assert.Equal(t, uint8(UtilExecWithoutOutput), uint8(0), "exec_without_output constant not correct assgined")
	assert.Equal(t, uint8(UtilExecWithOutput), uint8(1), "exec_without_output constant not correct assgined")

	assert.Equal(t, string(CNIAzure), "azure", "missmatch")
	assert.Equal(t, string(CNIKind), "kind", "missmatch")
	assert.Equal(t, string(CNINone), "none", "missmatch")
	assert.Equal(t, string(CNICilium), "cilium", "missmatch")
	assert.Equal(t, string(CNIFlannel), "flannel", "missmatch")
	assert.Equal(t, string(CNIKubenet), "kubenet", "missmatch")
}

func TestGetUsername(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, os.Getenv(UserDir), GetUserName(), "Unable to fetch correct username")
	} else {
		assert.Equal(t, os.Getenv(UserDir), GetUserName(), "Unable to fetch correct username")
	}
}

func TestCNIValidation(t *testing.T) {
	cnitests := map[string]bool{
		string(CNIAzure):   true,
		string(CNIKubenet): true,
		string(CNIFlannel): true,
		string(CNICilium):  true,
		string(CNIKind):    true,
		"abcd":             false,
		"":                 true,
	}
	for k, v := range cnitests {
		assert.Equal(t, v, ValidCNIPlugin(KsctlValidCNIPlugin(k)), "")
	}
}

func TestCreateSSHKeyPair(t *testing.T) {
	err := CreateSSHKeyPair(log, mainStateDoc)
	if err != nil {
		t.Fatal(err)
	}
	dump.Println(mainStateDoc.SSHKeyPair)
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

	var sshTest SSHCollection = &SSHPayload{}
	sshTest.Username("fake")
	sshTest.PrivateKey(mainStateDoc.SSHKeyPair.PrivateKey)
	assert.Assert(t, sshTest.Flag(UtilExecWithoutOutput).Script("").
		IPv4("A.A.A.A").
		FastMode(true).SSHExecute(log) != nil, "ssh should fail")

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

}

// TODO: Add testing for credentials
func TestSaveCred(t *testing.T) {}
