package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	cloudControlRes "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"github.com/ksctl/ksctl/poller"
	"github.com/stretchr/testify/assert"
)

var (
	storeHA types.StorageFactory

	fakeClient         *PreBootstrap
	dir                = filepath.Join(os.TempDir(), "ksctl-bootstrap-test")
	fakeStateFromCloud cloudControlRes.CloudResourceState

	parentCtx    context.Context
	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
)

func initPoller() {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		if org == "etcd-io" && repo == "etcd" {
			vers = append(vers, "v3.5.15")
		}
		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		return vers, nil
	})
}

func initClients() {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)
	parentCtx = context.WithValue(parentCtx, consts.KsctlTestFlagKey, "true")

	mainState := &storageTypes.StorageDocument{}
	if err := helpers.CreateSSHKeyPair(parentCtx, parentLogger, mainState); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	fakeStateFromCloud = cloudControlRes.CloudResourceState{
		SSHState: cloudControlRes.SSHInfo{
			PrivateKey: mainState.SSHKeyPair.PrivateKey,
			UserName:   "fakeuser",
		},
		Metadata: cloudControlRes.Metadata{
			ClusterName: "fake",
			Provider:    consts.CloudAzure,
			Region:      "fake",
			ClusterType: consts.ClusterTypeHa,
		},
		// public IPs
		IPv4ControlPlanes: []string{"A.B.C.4", "A.B.C.5", "A.B.C.6"},
		IPv4DataStores:    []string{"A.B.C.3"},
		IPv4WorkerPlanes:  []string{"A.B.C.2"},
		IPv4LoadBalancer:  "A.B.C.1",

		// Private IPs
		PrivateIPv4ControlPlanes: []string{"192.168.X.7", "192.168.X.9", "192.168.X.10"},
		PrivateIPv4DataStores:    []string{"192.168.5.2"},
		PrivateIPv4LoadBalancer:  "192.168.X.1",
	}

	fakeClient = NewPreBootStrap(parentCtx, parentLogger, mainState)
	if fakeClient == nil {
		panic("unable to initialize")
	}

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "fake", consts.ClusterTypeHa)
	_ = storeHA.Connect()
}

func TestMain(m *testing.M) {

	initPoller()
	initClients()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(fakeStateFromCloud, storeHA, consts.OperationCreate), nil, "should be initlize the state")
	noDS := len(fakeStateFromCloud.IPv4DataStores)

	err := fakeClient.ConfigureLoadbalancer(storeHA)
	if err != nil {
		t.Fatalf("Configure Datastore unable to operate %v", err)
	}

	assert.Equal(t, mainStateDocument.K8sBootstrap.B.HAProxyVersion, "3.0", "should be equal")

	for no := 0; no < noDS; no++ {
		err := fakeClient.ConfigureDataStore(no, storeHA)
		if err != nil {
			t.Fatalf("Configure Datastore unable to operate %v", err)
		}
	}

	assert.Equal(t, mainStateDocument.K8sBootstrap.B.EtcdVersion, "v3.5.15", "should be equal")
}
