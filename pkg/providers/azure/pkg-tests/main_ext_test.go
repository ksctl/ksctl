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

package pkg_tests_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksctl/ksctl/internal/cloudproviders/azure"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

var (
	fakeClientHA types.CloudFactory
	storeHA      types.StorageFactory

	fakeClientManaged types.CloudFactory
	storeManaged      types.StorageFactory

	fakeClientVars *azure.AzureProvider
	storeVars      types.StorageFactory
	parentCtx      context.Context
	parentLogger   types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-azure-pkg-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = azure.NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "fake",
		Provider:    consts.CloudAzure,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, azure.ProvideClient)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAzure, "fake", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}

func genResourceGroup(clusterName string, clusterType string) string {
	return fmt.Sprintf("ksctl-resgrp-%s-%s", clusterType, clusterName)
}
