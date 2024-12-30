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

package k3s

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/pkg/statefile"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/poller"
	"github.com/ksctl/ksctl/pkg/provider"
	"github.com/ksctl/ksctl/pkg/ssh"
	"github.com/ksctl/ksctl/pkg/storage"
	localstate "github.com/ksctl/ksctl/pkg/storage/host"
)

var (
	storeHA storage.Storage

	fakeClient         *K3s
	dir                = filepath.Join(os.TempDir(), "ksctl-k3s-test")
	fakeStateFromCloud provider.CloudResourceState

	parentCtx    context.Context
	parentLogger logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func initPoller() {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		if org == "k3s-io" && repo == "k3s" {
			vers = append(vers, "v1.30.3+k3s1")
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
		ClusterType:   consts.ClusterTypeHa,

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
