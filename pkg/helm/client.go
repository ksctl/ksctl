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
	"path/filepath"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Client struct {
	ctx context.Context
	log logger.Logger

	actionConfig *action.Configuration
	settings     *cli.EnvSettings

	ociPullDestDir *string
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

type options struct {
	runAsDebug bool
	kubeconfig *string
	ociPullDir *string
}
type Option func(options *options) error

func WithDebug() Option {
	return func(o *options) error {
		o.runAsDebug = true
		return nil
	}
}

func WithKubeconfig(kubeconfig string) Option {
	return func(o *options) error {
		o.kubeconfig = &kubeconfig
		return nil
	}
}

func WithOCIChartPullDestDir(dir string) Option {
	return func(o *options) error {
		f, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		f = "./" + f
		o.ociPullDir = &f
		return nil
	}
}

func NewClient(ctx context.Context, log logger.Logger, opts ...Option) (client *Client, err error) {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	client = new(Client)
	client.actionConfig = new(action.Configuration)
	client.settings = cli.New()
	client.log = log
	client.ctx = context.WithValue(
		ctx,
		consts.KsctlModuleNameKey,
		"helm-client",
	)

	client.settings.Debug = options.runAsDebug
	if options.ociPullDir != nil {
		client.ociPullDestDir = options.ociPullDir
	}

	_log := &CustomLogger{Logger: log, ctx: ctx}

	if err := patchHelmDirectories(client.ctx, client.log, client); err != nil {
		return nil, err
	}

	var getter genericclioptions.RESTClientGetter
	var namespaceToInit string

	if options.kubeconfig != nil {
		getter = NewRESTClientGetter(*options.kubeconfig)
		namespaceToInit = "" // When using an external kubeconfig, do not impose the current pod's namespace
	} else {
		getter = client.settings.RESTClientGetter()
		namespaceToInit = client.settings.Namespace()
	}

	if err := client.actionConfig.Init(
		getter,
		namespaceToInit, // Use the determined namespace
		os.Getenv("HELM_DRIVER"),
		_log.HelmDebugf,
	); err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			log.NewError(ctx, "failed to init helm client", "Reason", err),
		)
	}

	return client, nil
}
