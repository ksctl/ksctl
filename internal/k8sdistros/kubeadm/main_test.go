package kubeadm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	cloudControlRes "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"github.com/ksctl/ksctl/poller"
)

var (
	storeHA types.StorageFactory

	fakeClient         *Kubeadm
	dir                = filepath.Join(os.TempDir(), "ksctl-kubeadm-test")
	fakeStateFromCloud cloudControlRes.CloudResourceState
	parentCtx          context.Context
	parentLogger       types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
)

func NewClientHelper(x cloudControlRes.CloudResourceState, state *storageTypes.StorageDocument) *Kubeadm {
	kubeadmCtx = parentCtx
	log = parentLogger

	mainStateDocument = state
	mainStateDocument.K8sBootstrap = &storageTypes.KubernetesBootstrapState{}
	var err error
	mainStateDocument.K8sBootstrap.B.CACert, mainStateDocument.K8sBootstrap.B.EtcdCert, mainStateDocument.K8sBootstrap.B.EtcdKey, err = helpers.GenerateCerts(parentCtx, parentLogger, x.PrivateIPv4DataStores)
	if err != nil {
		return nil
	}

	mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes = x.IPv4ControlPlanes
	mainStateDocument.K8sBootstrap.B.PrivateIPs.ControlPlanes = x.PrivateIPv4ControlPlanes

	mainStateDocument.K8sBootstrap.B.PublicIPs.DataStores = x.IPv4DataStores
	mainStateDocument.K8sBootstrap.B.PrivateIPs.DataStores = x.PrivateIPv4DataStores

	mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes = x.IPv4WorkerPlanes

	mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer = x.IPv4LoadBalancer
	mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer = x.PrivateIPv4LoadBalancer
	mainStateDocument.K8sBootstrap.B.SSHInfo = x.SSHState

	return &Kubeadm{mu: &sync.Mutex{}}
}

func initPoller() {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		if org == "kubernetes" && repo == "kubernetes" {
			vers = append(vers, "v1.31.0")
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

	fakeClient = NewClientHelper(fakeStateFromCloud, &storageTypes.StorageDocument{})
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
