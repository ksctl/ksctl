package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"strings"
)

func getKubePrometheusComponentOverridings(p metadata.ComponentOverrides) (version *string, helmKubePromChartOverridings map[string]interface{}) {
	helmKubePromChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "helmKubePromChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmKubePromChartOverridings = v
			}
		}
	}
	return
}

func setKubePrometheusComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	helmKubePromChartOverridings map[string]any,
) {
	version = "latest"
	helmKubePromChartOverridings = map[string]any{}

	_version, _helmKubePromChartOverridings := getKubePrometheusComponentOverridings(p)
	if _version != nil {
		version = *_version
	}

	if _helmKubePromChartOverridings != nil {
		helmKubePromChartOverridings = _helmKubePromChartOverridings
	} else {
		helmKubePromChartOverridings = nil
	}

	return
}

func KubePrometheusStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {

	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(params)

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://prometheus-community.github.io/helm-charts",
			RepoName: "prometheus-community",
			Charts: []metadata.ChartOptions{
				{
					Name:            "prometheus-community/kube-prometheus-stack",
					Version:         version,
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
