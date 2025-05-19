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

package kubeadm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/certs"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/ssh"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	localstate "github.com/ksctl/ksctl/v2/pkg/storage/host"
)

var (
	storeHA storage.Storage

	fakeClient         *Kubeadm
	dir                = filepath.Join(os.TempDir(), "ksctl-kubeadm-test")
	fakeStateFromCloud provider.CloudResourceState
	parentCtx          context.Context
	parentLogger       logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func NewClientHelper(x provider.CloudResourceState, state *statefile.StorageDocument) *Kubeadm {
	p := &Kubeadm{mu: &sync.Mutex{}}

	p.ctx = parentCtx
	p.l = parentLogger
	p.state = state
	p.state.K8sBootstrap = &statefile.KubernetesBootstrapState{}

	var err error
	p.state.K8sBootstrap.B.CACert, p.state.K8sBootstrap.B.EtcdCert, p.state.K8sBootstrap.B.EtcdKey, err = certs.GenerateCerts(parentCtx, parentLogger, x.PrivateIPv4DataStores)
	if err != nil {
		return nil
	}

	p.state.K8sBootstrap.B.PublicIPs.ControlPlanes = x.IPv4ControlPlanes
	p.state.K8sBootstrap.B.PrivateIPs.ControlPlanes = x.PrivateIPv4ControlPlanes
	p.state.K8sBootstrap.B.PublicIPs.DataStores = x.IPv4DataStores
	p.state.K8sBootstrap.B.PrivateIPs.DataStores = x.PrivateIPv4DataStores
	p.state.K8sBootstrap.B.PublicIPs.WorkerPlanes = x.IPv4WorkerPlanes
	p.state.K8sBootstrap.B.PublicIPs.LoadBalancer = x.IPv4LoadBalancer
	p.state.K8sBootstrap.B.PrivateIPs.LoadBalancer = x.PrivateIPv4LoadBalancer
	p.state.K8sBootstrap.B.SSHInfo = statefile.SSHInfo{
		UserName:   x.SSHUserName,
		PrivateKey: x.SSHPrivateKey,
	}

	return p
}

func initPoller(c cache.Cache) {
	poller.InitSharedGithubReleaseFakePoller(c, func(org, repo string) ([]string, error) {
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

	mainState := &statefile.StorageDocument{}
	if err := ssh.CreateSSHKeyPair(parentCtx, parentLogger, mainState); err != nil {
		parentLogger.Error(err.Error())
		os.Exit(1)
	}

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "fake", consts.ClusterTypeSelfMang)
	fakeStateFromCloud = provider.CloudResourceState{
		SSHPrivateKey: mainState.SSHKeyPair.PrivateKey,
		SSHUserName:   "fakeuser",
		ClusterName:   "fake",
		Provider:      consts.CloudAzure,
		Region:        "fake",
		ClusterType:   consts.ClusterTypeSelfMang,

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

	fakeClient = NewClientHelper(fakeStateFromCloud, mainState)
	if fakeClient == nil {
		panic("unable to initialize")
	}

	fakeClient.store = storeHA

}

func TestMain(m *testing.M) {

	cc := cache.NewInMemCache(context.TODO())
	defer cc.Close()
	initPoller(cc)
	initClients()
	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
