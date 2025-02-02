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

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	localstate "github.com/ksctl/ksctl/v2/pkg/storage/host"
)

var (
	fakeClientHA *Provider
	storeHA      storage.Storage

	fakeClientManaged *Provider
	storeManaged      storage.Storage
	parentCtx         context.Context
	fakeClientVars    *Provider
	storeVars         storage.Storage

	parentLogger logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-aws-test")
)

func TestMain(m *testing.M) {

	parentCtx = context.WithValue(
		context.TODO(),
		consts.KsctlCustomDirLoc,
		dir)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAws, "fake-region", "demo", consts.ClusterTypeSelfMang)
	_ = storeVars.Connect()

	v, err := json.Marshal(statefile.CredentialsAws{
		AccessKeyId:     "fake",
		SecretAccessKey: "fake",
	})
	if err != nil {
		panic(err)
	}
	parentCtx = context.WithValue(parentCtx, consts.KsctlAwsCredentials, v)

	fakeClientVars, _ = NewClient(
		parentCtx,
		parentLogger,
		controller.Metadata{
			ClusterName: "demo",
			Region:      "fake-region",
			Provider:    consts.CloudAws,
			ClusterType: consts.ClusterTypeSelfMang,
		},
		&statefile.StorageDocument{},
		storeVars,
		ProvideClient)

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
