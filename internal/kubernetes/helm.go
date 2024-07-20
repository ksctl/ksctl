package kubernetes

import "github.com/ksctl/ksctl/internal/kubernetes/metadata"

func installHelm(client *K8sClusterClient, component *metadata.HelmHandler) error {

	repoName, repoUrl, charts := component.RepoName, component.RepoUrl, component.Charts

	if err := client.helmClient.RepoAdd(repoName, repoUrl); err != nil {
		return err
	}

	for _, chart := range charts {
		if err := client.helmClient.
			InstallChart(
				chart.ChartVer,
				chart.ChartName,
				chart.Namespace,
				chart.ReleaseName,
				chart.CreateNamespace,
				chart.Args,
			); err != nil {
			return err
		}
	}

	if err := client.helmClient.ListInstalledCharts(); err != nil {
		return err
	}
	return nil
}

func deleteHelm(client *K8sClusterClient, component *metadata.HelmHandler) error {

	charts := component.Charts

	for _, chart := range charts {
		if err := client.helmClient.
			UninstallChart(
				chart.Namespace,
				chart.ReleaseName,
			); err != nil {
			return err
		}
	}

	return nil
}
