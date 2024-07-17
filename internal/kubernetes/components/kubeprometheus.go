package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KubePrometheusStandardComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://prometheus-community.github.io/helm-charts",
			RepoName: "prometheus-community",
			Charts: []metadata.HelmOptions{
				{
					ChartName:       "prometheus-community/kube-prometheus-stack",
					ChartVer:        params.Version,
					ReleaseName:     "kube-prometheus-stack",
					Namespace:       "monitoring",
					CreateNamespace: true,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}
}
