package helmclient

import (
	"context"
	localStore "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"os"
)

type CustomLogger struct {
	Logger types.LoggerFactory
	ctx    context.Context
}

func (l *CustomLogger) HelmDebugf(format string, v ...any) {
	l.Logger.ExternalLogHandlerf(l.ctx, consts.LogInfo, format+"\n", v...)
}

func patchHelmDirectories(ctx context.Context, log types.LoggerFactory, client *HelmClient) error {
	usr, err := os.UserHomeDir()
	if err != nil {
		return ksctlErrors.ErrUnknown.Wrap(
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
		return ksctlErrors.ErrUnknown.Wrap(
			log.NewError(ctx, "failed to set env var", "Reason", _err),
		)
	}
	client.settings.RepositoryConfig = configPath
	client.settings.RegistryConfig = registryPath
	log.Print(ctx, "Updated the Helm configuration settings")

	return nil
}
