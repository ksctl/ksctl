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
	"os"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	controllerManaged "github.com/ksctl/ksctl/pkg/handler/cluster/managed"
)

func createManagedCluster(ksctlClient *controllerManaged.Controller) {
	l.Print(ctx, "Started to Create Cluster...")

	err := ksctlClient.Create()
	if err != nil {
		if ksctlErrors.IsInvalidCloudProvider(err) {
			l.Error("problem is invalid cloud provider")
		}
		if ksctlErrors.IsInvalidResourceName(err) {
			l.Error("problem from resource name")
		}
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}

func deleteManagedCluster(ksctlClient *controllerManaged.Controller) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.Delete()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}
