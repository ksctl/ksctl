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
