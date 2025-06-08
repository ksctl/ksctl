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

package main

import (
	"context"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/cache"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/storage/host"
	"github.com/ksctl/ksctl/v2/pkg/storage/mongodb"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"

	controllerCommon "github.com/ksctl/ksctl/v2/pkg/handler/cluster/common"
	controllerManaged "github.com/ksctl/ksctl/v2/pkg/handler/cluster/managed"
	controllerSelfManaged "github.com/ksctl/ksctl/v2/pkg/handler/cluster/selfmanaged"

	addonClusterMgt "github.com/ksctl/ksctl/v2/pkg/handler/addons"
)

var (
	l   logger.Logger
	ctx context.Context
)

func main() {
	ctx = context.WithValue(
		context.Background(),
		consts.KsctlModuleNameKey,
		"e2e",
	)

	ksctlConfig := context.WithValue(
		context.TODO(),
		consts.KsctlContextUser,
		"e2e",
	)

	timer := time.Now()

	flags := os.Getenv("E2E_FLAGS")
	rFlags := strings.Split(flags, ";")

	verbosityLevel := 0

	if slices.Contains(rFlags, "debug") {
		verbosityLevel = -1
	}

	for _, _flags := range rFlags {
		if strings.HasPrefix(_flags, "core_component_overridings=") {
			v := strings.TrimPrefix(_flags, "core_component_overridings=")

			ctx = context.WithValue(
				ctx,
				consts.KsctlComponentOverrides,
				v,
			)
		}
	}

	l = logger.NewStructuredLogger(verbosityLevel, os.Stdout)

	operation, meta := GetReqPayload(l)

	if meta.Provider == consts.CloudAws {
		ksctlConfig = CredsAws(ksctlConfig)
	}
	if meta.Provider == consts.CloudAzure {
		ksctlConfig = CredsAzure(ksctlConfig)
	}

	cc := cache.NewInMemCache(ctx)
	kscConfig := controller.KsctlWorkerConfiguration{
		WorkerCtx:   ksctlConfig,
		PollerCache: cc,
	}

	if meta.StateLocation == consts.StoreExtMongo { // mongodb storage
		client, err := mongodb.NewDBClient(ctx, CredsMongo(ctx))
		if err != nil {
			l.Error("unable to initialize the mongodb client", "Reason", err)
			os.Exit(1)
		}
		kscConfig.Storage, err = client.NewDatabaseClient(ksctlConfig, l)
		if err != nil {
			l.Error("unable to initialize the mongodb client", "Reason", err)
			os.Exit(1)
		}
	} else { // local storage
		kscConfig.Storage = host.NewClient(ksctlConfig, l)
	}

	defer cc.Close()

	l.Print(ctx, "Testing starting...")

	switch operation {
	case OpCreate, OpDelete, OpScaleUp, OpScaleDown:
		switch meta.ClusterType {
		case consts.ClusterTypeSelfMang:

			if meta.Provider == consts.CloudLocal {
				err := l.NewError(ctx, "ha not supported for local")
				l.Error("handled error", "catch", err)
				os.Exit(1)
			}
			managerClient, err := controllerSelfManaged.NewController(
				ctx,
				l,
				kscConfig,
				&controller.Client{
					Metadata: meta,
				},
			)
			if err != nil {
				l.Error("unable to initialize the ksctl manager", "Reason", err)
				os.Exit(1)
			}
			switch operation {
			case OpCreate:
				createHACluster(managerClient)
			case OpDelete:
				deleteHACluster(managerClient)
			case OpScaleUp:
				scaleupHACluster(managerClient)
			case OpScaleDown:
				scaleDownHACluster(managerClient)
			}

		case consts.ClusterTypeMang:
			managerClient, err := controllerManaged.NewController(
				ctx,
				l,
				kscConfig,
				&controller.Client{
					Metadata: meta,
				},
			)
			if err != nil {
				l.Error("unable to initialize the ksctl manager", "Reason", err)
				os.Exit(1)
			}
			switch operation {
			case OpCreate:
				createManagedCluster(managerClient)
			case OpDelete:
				deleteManagedCluster(managerClient)
			}
		}

	case OpGet, OpSwitch, OpInfo:
		managerClient, err := controllerCommon.NewController(
			ctx,
			l,
			kscConfig,
			&controller.Client{
				Metadata: meta,
			},
		)
		if err != nil {
			l.Error("unable to initialize the ksctl manager", "Reason", err)
			os.Exit(1)
		}
		switch operation {
		case OpGet:
			getClusters(managerClient)
		case OpInfo:
			infoClusters(managerClient)
		case OpSwitch:
			switchCluster(managerClient)
		}
	case OpEnableClusterMgt, OpDisableClusterMgt:
		cc, err := addonClusterMgt.NewController(
			ctx,
			l,
			kscConfig,
			&controller.Client{
				Metadata: meta,
			},
		)
		if err != nil {
			l.Error("unable to initialize the ksctl manager", "Reason", err)
			os.Exit(1)
		}
		switch operation {
		case OpEnableClusterMgt:
			enableClusterMgtAddon(cc)
		case OpDisableClusterMgt:
			disableClusterMgtAddon(cc)
		}

	default:
		l.Error("This operation is not supported")
		os.Exit(1)
	}

	l.Print(ctx, "Testing Completed", "TimeTaken", time.Since(timer).String())
}
