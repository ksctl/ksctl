package kubernetes

type helmClientMock struct{}

func (h helmClientMock) InstallChart(chartRef string, chartVer string, chartName string, namespace string, releaseName string, createNamespace bool, arguments map[string]interface{}) error {
	return nil
}

func (h helmClientMock) ListInstalledCharts() error {
	return nil
}

func (h helmClientMock) RepoAdd(repoName string, repoUrl string) error {
	return nil
}

func (h helmClientMock) UninstallChart(namespace string, releaseName string) error {
	return nil
}
