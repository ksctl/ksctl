package components

import (
	"strings"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/poller"
)

func getCiliumComponentOverridings(p metadata.ComponentOverrides) (version *string, ciliumChartOverridings map[string]any) {
	ciliumChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "ciliumChartOverridings":
			if v, ok := v.(map[string]any); ok {
				ciliumChartOverridings = v
			}
		}
	}
	return
}

func setCiliumComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	ciliumChartOverridings map[string]any,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("cilium", "cilium")
	if err != nil {
		return "", nil, err
	}

	ciliumChartOverridings = map[string]any{}

	_version, _ciliumChartOverridings := getCiliumComponentOverridings(p)

	version = getVersionIfItsNotNilAndLatest(_version, releases[0])

	if _ciliumChartOverridings != nil {
		ciliumChartOverridings = _ciliumChartOverridings
	} else {
		ciliumChartOverridings = nil
	}
	return
}

func CiliumStandardComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoName: "cilium",
			RepoUrl:  "https://helm.cilium.io/",
			Charts: []metadata.ChartOptions{
				{
					Name:            "cilium/cilium",
					Version:         version,
					ReleaseName:     "cilium",
					Namespace:       "kube-system",
					CreateNamespace: false,
					Args:            ciliumChartOverridings,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}, nil
}
