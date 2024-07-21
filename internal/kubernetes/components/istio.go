package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getIstioComponentOverridings(p metadata.ComponentOverrides) (version *string, helmBaseChartOverridings map[string]interface{}, helmIstiodChartOverridings map[string]interface{}) {
	helmBaseChartOverridings = nil // By default, it is nil
	helmIstiodChartOverridings = nil

	if p == nil {
		return nil, nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			version = utilities.Ptr(v.(string))
		case "helmBaseChartOverridings":
			helmBaseChartOverridings = v.(map[string]interface{})
		case "helmIstiodChartOverridings":
			helmIstiodChartOverridings = v.(map[string]interface{})
		}
	}
	return
}

func IstioStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {

	var (
		version                    = "latest"
		helmBaseChartOverridings   = map[string]any{}
		helmIstiodChartOverridings = map[string]any{}
	)

	_version, _helmBaseChartOverridings, _helmIstiodChartOverridings := getIstioComponentOverridings(params)

	if _version != nil {
		version = *_version
	}

	if _helmBaseChartOverridings != nil {
		helmBaseChartOverridings = _helmBaseChartOverridings
	} else {
		helmBaseChartOverridings = map[string]any{
			"defaultRevision": "default",
		}
	}

	if _helmIstiodChartOverridings != nil {
		helmIstiodChartOverridings = _helmIstiodChartOverridings
	} else {
		helmIstiodChartOverridings = nil
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://istio-release.storage.googleapis.com/charts",
			RepoName: "istio",
			Charts: []metadata.HelmOptions{
				{
					ChartName:       "istio/base",
					ChartVer:        version,
					ReleaseName:     "istio-base",
					Namespace:       "istio-system",
					CreateNamespace: true,
					Args:            helmBaseChartOverridings,
				},
				{
					ChartName:       "istio/istiod",
					ChartVer:        version,
					ReleaseName:     "istiod",
					Namespace:       "istio-system",
					CreateNamespace: false,
					Args:            helmIstiodChartOverridings,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}
}
