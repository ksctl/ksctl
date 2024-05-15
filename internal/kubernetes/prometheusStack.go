package kubernetes

func prometheusStackData(ver string) Application {

	return Application{
		Name:        "prometheus-community",
		Url:         "https://prometheus-community.github.io/helm-charts",
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			HelmOptions{
				chartName:       "prometheus-community/kube-prometheus-stack",
				chartVer:        ver,
				releaseName:     "kube-prometheus-stack",
				namespace:       "monitoring",
				createNamespace: true,
			},
		},
	}
}
