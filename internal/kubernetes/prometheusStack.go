package kubernetes

func prometheusStackData(ver string) Application {

	return Application{
		Name:        "prometheus-community",
		Url:         "https://prometheus-community.github.io/helm-charts",
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			{
				chartName:       "prometheus-community/kube-prometheus-stack",
				chartVer:        ver,
				releaseName:     "kube-prometheus-stack",
				namespace:       "monitoring",
				createNamespace: true,
			},
		},
	}
}

func kubePrometheusStandardMonitoring(ver string) ApplicationStack {
	return ApplicationStack{
		components: []StackComponent{
			{
				helm: &HelmHandler{
					repoUrl:  "https://prometheus-community.github.io/helm-charts",
					repoName: "prometheus-community",
					charts: []HelmOptions{
						{
							chartName:       "prometheus-community/kube-prometheus-stack",
							chartVer:        ver,
							releaseName:     "kube-prometheus-stack",
							namespace:       "monitoring",
							createNamespace: true,
						},
					},
				},
				handlerType: ComponentTypeHelm,
			},
		},
		Maintainer:  "github:dipankardas011",
		StackNameID: KubePrometheusStandardStackID,
	}
}
