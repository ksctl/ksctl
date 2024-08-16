package helmclient

import (
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

// var helmDriver string = os.Getenv("HELM_DRIVER")
//
//	func initActionConfig(settings *cli.EnvSettings, logger *log.Logger) (*action.Configuration, error) {
//		return initActionConfigList(settings, logger, false)
//	}
//
// func initActionConfigList(settings *cli.EnvSettings, logger *log.Logger, allNamespaces bool) (*action.Configuration, error) {
//
//		actionConfig := new(action.Configuration)
//
//		namespace := func() string {
//			// For list action, you can pass an empty string instead of settings.Namespace() to list
//			// all namespaces
//			if allNamespaces {
//				return ""
//			}
//			return settings.Namespace()
//		}()
//
//		if err := actionConfig.Init(
//			settings.RESTClientGetter(),
//			namespace,
//			helmDriver,
//			logger.Printf); err != nil {
//			return nil, err
//		}
//
//		return actionConfig, nil
//	}

func newRegistryClient(settings *cli.EnvSettings, plainHTTP bool) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	}
	if plainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

//
//func newRegistryClientTLS(settings *cli.EnvSettings, logger *log.Logger, certFile, keyFile, caFile string, insecureSkipTLSverify, plainHTTP bool) (*registry.Client, error) {
//	if certFile != "" && keyFile != "" || caFile != "" || insecureSkipTLSverify {
//		registryClient, err := registry.NewRegistryClientWithTLS(
//			logger.Writer(),
//			certFile,
//			keyFile,
//			caFile,
//			insecureSkipTLSverify,
//			settings.RegistryConfig,
//			settings.Debug)
//
//		if err != nil {
//			return nil, err
//		}
//		return registryClient, nil
//	}
//	registryClient, err := newRegistryClient(settings, plainHTTP)
//	if err != nil {
//		return nil, err
//	}
//	return registryClient, nil
//}
//
//func runInstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, releaseName string, chartRef string, chartVersion string, releaseValues map[string]interface{}) error {
//
//	actionConfig, err := initActionConfig(settings, logger)
//	if err != nil {
//		return fmt.Errorf("failed to init action config: %w", err)
//	}
//
//	installClient := action.NewInstall(actionConfig)
//
//	installClient.DryRunOption = "none"
//	installClient.ReleaseName = releaseName
//	installClient.Namespace = settings.Namespace()
//	installClient.Version = chartVersion
//
//	registryClient, err := newRegistryClientTLS(
//		settings,
//		logger,
//		installClient.CertFile,
//		installClient.KeyFile,
//		installClient.CaFile,
//		installClient.InsecureSkipTLSverify,
//		installClient.PlainHTTP)
//	if err != nil {
//		return fmt.Errorf("failed to created registry client: %w", err)
//	}
//	installClient.SetRegistryClient(registryClient)
//
//	chartPath, err := installClient.ChartPathOptions.LocateChart(chartRef, settings)
//	if err != nil {
//		return err
//	}
//
//	// providers := getter.All(settings)
//
//	chart, err := loader.Load(chartPath)
//	if err != nil {
//		return err
//	}
//
//	// // Check chart dependencies to make sure all are present in /charts
//	// if chartDependencies := chart.Metadata.Dependencies; chartDependencies != nil {
//	// 	if err := action.CheckDependencies(chart, chartDependencies); err != nil {
//	// 		err = fmt.Errorf("failed to check chart dependencies: %w", err)
//	// 		if !installClient.DependencyUpdate {
//	// 			return err
//	// 		}
//	//
//	// 		manager := &downloader.Manager{
//	// 			Out:              logger.Writer(),
//	// 			ChartPath:        chartPath,
//	// 			Keyring:          installClient.ChartPathOptions.Keyring,
//	// 			SkipUpdate:       false,
//	// 			Getters:          providers,
//	// 			RepositoryConfig: settings.RepositoryConfig,
//	// 			RepositoryCache:  settings.RepositoryCache,
//	// 			Debug:            settings.Debug,
//	// 			RegistryClient:   installClient.GetRegistryClient(),
//	// 		}
//	// 		if err := manager.Update(); err != nil {
//	// 			return err
//	// 		}
//	// 		// Reload the chart with the updated Chart.lock file.
//	// 		if chart, err = loader.Load(chartPath); err != nil {
//	// 			return fmt.Errorf("failed to reload chart after repo update: %w", err)
//	// 		}
//	// 	}
//	// }
//
//	release, err := installClient.RunWithContext(ctx, chart, releaseValues)
//	if err != nil {
//		return fmt.Errorf("failed to run install: %w", err)
//	}
//
//	logger.Printf("release created:\n%+v", *release)
//
//	return nil
//}

func (c *HelmClient) runPull(chartRef, chartVersion string) error {

	if chartVersion == "latest" {
		return fmt.Errorf("latest version is not supported for oci:// pull")
	}

	registryClient, err := newRegistryClient(c.settings, false)
	if err != nil {
		return fmt.Errorf("failed to created registry client: %w", err)
	}

	c.actionConfig.RegistryClient = registryClient

	pullClient := action.NewPullWithOpts(
		action.WithConfig(c.actionConfig))
	pullClient.DestDir = "./"
	pullClient.Settings = c.settings
	pullClient.Version = chartVersion

	result, err := pullClient.Run(chartRef)
	if err != nil {
		return fmt.Errorf("failed to pull chart: %w", err)
	}
	c.log.Success(c.ctx, "chart pulled", "result", result)

	return nil
}
