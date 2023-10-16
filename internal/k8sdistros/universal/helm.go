package universal

import (
	"fmt"
	"log"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type HelmClient struct {
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
}

type WorkLoad struct {
	chartVer    string
	chartName   string
	releaseName string
	namespace   string
	createns    bool
	args        map[string]interface{}
}

func (c *HelmClient) RepoAdd(repoName, repoUrl string) error {

	repoEntry := repo.Entry{
		Name: repoName,
		URL:  repoUrl,
	}

	r, err := repo.NewChartRepository(&repoEntry, getter.All(c.settings))
	if err != nil {
		return err
	}
	_, err = r.DownloadIndexFile()
	if err != nil {
		return err
	}
	// Read the existing repository file
	existingRepositoryFile, err := repo.LoadFile(c.settings.RepositoryConfig)
	if err != nil {
		return err
	}

	if !existingRepositoryFile.Has(repoEntry.Name) {
		existingRepositoryFile.Add(&repoEntry)

		err = existingRepositoryFile.WriteFile(c.settings.RepositoryConfig, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *HelmClient) InstallChart(chartVer, chartName, namespace, releaseName string, createNamespace bool, arguments map[string]interface{}) error {

	clientInstall := action.NewInstall(c.actionConfig)
	clientInstall.ChartPathOptions.Version = chartVer
	clientInstall.ReleaseName = releaseName
	clientInstall.Namespace = namespace

	clientInstall.CreateNamespace = createNamespace

	clientInstall.Wait = true
	clientInstall.Timeout = 5 * time.Minute

	chartPath, err := clientInstall.ChartPathOptions.LocateChart(chartName, c.settings)
	if err != nil {
		return err
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	_, err = clientInstall.Run(chartRequested, arguments)
	if err != nil {
		return err
	}
	return nil
}

func (c *HelmClient) ListInstalledCharts() error {

	client := action.NewList(c.actionConfig)
	// Only list deployed
	client.Deployed = true
	results, err := client.Run()
	if err != nil {
		return err
	}

	fmt.Println("Lists installed Charts")
	for _, rel := range results {
		fmt.Println(rel.Chart.Name(), rel.Namespace, rel.Info.Description)
	}
	fmt.Println()
	return nil
}

func (client *HelmClient) InitClient(kubeconfig string) error {

	client.settings = cli.New()
	client.settings.Debug = true
	client.settings.KubeConfig = kubeconfig

	client.actionConfig = new(action.Configuration)
	if err := client.actionConfig.Init(client.settings.RESTClientGetter(), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return err
	}
	return nil
}

func installHelm(client *Kubernetes, appStruct Application) error {

	repoName, repoUrl, charts := appStruct.Name, appStruct.Url, appStruct.HelmConfig

	if err := client.helmClient.RepoAdd(repoName, repoUrl); err != nil {
		return err
	}

	for _, chart := range charts {
		if err := client.helmClient.InstallChart(chart.chartVer, chart.chartName, chart.namespace, chart.releaseName, chart.createns, chart.args); err != nil {
			return err
		}
	}

	if err := client.helmClient.ListInstalledCharts(); err != nil {
		return err
	}
	return nil
}
