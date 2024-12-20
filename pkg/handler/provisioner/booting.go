package provisioner

import (
	"context"
	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/validation"
	"runtime/debug"
	"sort"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"

	"github.com/ksctl/ksctl/pkg/poller"

	localstate "github.com/ksctl/ksctl/pkg/storage/host"
	kubernetesstate "github.com/ksctl/ksctl/pkg/storage/kubernetes"
	externalmongostate "github.com/ksctl/ksctl/pkg/storage/mongodb"

	"github.com/ksctl/ksctl/pkg/consts"
)

var (
	controllerCtx context.Context

	stateDocument *statefile.StorageDocument
)

type managerInfo struct {
	log    logger.Logger
	client *Client
}

func (manager *managerInfo) startPoller(ctx context.Context) error {
	if _, ok := config.IsContextPresent(ctx, consts.KsctlTestFlagKey); !ok {
		poller.InitSharedGithubReleasePoller()
	} else {
		poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
			vers := []string{"v0.0.1"}

			if org == "etcd-io" && repo == "etcd" {
				vers = append(vers, "v3.5.15")
			}

			if org == "k3s-io" && repo == "k3s" {
				vers = append(vers, "v1.30.3+k3s1")
			}

			if org == "kubernetes" && repo == "kubernetes" {
				vers = append(vers, "v1.31.0")
			}

			sort.Slice(vers, func(i, j int) bool {
				return vers[i] > vers[j]
			})

			return vers, nil
		})
	}

	return nil
}

func (manager *managerInfo) initStorage(ctx context.Context) error {
	if !validation.ValidateStorage(manager.client.Metadata.StateLocation) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidStorageProvider,
			manager.log.NewError(
				controllerCtx, "Problem in validation", "storage", manager.client.Metadata.StateLocation,
			),
		)
	}
	switch manager.client.Metadata.StateLocation {
	case consts.StoreLocal:
		manager.client.Storage = localstate.NewClient(ctx, manager.log)
	case consts.StoreExtMongo:
		manager.client.Storage = externalmongostate.NewClient(ctx, manager.log)
	case consts.StoreK8s:
		manager.client.Storage = kubernetesstate.NewClient(ctx, manager.log)
	}

	if err := manager.client.Storage.Connect(); err != nil {
		return err
	}
	manager.log.Debug(ctx, "initialized storageFactory")
	return nil
}

func panicCatcher(log logger.Logger) {
	if r := recover(); r != nil {
		log.Error("Failed to recover stack trace", "error", r)
		debug.PrintStack()
	}
}

func (manager *managerInfo) setupConfigurations() error {

	if err := manager.validationFields(manager.client.Metadata); err != nil {
		return err
	}

	if err := validation.IsValidName(controllerCtx, manager.log, manager.client.Metadata.ClusterName); err != nil {
		return err
	}
	return nil
}
