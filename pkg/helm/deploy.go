package helm


func (k *Client) HelmDeploy(component *App) error {

	repoName, repoUrl, charts := component.RepoName, component.RepoUrl, component.Charts

	if err := k.RepoAdd(repoName, repoUrl); err != nil {
		return err
	}

	for _, chart := range charts {
		if err := k.InstallChart(
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
