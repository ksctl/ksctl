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
	"fmt"
	"path/filepath"
	"time"

	"helm.sh/helm/v3/pkg/registry"

	"github.com/fatih/color"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func (c *Client) RepoAdd(repoName, repoUrl string) error {

	if len(repoUrl) == 0 || len(repoName) == 0 {
		c.log.Print(c.ctx, "Skip repoAdd due to repo Url being empty, will be trying out oci://")
		return nil
	} else {

		repoEntry := repo.Entry{
			Name: repoName,
			URL:  repoUrl,
		}

		r, err := repo.NewChartRepository(&repoEntry, getter.All(c.settings))
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedHelmClient,
				c.log.NewError(c.ctx, "constructs ChartRepository", "Reason", err),
			)
		}
		_, err = r.DownloadIndexFile()
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedHelmClient,
				c.log.NewError(c.ctx, "failed to download the chart", "Reason", err),
			)
		}

		existingRepositoryFile, err := repo.LoadFile(c.settings.RepositoryConfig)
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedHelmClient,
				c.log.NewError(c.ctx, "failed to load the chart", "Reason", err),
			)
		}

		if !existingRepositoryFile.Has(repoEntry.Name) {
			existingRepositoryFile.Add(&repoEntry)

			err = existingRepositoryFile.WriteFile(c.settings.RepositoryConfig, 0644)
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedHelmClient,
					c.log.NewError(c.ctx, "failed to write the chart", "Reason", err),
				)
			}
		}
	}

	return nil
}

func (c *Client) UninstallChart(namespace, releaseName string) error {

	clientUninstall := action.NewUninstall(c.actionConfig)

	clientUninstall.Wait = true
	clientUninstall.Timeout = 5 * time.Minute

	_, err := clientUninstall.Run(releaseName)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			c.log.NewError(c.ctx, "failed uninstall the chart", "Reason", err),
		)
	}
	return nil
}

// TODO: need to have a upgrade and rollbacks
func (c *Client) UpdateChart() error {
	return nil
}

func (c *Client) RollbackChart() error {
	return nil
}

func (c *Client) InstallChart(
	chartRef,
	chartVer,
	chartName,
	namespace,
	releaseName string,
	createNamespace bool,
	arguments map[string]interface{}) error {

	if len(chartRef) != 0 && registry.IsOCI(chartRef) {
		if errOciPull := c.runPull(chartRef, chartVer); errOciPull != nil {
			return errOciPull
		}
	}

	clientInstall := action.NewInstall(c.actionConfig)

	// NOTE: Patch for the helm latest releases
	if chartVer == "latest" {
		chartVer = ""
	}

	clientInstall.ChartPathOptions.Version = chartVer
	clientInstall.ReleaseName = releaseName
	clientInstall.Namespace = namespace // FIXME: this is not working
	c.settings.SetNamespace(namespace)
	//	if c.settings.Namespace() != clientInstall.Namespace {
	//		panic(fmt.Sprintf("Namespace mismatch: %s != %s", c.settings.Namespace(), clientInstall.Namespace))
	//	}

	clientInstall.CreateNamespace = createNamespace

	clientInstall.Wait = true
	clientInstall.Timeout = 5 * time.Minute

	//////
	// registryClient, err := newRegistryClientTLS(
	// 	c.settings,
	// 	c.log.ExternalLogHandlerf,
	// 	clientInstall.CertFile,
	// 	clientInstall.KeyFile,
	// 	clientInstall.CaFile,
	// 	clientInstall.InsecureSkipTLSverify,
	// 	clientInstall.PlainHTTP)
	// if err != nil {
	// 	return fmt.Errorf("failed to created registry client: %w", err)
	// }
	// clientInstall.SetRegistryClient(registryClient)
	/////

	chartPath, err := func() (string, error) {
		if len(chartRef) != 0 && registry.IsOCI(chartRef) && c.ociPullDestDir != nil {
			return filepath.Join(
				*c.ociPullDestDir,
				chartName,
			), nil
		}
		return clientInstall.ChartPathOptions.LocateChart(chartName, c.settings)
	}()
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			c.log.NewError(c.ctx, "failed to locate chart", "Reason", err),
		)
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			c.log.NewError(c.ctx, "failed to load a chart", "Reason", err),
		)
	}

	_, err = clientInstall.Run(chartRequested, arguments)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			c.log.NewError(c.ctx, "failed to install a chart", "Reason", err),
		)
	}
	return nil
}

func (c *Client) ListInstalledCharts() error {

	client := action.NewList(c.actionConfig)
	// Only list deployed
	client.Deployed = true
	results, err := client.Run()
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedHelmClient,
			c.log.NewError(c.ctx, "failed to list installed charts", "Reason", err),
		)
	}

	c.log.Print(c.ctx, "Lists installed Charts")
	for _, rel := range results {
		c.log.Box(c.ctx,
			rel.Chart.Name(),
			fmt.Sprintf(
				"Namespace\n----\n%s\n\nVersion\n----\n%s\n\nDescription\n----\n%s\n\nStatus\n----\n%s\n\nNotes\n----\n%s",
				color.MagentaString(rel.Namespace),
				rel.Chart.AppVersion(),
				rel.Info.Description,
				rel.Info.Status,
				rel.Info.Notes))
	}
	return nil
}

////////////////// Using kubeconfig as content not as file in os //////////////////
// Reference: https://github.com/helm/helm/issues/6910#issuecomment-601277026

type SimpleRESTClientGetter struct {
	KubeConfig string
}

func NewRESTClientGetter(kubeConfig string) *SimpleRESTClientGetter {
	return &SimpleRESTClientGetter{
		KubeConfig: kubeConfig,
	}
}

func (c *SimpleRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(c.KubeConfig))
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *SimpleRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// The more groups you have, the more discovery requests you need to make.
	// given 25 groups (our groups + a few custom conf) with one-ish version each, discovery needs to make 50 requests
	// double it just so we don't end up here again for a while.  This config is only used for discovery.
	config.Burst = 100

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)
	return memory.NewMemCacheClient(discoveryClient), nil
}

func (c *SimpleRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, nil)
	return expander, nil
}

func (c *SimpleRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	// overrides.Context.Namespace = c.Namespace // COMMENTED OUT: It should not enforce any specific namespace from the worker's context

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}

///////////////////////////////////////////////////////////////////////////////////
