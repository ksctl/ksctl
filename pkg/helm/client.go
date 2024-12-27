package helm

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type Client struct {
	ctx context.Context
	log logger.Logger

	actionConfig *action.Configuration
	settings     *cli.EnvSettings
}

type ChartOptions struct {
	Version         string
	Name            string
	ReleaseName     string
	Namespace       string
	CreateNamespace bool
	Args            map[string]interface{}
	ChartRef        string // only use it for oci:// based charts
}

type App struct {
	RepoUrl  string
	RepoName string
	Charts   []ChartOptions
}

func NewKubeconfigHelmClient(ctx context.Context, log logger.Logger, kubeconfig string) (client *Client, err error) {
	client = new(Client)

	client.settings = cli.New()
	client.settings.Debug = true
	client.log = log
	client.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "helm-client")
	if err := patchHelmDirectories(ctx, log, client); err != nil {
		return nil, err
	}

	client.actionConfig = new(action.Configuration)

	_log := &CustomLogger{Logger: log, ctx: ctx}

	if err := client.actionConfig.Init(NewRESTClientGetter(client.settings.Namespace(), kubeconfig), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), _log.HelmDebugf); err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			log.NewError(ctx, "failed to init kubeconfig based helm client", "Reason", err),
		)
	}
	return client, nil
}

func NewInClusterHelmClient(ctx context.Context, log logger.Logger) (client *Client, err error) {
	client = new(Client)

	client.settings = cli.New()
	client.settings.Debug = true
	client.log = log
	client.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "helm-client")
	if err := patchHelmDirectories(ctx, log, client); err != nil {
		return nil, err
	}
	client.actionConfig = new(action.Configuration)

	_log := &CustomLogger{Logger: log, ctx: ctx}
	if err := client.actionConfig.Init(client.settings.RESTClientGetter(), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), _log.HelmDebugf); err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			log.NewError(ctx, "failed to init in-cluster helm client", "Reason", err),
		)
	}
	return client, nil
}
