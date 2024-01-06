package universal

import (
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
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

	log.Print("Lists installed Charts")
	for _, rel := range results {
		log.Print(rel.Chart.Name(), rel.Namespace, rel.Info.Description)
	}
	return nil
}

////////////////// Using kubeconfig as content not as file in os //////////////////
// Reference: https://github.com/helm/helm/issues/6910#issuecomment-601277026

type SimpleRESTClientGetter struct {
	Namespace  string
	KubeConfig string
}

func NewRESTClientGetter(namespace, kubeConfig string) *SimpleRESTClientGetter {
	return &SimpleRESTClientGetter{
		Namespace:  namespace,
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
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (c *SimpleRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	overrides.Context.Namespace = c.Namespace

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}

///////////////////////////////////////////////////////////////////////////////////

func (client *HelmClient) InitClient(kubeconfig string) error {

	client.settings = cli.New()
	client.settings.Debug = true

	client.actionConfig = new(action.Configuration)
	if err := client.actionConfig.Init(NewRESTClientGetter(client.settings.Namespace(), kubeconfig), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Debug); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func installHelm(client *Kubernetes, appStruct Application) error {

	repoName, repoUrl, charts := appStruct.Name, appStruct.Url, appStruct.HelmConfig

	if err := client.helmClient.RepoAdd(repoName, repoUrl); err != nil {
		return log.NewError(err.Error())
	}

	for _, chart := range charts {
		if err := client.helmClient.InstallChart(chart.chartVer, chart.chartName, chart.namespace, chart.releaseName, chart.createns, chart.args); err != nil {
			return log.NewError(err.Error())
		}
	}

	if err := client.helmClient.ListInstalledCharts(); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
