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

package azure

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	localstate "github.com/ksctl/ksctl/pkg/storage/host"
)

var (
	fakeClientHA *Provider
	storeHA      storage.Storage

	fakeClientManaged *Provider
	storeManaged      storage.Storage

	fakeClientVars *Provider
	storeVars      storage.Storage
	parentCtx      context.Context
	parentLogger   logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-azure-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAzure, "fake", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	fakeClientVars, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo",
			Region:      "fake",
			Provider:    consts.CloudAzure,
			IsHA:        true,
		},
		&statefile.StorageDocument{},
		storeVars,
		ProvideClient,
	)

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
