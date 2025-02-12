// Copyright 2025 Ksctl Authors
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

	addonClusterMgt "github.com/ksctl/ksctl/v2/pkg/handler/addons"
)

func enableClusterMgtAddon(ksctlClient *addonClusterMgt.AddonController) {
	l.Print(ctx, "Exec ksctl addons disable clustermgt...")

	vers, err := ksctlClient.ListAvailableVersions("kcm")
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}

	cc, err := ksctlClient.GetAddon("kcm")
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}

	if err := cc.Install(vers[0]); err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}

func disableClusterMgtAddon(ksctlClient *addonClusterMgt.AddonController) {
	l.Print(ctx, "Exec ksctl addons disable clustermgt...")

	cc, err := ksctlClient.GetAddon("kcm")
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}

	if err := cc.Uninstall(); err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}
