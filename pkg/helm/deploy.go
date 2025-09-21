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

func (k *Client) HelmDeploy(component *App) error {

	repoName, repoUrl, charts := component.RepoName, component.RepoUrl, component.Charts

	if err := k.RepoAdd(repoName, repoUrl); err != nil {
		return err
	}

	for _, chart := range charts {
		if err := k.UpgradeOrInstallChart(
			chart.ChartRef,
			chart.Version,
			chart.Name,
			chart.Namespace,
			chart.ReleaseName,
			chart.CreateNamespace,
			chart.Args,
		); err != nil {
			return err
		}
	}

	if err := k.ListInstalledCharts(); err != nil {
		return err
	}
	return nil
}

func (k *Client) HelmUninstall(component *App) error {

	charts := component.Charts

	for _, chart := range charts {
		if err := k.UninstallChart(
			chart.Namespace,
			chart.ReleaseName,
		); err != nil {
			return err
		}
	}

	return nil
}
