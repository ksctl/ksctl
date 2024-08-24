package components

import (
	"strings"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/poller"
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
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "helmBaseChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmBaseChartOverridings = v
			}
		case "helmIstiodChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmIstiodChartOverridings = v
			}
		}
	}
	return
}

func setIsitoComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	helmBaseChartOverridings map[string]any,
	helmIstiodChartOverridings map[string]any,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("istio", "istio")
	if err != nil {
		return "", nil, nil, err
	}

	helmBaseChartOverridings = map[string]any{}
	helmIstiodChartOverridings = map[string]any{}

	_version, _helmBaseChartOverridings, _helmIstiodChartOverridings := getIstioComponentOverridings(p)

	version = getVersionIfItsNotNilAndLatest(_version, releases[0])

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
	return
}

func IstioStandardComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {

	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://istio-release.storage.googleapis.com/charts",
			RepoName: "istio",
			Charts: []metadata.ChartOptions{
				{
					Name:            "istio/base",
					Version:         version,
					ReleaseName:     "istio-base",
					Namespace:       "istio-system",
					CreateNamespace: true,
					Args:            helmBaseChartOverridings,
				},
				{
					Name:            "istio/istiod",
					Version:         version,
					ReleaseName:     "istiod",
					Namespace:       "istio-system",
					CreateNamespace: false,
					Args:            helmIstiodChartOverridings,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}, nil
}
