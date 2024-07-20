package kubernetes

func installHelm(client *K8sClusterClient, component *HelmHandler) error {

	repoName, repoUrl, charts := component.repoName, component.repoUrl, component.charts

	if err := client.helmClient.RepoAdd(repoName, repoUrl); err != nil {
		return err
	}

	for _, chart := range charts {
		if err := client.helmClient.
			InstallChart(chart.chartVer, chart.chartName, chart.namespace, chart.releaseName, chart.createNamespace, chart.args); err != nil {
			return err
		}
	}

	if err := client.helmClient.ListInstalledCharts(); err != nil {
		return err
	}
	return nil
}

func deleteHelm(client *K8sClusterClient, component *HelmHandler) error {

	charts := component.charts

	for _, chart := range charts {
		if err := client.helmClient.
			UninstallChart(chart.namespace, chart.releaseName); err != nil {
			return err
		}
	}

	return nil
}
