package helmclient

import (
	"context"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"os"
)

type HelmClient struct {
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
	ctx          context.Context
	log          types.LoggerFactory
}

func NewKubeconfigHelmClient(ctx context.Context, log types.LoggerFactory, kubeconfig string) (client *HelmClient, err error) {
	client = new(HelmClient)

	client.settings = cli.New()
	client.settings.Debug = true
	client.log = log
	client.ctx = ctx
	if err := patchHelmDirectories(ctx, log, client); err != nil {
		return
	}

	client.actionConfig = new(action.Configuration)

	_log := &CustomLogger{Logger: log, ctx: ctx}

	if err := client.actionConfig.Init(NewRESTClientGetter(client.settings.Namespace(), kubeconfig), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), _log.HelmDebugf); err != nil {
		return nil, ksctlErrors.ErrFailedHelmClient.Wrap(
			log.NewError(ctx, "failed to init kubeconfig based helm client", "Reason", err),
		)
	}
	return client, nil
}

func NewInClusterHelmClient(ctx context.Context, log types.LoggerFactory) (client *HelmClient, err error) {
	client = new(HelmClient)

	client.settings = cli.New()
	client.settings.Debug = true
	client.log = log
	client.ctx = ctx
	if err := patchHelmDirectories(ctx, log, client); err != nil {
		return nil, err
	}
	client.actionConfig = new(action.Configuration)

	_log := &CustomLogger{Logger: log, ctx: ctx}
	if err := client.actionConfig.Init(client.settings.RESTClientGetter(), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), _log.HelmDebugf); err != nil {
		return nil, ksctlErrors.ErrFailedHelmClient.Wrap(
			log.NewError(ctx, "failed to init in-cluster helm client", "Reason", err),
		)
	}
	return client, nil
}
