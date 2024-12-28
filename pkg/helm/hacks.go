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

package helm

import (
	"context"
	"os"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/logger"
	localStore "github.com/ksctl/ksctl/pkg/storage/host"
)

type CustomLogger struct {
	Logger logger.Logger
	ctx    context.Context
}

func (l *CustomLogger) HelmDebugf(format string, v ...any) {
	l.Logger.ExternalLogHandlerf(l.ctx, logger.LogInfo, format+"\n", v...)
}

func patchHelmDirectories(ctx context.Context, log logger.Logger, client *Client) error {
	usr, err := os.UserHomeDir()
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(ctx, "failed to get the user home dir", "Reason", err),
		)
	}

	store := localStore.NewClient(ctx, log)

	pathConfig := []string{usr, ".config", "helm"}
	_, okConfig := store.PresentDirectory(pathConfig)
	if !okConfig {
		if _err := store.CreateDirectory(pathConfig); _err != nil {
			return _err
		}
	}

	pathRegistry := []string{usr, ".config", "helm", "registry"}
	_, okReg := store.PresentDirectory(pathRegistry)
	if !okReg {
		if _err := store.CreateDirectory(pathRegistry); _err != nil {
			return _err
		}
	}

	pathConfig = append(pathConfig, "repositories.yaml")
	configPath, _err := store.CreateFileIfNotPresent(pathConfig)
	if _err != nil {
		return _err
	}

	pathRegistry = append(pathRegistry, "config.json")
	registryPath, _err := store.CreateFileIfNotPresent(pathRegistry)
	if _err != nil {
		return _err
	}

	if err := store.Kill(); err != nil {
		return err
	}

	if _err := os.Setenv("HELM_DRIVER", "secrets"); _err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(ctx, "failed to set env var", "Reason", _err),
		)
	}
	client.settings.RepositoryConfig = configPath
	client.settings.RegistryConfig = registryPath
	log.Print(ctx, "Updated the Helm configuration settings")

	return nil
}
