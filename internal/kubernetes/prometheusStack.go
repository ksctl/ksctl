package kubernetes

func kubePrometheusStandardMonitoring(params applicationParams) ApplicationStack {
	return ApplicationStack{
		components: []StackComponent{
			{
				helm: &HelmHandler{
					repoUrl:  "https://prometheus-community.github.io/helm-charts",
					repoName: "prometheus-community",
					charts: []HelmOptions{
						{
							chartName:       "prometheus-community/kube-prometheus-stack",
							chartVer:        params.version,
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
