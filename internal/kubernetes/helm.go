package kubernetes

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	localStore "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"

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

type HelmOptions struct {
	chartVer        string
	chartName       string
	releaseName     string
	namespace       string
	createNamespace bool
	args            map[string]interface{}
}

func (c *HelmClient) RepoAdd(repoName, repoUrl string) error {

	repoEntry := repo.Entry{
		Name: repoName,
		URL:  repoUrl,
	}

	r, err := repo.NewChartRepository(&repoEntry, getter.All(c.settings))
	if err != nil {
		return log.NewError(kubernetesCtx, "constructs ChartRepository", "Reason", err)
	}
	_, err = r.DownloadIndexFile()
	if err != nil {
		return log.NewError(kubernetesCtx, "failed to download the chart", "Reason", err)
	}

	existingRepositoryFile, err := repo.LoadFile(c.settings.RepositoryConfig)
	if err != nil {
		return log.NewError(kubernetesCtx, "failed to load the chart", "Reason", err)
	}

	if !existingRepositoryFile.Has(repoEntry.Name) {
		existingRepositoryFile.Add(&repoEntry)

		err = existingRepositoryFile.WriteFile(c.settings.RepositoryConfig, 0644)
		if err != nil {
			return log.NewError(kubernetesCtx, "failed to write the chart", "Reason", err)
		}
	}

	return nil
}

func (c *HelmClient) UninstallChart(namespace, releaseName string) error {

	clientUninstall := action.NewUninstall(c.actionConfig)

	clientUninstall.Wait = true
	clientUninstall.Timeout = 5 * time.Minute

	_, err := clientUninstall.Run(releaseName)
	if err != nil {
		return log.NewError(kubernetesCtx, "failed uninstall the chart", "Reason", err)
	}
	return nil
}

func (c *HelmClient) InstallChart(chartVer, chartName, namespace, releaseName string, createNamespace bool, arguments map[string]interface{}) error {

	clientInstall := action.NewInstall(c.actionConfig)

	// NOTE: Patch for the helm latest releases
	if chartVer == "latest" {
		chartVer = ""
	}

	clientInstall.ChartPathOptions.Version = chartVer
	clientInstall.ReleaseName = releaseName
	clientInstall.Namespace = namespace

	clientInstall.CreateNamespace = createNamespace

	clientInstall.Wait = true
	clientInstall.Timeout = 5 * time.Minute

	chartPath, err := clientInstall.ChartPathOptions.
		LocateChart(chartName, c.settings)
	if err != nil {
		return log.NewError(kubernetesCtx, "failed to locate chart", "Reason", err)
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return log.NewError(kubernetesCtx, "failed to load a chart", "Reason", err)
	}

	_, err = clientInstall.Run(chartRequested, arguments)
	if err != nil {
		return log.NewError(kubernetesCtx, "failed to install a chart", "Reason", err)
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

	log.Print(kubernetesCtx, "Lists installed Charts")
	for _, rel := range results {
		log.Box(kubernetesCtx,
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
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, nil)
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

type CustomLogger struct {
	Logger types.LoggerFactory
}

func (l *CustomLogger) HelmDebugf(format string, v ...interface{}) {
	l.Logger.ExternalLogHandlerf(kubernetesCtx, consts.LOG_INFO, format+"\n", v...)
}

func patchHelmDirectories(client *HelmClient) error {
	usr, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	store := localStore.NewClient(kubernetesCtx, log)

	pathConfig := []string{usr, ".config", "helm"}
	_, okConfig := store.PresentDirectory(pathConfig)
	if !okConfig {
		if _err := store.CreateDirectory(pathConfig); _err != nil {
			return _err
		}
	}

	// pathCache := []string{usr, ".cache", "helm", "repository"}
	// cachePath, okCache := store.PresentDirectory(pathCache)
	// if !okCache {
	// 	if _err := store.CreateDirectory(pathCache); _err != nil {
	// 		return _err
	// 	}
	// }

	pathRegistry := []string{usr, ".config", "helm", "registry"}
	_, okReg := store.PresentDirectory(pathRegistry)
	if !okReg {
		if _err := store.CreateDirectory(pathRegistry); _err != nil {
			return _err
		}
	}

	pathConfig = append(pathConfig, "repositories.yaml")
	configPath, _err := store.CreateFileIfNotPresent(pathConfig)
	if _err != nil {
		return _err
	}

	pathRegistry = append(pathRegistry, "config.json")
	registryPath, _err := store.CreateFileIfNotPresent(pathRegistry)
	if _err != nil {
		return _err
	}

	if err := store.Kill(); err != nil {
		return err
	}

	if _err := os.Setenv("HELM_DRIVER", "secrets"); _err != nil {
		return _err
	}
	client.settings.RepositoryConfig = configPath
	// client.settings.RepositoryCache = cachePath
	client.settings.RegistryConfig = registryPath
	log.Print(kubernetesCtx, "Updated the Helm configuration settings")

	return nil
}

func (client *HelmClient) NewKubeconfigHelmClient(kubeconfig string) error {

	client.settings = cli.New()
	client.settings.Debug = true
	if err := patchHelmDirectories(client); err != nil {
		return err
	}

	client.actionConfig = new(action.Configuration)

	_log := &CustomLogger{Logger: log}

	if err := client.actionConfig.Init(NewRESTClientGetter(client.settings.Namespace(), kubeconfig), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), _log.HelmDebugf); err != nil {
		return log.NewError(kubernetesCtx, "failed to init kubeconfig based helm client", "Reason", err)
	}
	return nil
}

func (client *HelmClient) NewInClusterHelmClient() error {

	client.settings = cli.New()
	client.settings.Debug = true
	if err := patchHelmDirectories(client); err != nil {
		return err
	}
	client.actionConfig = new(action.Configuration)

	_log := &CustomLogger{Logger: log}
	if err := client.actionConfig.Init(client.settings.RESTClientGetter(), client.settings.Namespace(), os.Getenv("HELM_DRIVER"), _log.HelmDebugf); err != nil {
		return log.NewError(kubernetesCtx, "failed to init in-cluster helm client", "Reason", err)
	}
	return nil
}

func installHelm(client *Kubernetes, appStruct Application) error {

	repoName, repoUrl, charts := appStruct.Name, appStruct.Url, appStruct.HelmConfig

	if err := client.helmClient.
		RepoAdd(repoName, repoUrl); err != nil {
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

func deleteHelm(client *Kubernetes, appStruct Application) error {

	charts := appStruct.HelmConfig

	for _, chart := range charts {
		if err := client.helmClient.
			UninstallChart(chart.namespace, chart.releaseName); err != nil {
			return err
		}
	}

	return nil
}
