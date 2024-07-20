package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getKubePrometheusComponentOverridings(p metadata.ComponentOverriding) (version *string, helmKubePromChartOverridings map[string]interface{}) {
	helmKubePromChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			version = utilities.Ptr(v.(string))
		case "helmKubePromChartOverridings":
			helmKubePromChartOverridings = v.(map[string]interface{})
		}
	}
	return
}

func KubePrometheusStandardComponent(params metadata.ComponentOverriding) metadata.StackComponent {
	var (
		version                      = "latest"
		helmKubePromChartOverridings = map[string]any{}
	)

	_version, _helmKubePromChartOverridings := getKubePrometheusComponentOverridings(params)
	if _version != nil {
		version = *_version
	}

	if _helmKubePromChartOverridings != nil {
		helmKubePromChartOverridings = _helmKubePromChartOverridings
	} else {
		helmKubePromChartOverridings = nil
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://prometheus-community.github.io/helm-charts",
			RepoName: "prometheus-community",
			Charts: []metadata.HelmOptions{
				{
					ChartName:       "prometheus-community/kube-prometheus-stack",
					ChartVer:        version,
					ReleaseName:     "kube-prometheus-stack",
					Namespace:       "monitoring",
					CreateNamespace: true,
					Args:            helmKubePromChartOverridings,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}
}
