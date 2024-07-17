package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func IstioStandardComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://istio-release.storage.googleapis.com/charts",
			RepoName: "istio",
			Charts: []metadata.HelmOptions{
				{
					ChartName:       "istio/base",
					ChartVer:        params.Version,
					ReleaseName:     "istio-base",
					Namespace:       "istio-system",
					CreateNamespace: true,
					Args: map[string]interface{}{
						"defaultRevision": "default",
					},
				},
				{
					ChartName:       "istio/istiod",
					ChartVer:        params.Version,
					ReleaseName:     "istiod",
					Namespace:       "istio-system",
					CreateNamespace: false,
					Args:            nil,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}
}
