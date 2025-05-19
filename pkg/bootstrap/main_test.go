// Copyright 2024 Ksctl Authors
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

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/ssh"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	localstate "github.com/ksctl/ksctl/v2/pkg/storage/host"
	"github.com/stretchr/testify/assert"
)

var (
	storeHA storage.Storage

	fakeClient         *PreBootstrap
	dir                = filepath.Join(os.TempDir(), "ksctl-bootstrap-test")
	fakeStateFromCloud provider.CloudResourceState

	parentCtx    context.Context
	parentLogger logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func initPoller(c cache.Cache) {
	poller.InitSharedGithubReleaseFakePoller(c, func(org, repo string) ([]string, error) {
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

	mainState := &statefile.StorageDocument{}
	if err := ssh.CreateSSHKeyPair(parentCtx, parentLogger, mainState); err != nil {
		parentLogger.Error(err.Error())
		os.Exit(1)
	}
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

	storeHA = localstate.NewClient(parentCtx, parentLogger)
	_ = storeHA.Setup(consts.CloudAzure, "fake", "fake", consts.ClusterTypeSelfMang)

	fakeClient = NewPreBootStrap(parentCtx, parentLogger, mainState, storeHA)
	if fakeClient == nil {
		panic("unable to initialize")
	}

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

func TestOverallScriptsCreation(t *testing.T) {
	assert.Equal(t, fakeClient.Setup(&fakeStateFromCloud, consts.OperationCreate), nil, "should be initlize the state")
	noDS := len(fakeStateFromCloud.IPv4DataStores)

	err := fakeClient.ConfigureLoadbalancer()
	if err != nil {
		t.Fatalf("Configure Datastore unable to operate %v", err)
	}

	assert.Equal(t, *fakeClient.state.Versions.HAProxy, "3.0", "should be equal")

	for no := 0; no < noDS; no++ {
		err := fakeClient.ConfigureDataStore(no, "")
		if err != nil {
			t.Fatalf("Configure Datastore unable to operate %v", err)
		}
	}

	assert.Equal(t, *fakeClient.state.Versions.Etcd, "v3.5.15", "should be equal")
}
