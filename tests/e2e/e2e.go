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

	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	l   types.LoggerFactory
	ctx context.Context
)

func main() {
	ctx = context.WithValue(
		context.WithValue(
			context.Background(),
			consts.KsctlContextUserID,
			"e2e",
		),
		consts.KsctlModuleNameKey,
		"e2e",
	)

	timer := time.Now()

	flags := os.Getenv("E2E_FLAGS")
	rFlags := strings.Split(flags, ";")

	verbosityLevel := 0

	if slices.Contains[[]string, string](rFlags, "debug") {
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

	l.Print(ctx, "Testing starting...")

	switch operation {
	case OpCreate, OpDelete, OpScaleUp, OpScaleDown:
		switch meta.IsHA {
		case true:

			if meta.Provider == consts.CloudLocal {
				err := l.NewError(ctx, "ha not supported for local")
				l.Error("handled error", "catch", err)
				os.Exit(1)
			}
			managerClient, err := controllers.NewManagerClusterSelfManaged(
				ctx,
				l,
				&types.KsctlClient{
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

		case false:
			managerClient, err := controllers.NewManagerClusterManaged(
				ctx,
				l,
				&types.KsctlClient{
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

	case OpCreds, OpGet, OpSwitch, OpInfo:
		managerClient, err := controllers.NewManagerClusterKsctl(
			ctx,
			l,
			&types.KsctlClient{
				Metadata: meta,
			},
		)
		if err != nil {
			l.Error("unable to initialize the ksctl manager", "Reason", err)
			os.Exit(1)
		}
		switch operation {
		case OpCreds:
			creds(managerClient)
		case OpGet:
			getClusters(managerClient)
		case OpInfo:
			infoClusters(managerClient)
		case OpSwitch:
			switchCluster(managerClient)
		}

	default:
		l.Error("This operation is not supported")
		os.Exit(1)
	}

	l.Print(ctx, "Testing Completed", "TimeTaken", time.Since(timer).String())
}
